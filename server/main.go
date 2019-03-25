package main

import (
	"flag"
	"log"

	"github.com/sirupsen/logrus"

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

func newLogger(cfg *config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(cfg.Level)
	return logger
}
