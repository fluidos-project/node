package flags

import "time"

// NAMESPACES flags
var (
	DEFAULT_NAMESPACE             string = "default"
	PC_DEFAULT_NAMESPACE          string = "default"
	SOLVER_DEFAULT_NAMESPACE      string = "default"
	DISCOVERY_DEFAULT_NAMESPACE   string = "default"
	FLAVOUR_DEFAULT_NAMESPACE     string = "default"
	CONTRACT_DEFAULT_NAMESPACE    string = "default"
	RESERVATION_DEFAULT_NAMESPACE string = "default"
	TRANSACTION_DEFAULT_NAMESPACE string = "default"
)

// EXPIRATION flags
var (
	EXPIRATION_PHASE_RUNNING = 2 * time.Minute
	EXPIRATION_SOLVER        = 5 * time.Minute
	EXPIRATION_TRANSACTION   = 20 * time.Second
	EXPIRATION_CONTRACT      = 365 * 24 * time.Hour
)

// TODO: TO BE REVIEWED
var (
	// THESE SHOULD BE THE NODE IDENTITY OF THE FLUIDOS NODE
	CLIENT_ID string
	DOMAIN    string
	IP_ADDR   string
	// THIS SHOULD BE PROVIDED BY THE NETWORK MANAGER
	SERVER_ADDR      string
	SERVER_ADDRESSES = []string{SERVER_ADDR}
	// REAR Gateway http port
	HTTP_PORT string
	// REAR Gateway grpc address
	GRPC_PORT string
)

var (
	RESOURCE_TYPE string
	AMOUNT        string
	CURRENCY      string
	PERIOD        string
	CPU_MIN       int64
	MEMORY_MIN    int64
	CPU_STEP      int64
	MEMORY_STEP   int64
	MIN_COUNT     int64
	MAX_COUNT     int64
)
