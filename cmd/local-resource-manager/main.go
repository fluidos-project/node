package main

import (
	"context"
	"flag"
	"fmt"

	localResourceManager "fluidos.eu/node/pkg/local-resource-manager"
	"fluidos.eu/node/pkg/utils/flags"
	"fluidos.eu/node/pkg/utils/services"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func main() {
	done := make(chan bool)
	flag.StringVar(&flags.DOMAIN, "domain", "fluidos.eu", "Domain name")
	flag.StringVar(&flags.IP_ADDR, "ip", "", "IP address of the node")
	flag.StringVar(&flags.AMOUNT, "amount", "", "Amount of money set for the flavours of this node")
	flag.StringVar(&flags.CURRENCY, "currency", "", "Currency of the money set for the flavours of this node")
	flag.StringVar(&flags.PERIOD, "period", "", "Period set for the flavours of this node")
	flag.StringVar(&flags.FLAVOUR_DEFAULT_NAMESPACE, "flavour-namespace", "default", "Namespace where the flavour CRs are created")
	flag.StringVar(&flags.RESOURCE_TYPE, "resources-types", "k8s-fluidos", "Type of the Flavour related to k8s resources")
	flag.Int64Var(&flags.CPU_MIN, "cpu-min", 0, "Minimum CPU value")
	flag.Int64Var(&flags.MEMORY_MIN, "memory-min", 0, "Minimum memory value")
	flag.Int64Var(&flags.CPU_STEP, "cpu-step", 0, "CPU step value")
	flag.Int64Var(&flags.MEMORY_STEP, "memory-step", 0, "Memory step value")
	flag.Int64Var(&flags.MIN_COUNT, "min-count", 0, "Minimum number of flavours")
	flag.Int64Var(&flags.MAX_COUNT, "max-count", 0, "Maximum number of flavours")

	flag.Parse()
	ctx := context.Background()

	cl, err := services.GetKClient(ctx)
	utilruntime.Must(err)

	localResourceManager.StartController(cl)
	fmt.Println("Started controller for monitoring local resources")

	<-done
}
