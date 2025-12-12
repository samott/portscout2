package main

import (
	"log/slog"

	"github.com/samott/portscout2/repo"
)

func main() {
	slog.Info("portscout2 running...")

	ports := repo.FindUpdated("/usr/ports", "b700f9a18a81834a7e5c2046cda87c290bfa229a")

	slog.Info("Ports", "ports", ports)
}
