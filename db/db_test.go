package db

import (
	"flag"
	"log/slog"
	"os"
	"testing"

	"github.com/samott/portscout2/config"
	"github.com/samott/portscout2/types"
)

var db *DB

func TestMain(m *testing.M) {
	configFile := flag.String("config", "../portscout.test.yaml", "path to configuration file")

	flag.Parse()

	cfg, err := config.LoadConfig(*configFile)

	if err != nil {
		slog.Error("Failed to load config file " + *configFile)
		os.Exit(1)
	}

	db, err = NewDB(cfg.Db.Url)

	code := m.Run()

	os.Exit(code)
}

func TestUpdatePort(t *testing.T) {
	err := db.RemovePort(types.PortName{
		Name:     "test",
		Category: "cat",
	})

	if err != nil {
		t.Fatal("RemovePort failed")
	}

	err = db.UpdatePort(types.PortInfo{
		Name: types.PortName{
			Name:     "test",
			Category: "cat",
		},
		Maintainer: "test@example.net",
	})

	if err != nil {
		t.Fatal("UpdatePort failed")
	}

	port, err := db.GetPortByName(types.PortName{
		Name:     "test",
		Category: "cat",
	})

	if err != nil {
		t.Fatal("GetPortByName failed")
	}

	if port == nil {
		t.Fatal("Port not found")
	}

	if port.Name.Name != "test" || port.Name.Category != "cat" || port.Maintainer != "test@example.net" {
		t.Fatal("Incorrect field value")
	}

	err = db.UpdatePort(types.PortInfo{
		Name: types.PortName{
			Name:     "test",
			Category: "cat",
		},
		Maintainer: "newmaintainer@example.net",
	})

	if err != nil {
		t.Fatal("UpdatePort failed")
	}

	port, err = db.GetPortByName(types.PortName{
		Name:     "test",
		Category: "cat",
	})

	if err != nil {
		t.Fatal("GetPortByName failed")
	}

	if port == nil {
		t.Fatal("Port not found")
	}

	if port.Name.Name != "test" || port.Name.Category != "cat" || port.Maintainer != "newmaintainer@example.net" {
		t.Fatal("Incorrect field value")
	}

	err = db.RemovePort(types.PortName{
		Name:     "test",
		Category: "cat",
	})

	if err != nil {
		t.Fatal("RemovePort failed")
	}

	port, err = db.GetPortByName(types.PortName{
		Name:     "test",
		Category: "cat",
	})

	if err != nil {
		t.Fatal("GetPortByName failed")
	}

	if port != nil {
		t.Fatal("Port not removed")
	}
}

func TestGetPortUpdates(t *testing.T) {
	err := db.RemovePort(types.PortName{
		Name:     "test-44",
		Category: "mycat-1",
	})

	if err != nil {
		t.Fatal("RemovePort failed")
	}

	err = db.RemovePort(types.PortName{
		Name:     "test-16",
		Category: "mycat-2",
	})

	if err != nil {
		t.Fatal("RemovePort failed")
	}

	err = db.UpdatePort(types.PortInfo{
		Name: types.PortName{
			Name:     "test-44",
			Category: "mycat-1",
		},
		Maintainer: "mycatguy@example.net",
	})

	if err != nil {
		t.Fatal("UpdatePort failed")
	}

	err = db.UpdatePort(types.PortInfo{
		Name: types.PortName{
			Name:     "test-16",
			Category: "mycat-2",
		},
		Maintainer: "mycatguy@example.net",
	})

	if err != nil {
		t.Fatal("UpdatePort failed")
	}

	cat := "mycat-1"
	maintainer := "mycatguy@example.net"

	updates, err := db.GetPortUpdates(&cat, nil)

	if err != nil {
		t.Fatal("GetPortUpdates failed")
	}

	if len(updates) != 1 {
		t.Fatal("Incorrect number of updates found for category")
	}

	updates, err = db.GetPortUpdates(nil, &maintainer)

	if err != nil {
		t.Fatal("GetPortUpdates failed")
	}

	if len(updates) != 2 {
		t.Fatal("Incorrect number of updates found for maintainer")
	}
}
