// Copyright 2022-2024 FLUIDOS Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkmanager

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkv1alpha1 "github.com/fluidos-project/node/apis/network/v1alpha1"
	nodecorev1alpha1 "github.com/fluidos-project/node/apis/nodecore/v1alpha1"
	"github.com/fluidos-project/node/pkg/utils/flags"
	"github.com/fluidos-project/node/pkg/utils/getters"
	"github.com/fluidos-project/node/pkg/utils/namings"
	"github.com/fluidos-project/node/pkg/utils/resourceforge"
)

// clusterRole
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=brokers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.fluidos.eu,resources=brokers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// BrokerClient keeps all the necessary class data.
type BrokerClient struct {
	ID   *nodecorev1alpha1.NodeIdentity
	ctx  context.Context
	canc context.CancelFunc

	subFlag    bool
	pubFlag    bool
	brokerName string
	serverAddr string
	clientCert *corev1.Secret
	rootCert   *corev1.Secret

	brokerConn *brokerConnection
}

// BrokerConnection keeps all the broker connection data.
type brokerConnection struct {
	amqpConn     *amqp.Connection
	amqpChan     *amqp.Channel
	exchangeName string
	routingKey   string
	queueName    string
	inboundMsgs  <-chan amqp.Delivery
	outboundMsg  []byte
	confirms     chan amqp.Confirmation
}

// SetupBrokerClient sets the Broker Client from NM reconcile.
func (bc *BrokerClient) SetupBrokerClient(cl client.Client, broker *networkv1alpha1.Broker) error {
	klog.Info("Setting up Broker Client routines")

	bc.ctx, bc.canc = context.WithCancel(context.Background())
	ctx := bc.ctx
	var err error

	bc.ID = getters.GetNodeIdentity(ctx, cl)
	if bc.ID == nil {
		return fmt.Errorf("failed to get Node Identity")
	}

	// Server address and broker name.
	bc.brokerName = broker.Spec.Name
	bc.serverAddr = broker.Spec.Address

	bc.brokerConn = &brokerConnection{}
	bc.brokerConn.exchangeName = "DefaultPeerRequest"

	bc.brokerConn.outboundMsg, err = json.Marshal(bc.ID)
	if err != nil {
		return err
	}

	switch role := broker.Spec.Role; role {
	case "publisher":
		bc.pubFlag = true
		bc.subFlag = false
		klog.Infof("brokerClient %s set as publisher only", bc.brokerName)
	case "subscriber":
		bc.pubFlag = false
		bc.subFlag = true
		klog.Infof("brokerClient %s set as subscriber only", bc.brokerName)
	default:
		bc.pubFlag = true
		bc.subFlag = true
		klog.Infof("brokerClient %s set as publisher and subscriber", bc.brokerName)
	}

	// Certificates.
	bc.clientCert = &corev1.Secret{}
	bc.rootCert = &corev1.Secret{}

	klog.Infof("Root Secret Name: %s\n", broker.Spec.CaCert.Name)
	klog.Infof("Client Secret Name: %s\n", broker.Spec.ClCert.Name)
	secretNamespace := "fluidos"

	err = bc.extractSecret(cl, broker.Spec.ClCert.Name, secretNamespace, bc.clientCert)
	if err != nil {
		return err
	}
	err = bc.extractSecret(cl, broker.Spec.CaCert.Name, secretNamespace, bc.rootCert)
	if err != nil {
		return err
	}

	// Extract certs and key.
	clientCert, ok := bc.clientCert.Data["tls.crt"]
	if !ok {
		klog.Error("missing certificate: 'tls.crt' not found in clCert Data")
	}

	clientKey, ok := bc.clientCert.Data["tls.key"]
	if !ok {
		klog.Error("missing key: 'tls.key' not found in clCert Data")
	}

	caCertData, ok := bc.rootCert.Data["CA_cert.pem"]
	if !ok {
		klog.Error("missing certificate: 'tls.crt' not found in CACert Data")
	}

	// Load client cert and privKey.
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		klog.Error("error X509KeyPair: %v", err)
		return err
	}

	// Load root cert.
	caCertPool := x509.NewCertPool()
	ok = caCertPool.AppendCertsFromPEM(caCertData)
	if !ok {
		klog.Error("AppendCertsFromPEM error: %v", ok)
	}

	// Routing key for topic.
	bc.brokerConn.routingKey, err = extractCNfromCert(&clientCert)
	if err != nil {
		klog.Error("Common Name extraction error: %v", err)
	}
	bc.brokerConn.queueName = bc.brokerConn.routingKey

	// TLS config.
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   bc.serverAddr,
		MinVersion:   tls.VersionTLS12,
	}

	err = bc.brokerConnectionConfig(tlsConfig)

	return err
}

// ExecuteBrokerClient executes the Network Manager Broker routines.
func (bc *BrokerClient) ExecuteBrokerClient(cl client.Client) error {
	// Start sending messages
	klog.Info("executing broker client routines")
	var err error
	if bc.pubFlag {
		go func() {
			bc.publishOnBroker()
		}()
	}

	// Start receiving messages
	if bc.subFlag {
		go func() {
			if err = bc.readMsgOnBroker(bc.ctx, cl); err != nil {
				klog.ErrorS(err, "error receiving advertisement")
			}
		}()
	}
	return err
}

