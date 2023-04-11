package evesdedb

import (
	"database/sql"
	"fmt"
)

const DBNAME = "eve_sde.sqlite"

type EveSDEDB struct {
	evesde *sql.DB
}

func New(basepath string) (*EveSDEDB, error) {
	db, err := loadDB(basepath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	return &EveSDEDB{evesde: db}, nil
}

func (db *EveSDEDB) Close() error {
	return db.evesde.Close()
}
