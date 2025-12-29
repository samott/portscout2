package db

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/samott/portscout2/types"
)

type DB struct {
	db  *sql.DB
	gdb *goqu.Database
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

	query := db.gdb.Insert("ports").Rows(goqu.Record{
		"name":     port.Name.Name,
		"category": port.Name.Category,
		"version":  port.DistVersion,
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
