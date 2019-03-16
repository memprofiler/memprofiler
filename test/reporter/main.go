package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/client"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "-c", "", "configuration file")
	flag.Parse()

	logger := logrus.New()

	// prepare config
	cfg, err := FromYAMLFile(configPath)
	if err != nil {
		logger.Fatal(err)
	}

	// run memprofiler client
	pf, err := client.NewProfiler(client.LoggerFromLogrus(logger), cfg.Client)
	if err != nil {
		logger.Fatal(err)
	}
	defer pf.Quit()

	// run memory consumption scenario
	errChan := make(chan error, 1)
	pb := newPlayback(newContainer(), cfg.Scenario, errChan)
	pb.Start()
	defer pb.Stop()

	// wait for signals or fatal errors
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	select {
	case <-signalChan:
		logger.Warning("Interrupt signal has been received")
	case err := <-errChan:
		if err != nil {
			logger.WithError(err).Error("Fatal error, going to terminate server")
		} else {
			logger.Warning("Going to terminate server")
		}
	}

}
