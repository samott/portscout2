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
	Name       string
	Version    string
	NewVersion *string `db:"newVersion"`
	Category   string
	CheckedAt  *time.Time `db:"checkedAt"`
	UpdatedAt  *time.Time `db:"updatedAt"`
	Maintainer string
	GitHub     *string `db:"gitHub"`
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

	query := db.gdb.Insert("ports").Rows(goqu.Record{
		"name":       port.Name.Name,
		"category":   port.Name.Category,
		"version":    port.DistVersion,
		"maintainer": port.Maintainer,
		"gitHub":     github,
	}).OnConflict(goqu.DoUpdate(
		"category, name",
		goqu.Record{
			"version": port.DistVersion,
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
		"category": port.Name,
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

		ports = append(ports, types.PortInfo{
			Name: types.PortName{
				Category: row.Category,
				Name:     row.Name,
			},
			Maintainer: row.Maintainer,
			GitHub:     github,
		})
	}

	return ports, nil
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