func (bc *BrokerClient) publishOnBroker() {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:

			// Pub on exchange
			err := bc.brokerConn.amqpChan.Publish(
				bc.brokerConn.exchangeName,
				bc.brokerConn.routingKey,
				true,  // Mandatory: if not routable -> error
				false, // Immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        bc.brokerConn.outboundMsg,
					Expiration:  "30000", // TTL ms
				})
			if err != nil {
				klog.Error("Error pub message: %v", err)
			}

			select {
			case confirm := <-bc.brokerConn.confirms:
				if confirm.Ack {
					klog.Info("Message successfully published!")
				} else {
					klog.Info("Message failed to publish!")
				}
			case <-time.After(5 * time.Second): // Timeout
				klog.Info("No confirmation received, message status unknown.")
			}

		case <-bc.ctx.Done():
			ticker.Stop()
			klog.Info("Ticker stopped\n")
			return
		}
	}
}

func (bc *BrokerClient) readMsgOnBroker(ctx context.Context, cl client.Client) error {
	klog.Info("Listening from Broker")
	for d := range bc.brokerConn.inboundMsgs {
		klog.Info("Received remote advertisement from BROKER\n")
		var remote NetworkManager
		err := json.Unmarshal(d.Body, &remote.ID)
		if err != nil {
			klog.Error("Error unmarshalling message: ", err)
			continue
		}
		// Check if received advertisement is remote
		if bc.ID.IP != remote.ID.IP {
			// Create knownCluster CR
			kc := &networkv1alpha1.KnownCluster{}

			if err := cl.Get(ctx, client.ObjectKey{Name: namings.ForgeKnownClusterName(remote.ID.NodeID), Namespace: flags.FluidosNamespace}, kc); err != nil {
				if client.IgnoreNotFound(err) == nil {
					klog.InfoS("KnownCluster not found: creating form Broker", remote.ID.NodeID)

					// Create new KnownCluster CR
					if err := cl.Create(ctx, resourceforge.ForgeKnownCluster(remote.ID.NodeID, remote.ID.IP)); err != nil {
						return err
					}
					klog.InfoS("KnownCluster created from Broker", "ID", remote.ID.NodeID)
				}
			} else {
				klog.Info("KnownCluster already present: updating from Broker", remote.ID.NodeID)
				kc.UpdateStatus()

				// Update fetched KnownCluster CR
				err := cl.Status().Update(ctx, kc)
				if err != nil {
					return err
				}
				klog.InfoS("KnownCluster updated from Broker", "ID", kc.ObjectMeta.Name)
			}
		}
	}
	return nil
}

func extractCNfromCert(certPEM *[]byte) (string, error) {
	var err error
	var cert *x509.Certificate
	var CN = ""

	// Decode PEM cert
	block, _ := pem.Decode(*certPEM)
	if block == nil {
		klog.Error("Error decoding certificate PEM in CN extraction")
	} else {
		// Parsing X.509
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			klog.Error("Error parsing certificate X.509 in CN extraction: %v", err)
		} else {
			CN = cert.Subject.CommonName
		}
	}
	return strings.TrimSpace(CN), err
}

func (bc *BrokerClient) brokerConnectionConfig(tlsConfig *tls.Config) error {
	var err error
	config := amqp.Config{
		SASL:            []amqp.Authentication{&amqp.ExternalAuth{}}, // auth EXTERNAL
		TLSClientConfig: tlsConfig,                                   // config TLS
		Vhost:           "/",                                         // vhost
		Heartbeat:       10 * time.Second,                            // heartbeat
	}

	// Config connection
	serverURL := "amqps://" + bc.serverAddr + ":5671/"

	bc.brokerConn.amqpConn, err = amqp.DialConfig(serverURL, config)
	if err != nil {
		klog.Error("RabbitMQ connection error: %v", err)
		return err
	}

	// Channel creation
	bc.brokerConn.amqpChan, err = bc.brokerConn.amqpConn.Channel()
	if err != nil {
		klog.Error("channel creation error: %v", err)
		return err
	}

	// Queue subscrition
	bc.brokerConn.inboundMsgs, err = bc.brokerConn.amqpChan.Consume(
		bc.brokerConn.queueName, // queue name
		"",                      // consumer name (empty -> generated)
		true,                    // AutoAck
		false,                   // Exclusive: queue is accessible only from this consumer
		true,                    // false,        // NoLocal: does not receive selfpublished messages
		false,                   // NoWait: server confirmation
		nil,                     // Arguments
	)
	if err != nil {
		klog.Error("Error subscribing queue: %s", err)
		return err
	}

	// Write confirm broker
	if err := bc.brokerConn.amqpChan.Confirm(false); err != nil {
		klog.Error("Failed to enable publisher confirms: %v", err)
		return err
	}

	// Channels for write confirm
	bc.brokerConn.confirms = bc.brokerConn.amqpChan.NotifyPublish(make(chan amqp.Confirmation, 1))

	klog.InfoS("Node", "ID", bc.ID.NodeID, "Client Address", bc.ID.IP, "Server Address", bc.serverAddr, "RoutingKey", bc.brokerConn.routingKey)

	return nil
}

func (bc *BrokerClient) extractSecret(cl client.Client, secretName, secretNamespace string, secretDest *corev1.Secret) error {
	err := cl.Get(context.TODO(), client.ObjectKey{
		Name:      secretName,
		Namespace: secretNamespace,
	}, secretDest)
	if err != nil {
		klog.Error("Error retrieving Secret: %v\n", err)
		return err
	}
	return nil
}
