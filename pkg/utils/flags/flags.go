package flags

import "time"

var (
	RESOURCE_TYPE                 string
	DEFAULT_NAMESPACE             string = "default"
	PC_DEFAULT_NAMESPACE          string = "default"
	SOLVER_DEFAULT_NAMESPACE      string = "default"
	DISCOVERY_DEFAULT_NAMESPACE   string = "default"
	FLAVOUR_DEFAULT_NAMESPACE     string = "default"
	CONTRACT_DEFAULT_NAMESPACE    string = "default"
	RESERVATION_DEFAULT_NAMESPACE string = "default"
	TRANSACTION_DEFAULT_NAMESPACE string = "default"
	CLIENT_ID                     string
	SERVER_ADDR                   string
	SERVER_ADDRESSES              = []string{SERVER_ADDR}
	HTTP_PORT                     string
	WorkerLabelKey                string
	DOMAIN                        string
	IP_ADDR                       string
	AMOUNT                        string
	CURRENCY                      string
	PERIOD                        string
	CPU_MIN                       int64
	MEMORY_MIN                    int64
	CPU_STEP                      int64
	MEMORY_STEP                   int64
	MIN_COUNT                     int64
	MAX_COUNT                     int64
	EXPIRATION_PHASE_RUNNING      = 2 * time.Minute
	EXPIRATION_SOLVER             = 5 * time.Minute
	EXPIRATION_TRANSACTION        = 20 * time.Second
	EXPIRATION_CONTRACT           = 365 * 24 * time.Hour
)
