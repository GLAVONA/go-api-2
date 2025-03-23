package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type user struct {
	Id       int       `json:"id"`
	Name     string    `json:"name"`
	Age      int       `json:"age"`
	CreateAt time.Time `json:"createdAt"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}

func getUser(w http.ResponseWriter, r *http.Request) {

}

func getUsers(w http.ResponseWriter, r *http.Request) {
}

func createUser(w http.ResponseWriter, r *http.Request) {

	user := user{}

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)

		return
	}

	res, dbErr := server.db.Exec("INSERT INTO users (name,age,created_at) VALUES(?,?,?)", user.Name, user.Age, time.Now())
	if dbErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Failed to insert user: %v", dbErr)
		return
	}
	id, _ := res.LastInsertId()
	user.Id = int(id)

	respondWithJSON(w, http.StatusOK, user)
}
