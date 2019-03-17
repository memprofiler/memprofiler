package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/test/reporter/config"
	"github.com/memprofiler/memprofiler/test/reporter/launcher"
	"github.com/memprofiler/memprofiler/utils"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "-c", "", "configuration file")
	flag.Parse()

	logger := logrus.New()

	// prepare config
	cfg, err := config.FromYAMLFile(configPath)
	if err != nil {
		logger.Fatal(err)
	}

	// launch application
	errChan := make(chan error, 1)
	ln, err := launcher.New(logger, cfg, errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer ln.Quit()

	utils.BlockOnSignal(logger, errChan)
}
