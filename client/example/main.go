package example

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

func Example() {
	// prepare client configuration
	cfg := &client.Config{
		// server address
		ServerEndpoint: "localhost:46219",
		// description of your application instance
		ServiceDescription: &schema.ServiceDescription{
			ServiceType:     "test_application",
			ServiceInstance: "node_1",
		},
		// granularity
		Periodicity: &utils.Duration{Duration: time.Second},
		// logging setting
		Verbose: false,
	}

	// you can implement your own logger
	log := client.LoggerFromLogrus(logrus.New())

	// run profiler and stop it explicitly on exit
	profiler, err := client.NewProfiler(log, cfg)
	if err != nil {
		panic(err)
	}
	profiler.Start()
	defer profiler.Stop()

	// ...
}
