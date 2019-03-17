package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/memprofiler/memprofiler/test/reporter/lib"
	"github.com/memprofiler/memprofiler/utils"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "-c", "", "configuration file")
	flag.Parse()

	logger := logrus.New()

	// prepare config
	cfg, err := lib.FromYAMLFile(configPath)
	if err != nil {
		logger.Fatal(err)
	}

	// launch application
	errChan := make(chan error, 1)
	ln, err := lib.NewLauncher(cfg, logger, errChan)
	if err != nil {
		logger.Fatal(err)
	}
	defer ln.Quit()

	utils.BlockOnSignal(logger, errChan)
}
