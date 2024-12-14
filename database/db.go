package database

import (
	"database/sql"
	"log"
)

func Initdb() {

	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
