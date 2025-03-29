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

	createTable(db, "users", `id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))`)

	createTrigger(db, "update_timestamp", `BEFORE UPDATE ON users 
	FOR EACH ROW 
	BEGIN
		UPDATE users 
		SET updated_at = datetime('now') 
		WHERE id = OLD.id;
	END;
`)

	server.db = db
}

func createTrigger(db *sql.DB, name, schema string) {
	_, err := db.Exec(fmt.Sprintf(`CREATE TRIGGER IF NOT EXISTS %v %v`, name, schema))
	if err != nil {
		fmt.Println("Couldn't create trigger: ", err)
	}
}

func createTable(db *sql.DB, name string, schema string) {
	_, err := db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %v (%v)`, name, schema))

	if err != nil {
		fmt.Println("Couldn't create table: ", err)
	}
}
