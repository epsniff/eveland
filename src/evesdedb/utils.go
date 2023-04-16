package evesdedb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func loadDB(basepath string) (*sql.DB, error) {
	var sqliteFile = basepath + "/" + DBNAME
	// error if database file does not exist
	if _, err := os.Stat(sqliteFile); os.IsNotExist(err) {
		// go download the latest SDE sqllite database from https://www.fuzzwork.co.uk/dump/
		// and put it in the basepath:
		//   $ curl -O https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
		//   $ bunzip2 sqlite-latest.sqlite.bz2
		//   $ mv sqlite-latest.sqlite _data/eve_sde.sqlite
		return nil, fmt.Errorf("SDE database file does not exist: %s", sqliteFile)
	}
	db, err := sql.Open("sqlite3", sqliteFile)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}
	return db, nil
}
