package main

import (
	"database/sql"
	"fmt"
	"log"
)

func initDb(dbName string) {
	db, err := sql.Open("sqlite", fmt.Sprint("./", dbName))

	if err != nil {
		log.Fatal("Couldn't connect to database")
	}

	initTable(db, "users", `id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	age INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL`)

	server.db = db
}

func initTable(db *sql.DB, name string, schema string) {
	_, err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v (%v)`, name, schema))

	if err != nil {
		fmt.Println("Couldn't create table: ", err)
	}
}
