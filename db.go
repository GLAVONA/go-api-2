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
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    name TEXT NOT NULL,
    age INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL`)

	server.db = db
}

func initTable(db *sql.DB, name string, schema string) {
	_, err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v (%v)`, name, schema))

	if err != nil {
		fmt.Println("Couldn't create table: ", err)
	}
}
