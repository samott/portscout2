package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/samott/portscout2/types"
)

type DB struct {
	db  *sql.DB
	gdb *goqu.Database
}

type portEntry struct {
	Name        string
	Version     string
	NewVersion  *string `db:"newVersion"`
	Category    string
	CheckedAt   *time.Time `db:"checkedAt"`
	UpdatedAt   *time.Time `db:"updatedAt"`
	Portscout   string     `db:"portscout"`
	Maintainer  string
	MasterSites string  `db:"masterSites"`
	DistFiles   string  `db:"distFiles"`
	GitHub      *string `db:"gitHub"`
	Config      string  `db:"portConfig"`
}

func NewDB(dbUrl string) (*DB, error) {
	db, err := sql.Open("pgx", dbUrl)

	if err != nil {
		return nil, err
	}

	return &DB{
		db:  db,
		gdb: goqu.New("postgres", db),
	}, nil
}

func (db *DB) Close() {
	if db.db != nil {
		db.db.Close()
	}
}

func (db *DB) UpdatePort(port types.PortInfo) error {
	slog.Info("Updating port", "port", port.Name)

	var github *string

	if port.GitHub != nil {
		ghbytes, err := json.Marshal(port.GitHub)
		if err != nil {
			return fmt.Errorf("Unable to marshal GitHub field to JSON: %w", err)
		}
		ghstring := string(ghbytes)
		github = &ghstring
	} else {
		github = nil
	}

	pcbytes, err := json.Marshal(port.Config)
	if err != nil {
		return fmt.Errorf("Unable to marshal PortConfig field to JSON: %w", err)
	}
	portConfig := string(pcbytes)

	masterSites := types.MarshalTaggedLists(port.MasterSites)
	distFiles := types.MarshalTaggedLists(port.DistFiles)

	query := db.gdb.Insert("ports").Rows(goqu.Record{
		"name":        port.Name.Name,
		"category":    port.Name.Category,
		"version":     port.DistVersion,
		"maintainer":  port.Maintainer,
		"masterSites": masterSites,
		"distFiles":   distFiles,
		"gitHub":      github,
		"portscout":   port.Portscout,
		"portConfig":  portConfig,
	}).OnConflict(goqu.DoUpdate(
		"category, name",
		goqu.Record{
			"name":        port.Name.Name,
			"category":    port.Name.Category,
			"version":     port.DistVersion,
			"maintainer":  port.Maintainer,
			"masterSites": masterSites,
			"distFiles":   distFiles,
			"gitHub":      github,
			"portscout":   port.Portscout,
			"portConfig":  portConfig,
		},
	)).Prepared(true)

	sql, args, err := query.ToSQL()

	if err != nil {
		return err
	}

	_, err = db.db.Exec(sql, args...)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) RemovePort(port types.PortName) error {
	slog.Info("Removing port", "port", port.Name)

	query := db.gdb.From("ports").Delete().Where(goqu.Ex{
		"name":     port.Name,
		"category": port.Category,
	}).Prepared(true)

	sql, args, err := query.ToSQL()

	if err != nil {
		return err
	}

	_, err = db.db.Exec(sql, args...)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) RemovePorts(ports []types.PortName) error {
	// TODO: reimplement more efficiently
	for _, port := range ports {
		err := db.RemovePort(port)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetPorts(limit uint, offset uint) ([]types.PortInfo, error) {
	query := db.gdb.From("ports").Limit(limit).Offset(offset).Prepared(true)

	var rows []portEntry

	ports := make([]types.PortInfo, 0, limit)

	err := query.ScanStructs(&rows)

	if err != nil {
		return nil, fmt.Errorf("Error while scanning structs: %w", err)
	}

	for _, row := range rows {
		var github *types.GitHubInfo

		if row.GitHub != nil {
			err := json.Unmarshal([]byte(*row.GitHub), &github)

			if err != nil {
				return nil, fmt.Errorf("Error while unmarshalling GitHub JSON: %w", err)
			}
		} else {
			github = nil
		}

		masterSites := types.UnmarshalTaggedLists(row.MasterSites)
		distFiles := types.UnmarshalTaggedLists(row.DistFiles)

		var portConfig types.PortConfig

		err := json.Unmarshal([]byte(row.Config), &portConfig)

		if err != nil {
			return nil, fmt.Errorf("Error while unmarshalling PortConfig JSON: %w", err)
		}

		ports = append(ports, types.PortInfo{
			Name: types.PortName{
				Category: row.Category,
				Name:     row.Name,
			},
			Portscout:   row.Portscout,
			Maintainer:  row.Maintainer,
			MasterSites: masterSites,
			DistFiles:   distFiles,
			GitHub:      github,
			Config:      portConfig,
		})
	}

	return ports, nil
}

func (db *DB) GetPortByName(portName types.PortName) (*types.PortInfo, error) {
	query := db.gdb.From("ports").
		Where(goqu.Ex{
			"name":     portName.Name,
			"category": portName.Category,
		}).Prepared(true)

	var row portEntry
	var port types.PortInfo

	_, err := query.ScanStruct(&row)

	if err != nil {
		return nil, fmt.Errorf("Error while scanning structs: %w", err)
	}

	var github *types.GitHubInfo

	if row.GitHub != nil {
		err := json.Unmarshal([]byte(*row.GitHub), &github)

		if err != nil {
			return nil, fmt.Errorf("Error while unmarshalling GitHub JSON: %w", err)
		}
	} else {
		github = nil
	}

	masterSites := types.UnmarshalTaggedLists(row.MasterSites)
	distFiles := types.UnmarshalTaggedLists(row.DistFiles)

	var portConfig types.PortConfig

	err = json.Unmarshal([]byte(row.Config), &portConfig)

	if err != nil {
		return nil, fmt.Errorf("Error while unmarshalling PortConfig JSON: %w", err)
	}

	port = types.PortInfo{
		Name: types.PortName{
			Category: row.Category,
			Name:     row.Name,
		},
		Portscout:   row.Portscout,
		Maintainer:  row.Maintainer,
		MasterSites: masterSites,
		DistFiles:   distFiles,
		GitHub:      github,
		Config:      portConfig,
	}

	return &port, nil
}

func (db *DB) GetPortUpdates(category *string, maintainer *string) ([]types.PortUpdate, error) {
	query := db.gdb.From("ports").Prepared(true)

	if category != nil {
		query = query.Where(goqu.Ex{
			"category": category,
		})
	}

	if maintainer != nil {
		query = query.Where(goqu.Ex{
			"maintainer": maintainer,
		})
	}

	var rows []portEntry

	ports := make([]types.PortUpdate, 0)

	err := query.ScanStructs(&rows)

	if err != nil {
		return nil, fmt.Errorf("Error while scanning structs: %w", err)
	}

	for _, row := range rows {
		ports = append(ports, types.PortUpdate{
			Name: types.PortName{
				Category: row.Category,
				Name:     row.Name,
			},
			Maintainer: row.Maintainer,
			Version:    row.Version,
			NewVersion: row.NewVersion,
			UpdatedAt:  row.UpdatedAt,
			CheckedAt:  row.CheckedAt,
		})
	}

	return ports, nil
}

func (db *DB) GetMaintainerStats() ([]types.MaintainerStats, error) {
	inner := goqu.
		From("portdata").
		Select(
			goqu.L("LOWER(maintainer)").As("maintainer"),
			goqu.COUNT("*").As("total"),
			goqu.L(`
				COUNT(*) FILTER (
					WHERE newVersion IS NOT NULL
					  AND newVersion != version
				)
			`).As("withNewDistfile"),
		).
		GroupBy(
			goqu.L("LOWER(maintainer)"),
		)

	query := goqu.
		From(inner.As("pd1")).
		Select(
			goqu.C("maintainer"),
			goqu.C("total"),
			goqu.COALESCE(goqu.C("withNewDistfile"), 0).As("withNewDistfile"),
			goqu.L(`
				100.0 * COALESCE(withNewDistfile, 0) / total
			`).As("percentage"),
		)

	sql, args, err := query.ToSQL()

	if err != nil {
		return nil, err
	}

	_, err = db.db.Exec(sql, args...)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (db *DB) GetLastCommit() (string, error) {
	query := db.gdb.From("repo").Select("lastCommit").Limit(1).Prepared(true)

	var lastCommit string
	found, err := query.ScanVal(&lastCommit)

	if err != nil {
		return "", err
	}

	if !found {
		return "", errors.New("tree state table not found in database")
	}

	return lastCommit, nil
}

func (db *DB) SetLastCommit(lastCommit string) error {
	query := db.gdb.Update("repo").Set(
		goqu.Record{
			"lastCommit": lastCommit,
		},
	)

	sql, args, err := query.ToSQL()

	if err != nil {
		return err
	}

	_, err = db.db.Exec(sql, args...)

	if err != nil {
		return err
	}

	return nil
}
