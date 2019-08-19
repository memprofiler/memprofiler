package main

import (
	"flag"
	"log"
	"os"

	"github.com/rs/zerolog"

	"github.com/memprofiler/memprofiler/server/config"
	"github.com/memprofiler/memprofiler/server/launcher"
	"github.com/memprofiler/memprofiler/utils"
)

func main() {
	cfgPath := flag.String("c", "", "path to config file")
	flag.Parse()

	cfg, err := config.FromYAMLFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	logger := newLogger(cfg.Logging)

	errChan := make(chan error, 2)
	l, err := launcher.New(logger, cfg, errChan)
	if err != nil {
		log.Fatal(err)
	}
	l.Start()
	defer l.Stop()

	utils.BlockOnSignal(logger, errChan)
}

func newLogger(cfg *config.LoggingConfig) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(cfg.Level)
	return &logger
}
