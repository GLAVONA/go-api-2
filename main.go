package main

import (
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

var server *APIServer

var clients = Clients{
	cMap: make(map[string]*Client),
}

func main() {

	server = NewAPIServer(":8080")

	initDb("db")

	initRoutes()

	cleanUpClients(&clients)

	fmt.Println("Listening on: ", server.addr)
	log.Fatal(http.ListenAndServe(server.addr, server.router))

}
