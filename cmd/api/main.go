package main

import (
	"net/http"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"regexp"

	"github.com/samott/portscout2/config"
	"github.com/samott/portscout2/db"
	"github.com/samott/portscout2/types"
)

type Api struct {
	db  *db.DB
	cfg *config.Config
}

var emailRegex = regexp.MustCompile(
	`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`,
)

func (api *Api) getByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.PathValue("category")

	ports, err := api.db.GetPortUpdates(&category, nil)

	if err != nil {
		slog.Error("Error", "err", err)
	}

	result, err := json.Marshal(struct {
		Ports []types.PortUpdate `json:"ports"`
	}{
		Ports: ports,
	})

	if err != nil {
		slog.Error("Error", "err", err)
	}

	w.Write(result)
}

func (api *Api) getByMaintainer(w http.ResponseWriter, r *http.Request) {
	maintainer := r.PathValue("maintainer")

	if !emailRegex.MatchString(maintainer) {
		http.Error(w, "missing or invalid parameter: maintainer", http.StatusBadRequest)
	}

	ports, err := api.db.GetPortUpdates(nil, &maintainer)

	if err != nil {
		slog.Error("Error", "err", err)
	}

	result, err := json.Marshal(struct {
		Ports []types.PortUpdate `json:"ports"`
	}{
		Ports: ports,
	})

	if err != nil {
		slog.Error("Error", "err", err)
	}

	w.Write(result)
}

func main() {
	mux := http.NewServeMux()

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

	api := &Api{
		db:  db,
		cfg: cfg,
	}

	mux.HandleFunc("GET /updates/category/{category}", api.getByCategory)
	mux.HandleFunc("GET /updates/maintainer/{maintainer}", api.getByMaintainer)

	listenStr := fmt.Sprintf(":%d", cfg.Api.Port)

	slog.Info("Starting API service...", "peer", listenStr)

	err = http.ListenAndServe(listenStr, mux)

	if err != nil {
		log.Fatal(err)
	}
}
