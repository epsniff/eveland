package dbsdeutils

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const DBNAME = "eve_sde.sqlite"

func LoadDB(basepath string) (*sql.DB, error) {
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

func ShowAllColumns(basepath string, tableName string) error {

	db, err := LoadDB(basepath)
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	err = listTableColumns(db, tableName)
	if err != nil {
		return fmt.Errorf("error listing table columns: %v", err)
	}

	return nil
}

func listTableColumns(db *sql.DB, tableName string) error {
	// Execute the query to get the column names of the table
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return err
	}

	// Loop through the rows and print the column names
	fmt.Printf("Columns of table %s:\n", tableName)
	colNames := make([]string, 0)
	for rows.Next() {

		var cid int
		var name string
		var dataType string
		var notNull bool
		var defaultValue sql.NullString
		var pk int // pk is a bool, but we use int 1 or 0 as true/false

		err = rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			rows.Close()
			return err
		}

		// Convert integer value of "pk" to boolean value
		var primaryKey bool
		if pk == 1 {
			primaryKey = true
		}

		colNames = append(colNames, name)
		fmt.Printf("name:%v isPrimary:%v type:%v\n", name, primaryKey, dataType)
	}
	rows.Close()

	// Query the database for the first 3 rows from the table "mytable"
	rows, err = db.Query("SELECT * FROM " + tableName + " LIMIT 3") // TODO make this Random out of a much larger set.
	if err != nil {
		return fmt.Errorf("error querying table: %v", err)
	}
	defer rows.Close()

	// Iterate over the rows and print each one
	for rows.Next() {
		values := make([]interface{}, len(colNames))
		for i := range values {
			values[i] = new(interface{})
		}

		err = rows.Scan(values...)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}

		// Print out the values of the row
		for i, value := range values {
			fmt.Printf("%s: %v ", colNames[i], *value.(*interface{}))
		}
		fmt.Println()
	}

	return nil
}

func ShowAllTables(basepath string) error {
	db, err := LoadDB(basepath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return fmt.Errorf("error querying tables: %v", err)
	}
	defer rows.Close()

	fmt.Println("Tables:")
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			return fmt.Errorf("error scanning table name: %v", err)
		}
		fmt.Println(tableName)
	}

	return nil
}
