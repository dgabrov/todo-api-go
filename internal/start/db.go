package start

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/dgb9/todo-api/internal/data"
)

func connectDb(config data.DbConfig) (*sql.DB, error) {
	user := config.User
	password := config.Password
	database := config.Database
	machine := config.Machine
	port := config.Port

	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", user, password, machine, port, database)
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}

	// try to run some select something
	err = testConnection(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func testConnection(db *sql.DB) error {
	var i int
	rows, err := db.Query("select 1")
	if err != nil {
		return err
	}

	defer rows.Close()

	if !rows.Next() {
		return errors.New("expected at least one record, got none")
	}

	err = rows.Scan(&i)

	if err != nil {
		return err
	}

	if i != 1 {
		return fmt.Errorf("expected value 1 for test retrieval value, got %d", i)
	}

	return nil
}
