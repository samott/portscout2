package main

import (
	"flag"
	"os"

	"context"
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

	// Stage 1: sync the database from repo and ports tree

	lastCommitHash, err := db.GetLastCommit()

	if err != nil {
		slog.Error("Failed to get last commit hash", "err", err)
		os.Exit(1)
	}

	var ports map[types.PortName]repo.PortChange
	var headHash string

	if lastCommitHash == "" {
		headHash, ports, err = repo.FindAllPorts(cfg.Tree.PortsDir)
	} else {
		headHash, ports, err = repo.FindUpdated(cfg.Tree.PortsDir, lastCommitHash)
	}

	if err != nil {
		slog.Error("Failed to query ports repo", "err", err)
		os.Exit(1)
	}

	slog.Info("Ports", "ports", ports)

	tr := tree.NewTree(cfg.Tree.MakeCmd, cfg.Tree.PortsDir, cfg.Tree.MakeThreads)
	ctx, cancel := context.WithCancel(context.Background())

	if len(ports) > 0 {
		go tr.QueryPorts(ctx)

		go func() {
			for name, change := range ports {
				if change == repo.PortRemoved {
					err := db.RemovePort(name)

					if err != nil {
						slog.Error("Error removing database entry", "err", err)
						cancel()
						break
					}
				} else {
					tr.In() <- tree.QueryJob{Port: name}
				}
			}
			close(tr.In())
		}()

		for port := range tr.Out() {
			if port.Err == nil {
				err := db.UpdatePort(port.Info)

				if err != nil {
					slog.Error("Error updating database entry", "err", err)
					cancel()
					break
				}
			}
		}
	}

	if ctx.Err != nil {
		slog.Error("Aborting due to context error")
		os.Exit(1)
	}

	err = db.SetLastCommit(headHash)

	if err != nil {
		slog.Error("Failed to set last commit hash", "err", err)
		os.Exit(1)
	}

	// Stage 2: find updates

	// ...
}
