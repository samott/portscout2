package main

import (
	"flag"
	"os"

	"log/slog"

	"github.com/samott/portscout2/config"
	"github.com/samott/portscout2/repo"
)

func main() {
	slog.Info("portscout2 running...")

	configFile := flag.String("config", "portscout.yaml", "path to configuration file")

	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)

	if err != nil {
		slog.Error("Failed to load config file " + *configFile)
		os.Exit(1)
	}

	ports := repo.FindUpdated(cfg.PortsDir, "b700f9a18a81834a7e5c2046cda87c290bfa229a")

	slog.Info("Ports", "ports", ports)
}
