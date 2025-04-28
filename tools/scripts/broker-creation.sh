#!/usr/bin/bash

read_input() {
    local prompt="$1"
    local var_name="$2"

    read -p "$prompt: " -r value
    eval "$var_name=\"$value\""
}

if [[ -n "${KUBECONFIG}" ]]; then
  kubeconfig=$KUBECONFIG
else read_input "KUBECONFIG not set, please set it" kubeconfig
fi

broker_name="null"
address="null"
broker_ca_cert="null"
broker_client_cert="null"
broker_priv_key="null"
role="null"
rule_file="null"
metric_file="null"

read_input "Broker's name of your choice" "broker_name"
read_input "Broker server address (must match certificate CN)" "address"
read_input ".pem ROOT certificate" "broker_ca_cert"
read_input ".pem client certificate" "broker_client_cert"
read_input ".pem private key" "broker_priv_key"
read_input "Type the role: publisher | subscriber | both" "role"
read_input ".json file for RULE" "rule_file"
read_input ".json file for metrics" "metric_file"

rule_json=$(jq -c . "$rule_file" | sed 's/"/\\"/g')
metric_json=$(jq -c . "$metric_file" | sed 's/"/\\"/g')

broker_ca_secret="$broker_name"-ca-"$RANDOM"
broker_client_secret="$broker_name"-cl-"$RANDOM"

#create the secrets
kubectl create secret tls $broker_client_secret --cert=$broker_client_cert --key=$broker_priv_key --namespace=fluidos --kubeconfig "$kubeconfig"
status=$?
if [ "$status" -ne 0 ]; then 
  exit 1
fi

kubectl create secret generic $broker_ca_secret --from-file=$broker_ca_cert --namespace=fluidos  --kubeconfig "$kubeconfig"
status=$?
if [ "$status" -ne 0 ]; then 
  kubectl delete secret broker_client_secret -n fluidos 
  exit 1
fi

# cr yaml
cat <<EOF > ./$broker_name.yaml
apiVersion: network.fluidos.eu/v1alpha1
kind: Broker
metadata:
  name: $broker_name
  namespace: fluidos
spec:
  name: $broker_name
  address: $address
  role: $role
  rule: "$rule_json"
  metric: "$metric_json"
  cacert: $broker_ca_secret
  clcert: $broker_client_secret

EOF

if [ -f "$broker_name.yaml" ]; then
  kubectl apply -f $broker_name.yaml
  rm -f "$broker_name.yaml"
fi
