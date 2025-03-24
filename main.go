package main

import (
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

var server *APIServer

func main() {

	server = NewAPIServer(":8080")

	initDb("db")

	initMiddleware()

	initRoutes()

	fmt.Println("Listening on: ", server.addr)
	log.Fatal(http.ListenAndServe(server.addr, server.router))

}
