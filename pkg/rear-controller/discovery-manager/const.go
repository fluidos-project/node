package discoverymanager

const (
	SERVER_ADDR       = "http://localhost:14144/api"
	K8S_TYPE          = "k8s-fluidos"
	DEFAULT_NAMESPACE = "default"
	CLIENT_ID         = "topix.fluidos.eu"
)

// We define different server addresses, much more dynamic and less hardcoded
var (
	SERVER_ADDRESSES = []string{"http://localhost:14144/api"}
)
