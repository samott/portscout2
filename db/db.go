package db

import (
	"database/sql"
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

func (db *DB) RemovePort(port types.PortInfo) error {
	slog.Info("Removing port", "port", port.Name)

	query := db.gdb.From("ports").Delete().Where(goqu.Ex{
		"name":     port.Name.Name,
		"category": port.Name.Category,
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

func (db *DB) RemovePorts(ports []types.PortInfo) error {
	// TODO: reimplement more efficiently
	for _, port := range ports {
		err := db.RemovePort(port)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) GetPorts(limit int, offset int) {
}
