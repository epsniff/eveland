package evesdedb

import (
	"context"
	"database/sql"
	"fmt"
)

func (e *EveSDEDB) ShowAllTables(ctx context.Context) error {
	rows, err := e.evesde.Query("SELECT name FROM sqlite_master WHERE type='table'")
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

func (e *EveSDEDB) ShowAllColumns(ctx context.Context, tableName string) error {
	err := listTableColumns(e.evesde, tableName)
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
