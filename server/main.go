package main

import (
	"flag"
	"log"

	"github.com/memprofiler/memprofiler/server/config"
)

func main() {
	cfgPath := flag.String("c", "", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfigFromFile(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}
