package main

import (
	"flag"
	"os"

	"log/slog"

	"github.com/samott/portscout2/config"
	"github.com/samott/portscout2/db"
	"github.com/samott/portscout2/repo"
	"github.com/samott/portscout2/tree"
	"github.com/samott/portscout2/types"
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

	db, err := db.NewDB(cfg.Db.Url)

	if err != nil {
		slog.Error("Failed to connect to database")
		os.Exit(1)
	}

	ports := repo.FindUpdated(cfg.Tree.PortsDir, "b700f9a18a81834a7e5c2046cda87c290bfa229a")

	slog.Info("Ports", "ports", ports)

	tr := tree.NewTree(cfg.Tree.MakeCmd, cfg.Tree.PortsDir, cfg.Tree.MakeThreads)

	updatedPorts := make([]types.PortName, 0, len(ports))
	removedPorts := make([]types.PortName, 0)

	for name, change := range ports {
		if change == repo.PortRemoved {
			removedPorts = append(removedPorts, name)
		} else {
			updatedPorts = append(updatedPorts, name)
		}
	}

	db.RemovePorts(removedPorts)

	_, err = tr.QueryPorts(updatedPorts, func(pi types.PortInfo) {
		slog.Info("Port", "info", pi)
	})

	if err != nil {
		slog.Error("Failed to query all ports", "err", err)
		os.Exit(1)
	}
}
