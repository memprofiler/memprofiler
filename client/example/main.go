package example

import (
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/client"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/utils"
)

// Example of client integration
func Example() {
	// prepare client configuration
	cfg := &client.Config{
		// server address
		ServerEndpoint: "localhost:46219",
		// description of your application instance
		InstanceDescription: &schema.InstanceDescription{
			ServiceName:  "test_application",
			InstanceName: "node_1",
		},
		// granularity
		Periodicity: &utils.Duration{Duration: time.Second},
		// logging setting
		Verbose: false,
	}

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// you can implement your own logger
	logger := client.LoggerFromZeroLog(&log)

	// run profiler and stop it explicitly on exit
	profiler, err := client.NewProfiler(logger, cfg)
	if err != nil {
		panic(err)
	}
	profiler.Start()
	defer profiler.Stop()

	// ...
}
