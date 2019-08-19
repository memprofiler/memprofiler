package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/test/reporter/config"
	"github.com/memprofiler/memprofiler/test/reporter/launcher"
	"github.com/memprofiler/memprofiler/utils"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "-c", "", "configuration file")
	flag.Parse()

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// prepare config
	cfg, err := config.FromYAMLFile(configPath)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	// launch application
	errChan := make(chan error, 1)
	l, err := launcher.New(&logger, cfg, errChan)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}
	l.Start()
	defer l.Stop()

	utils.BlockOnSignal(&logger, errChan)
}
