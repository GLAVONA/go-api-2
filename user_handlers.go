package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type user struct {
	Id       int       `json:"id"`
	Name     string    `json:"name"`
	Age      int       `json:"age"`
	CreateAt time.Time `json:"createdAt"`
}

// type account struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithJSON(w, http.StatusOK, "No such user")
	}

	row := server.db.QueryRow(`SELECT * FROM users WHERE id = ?`, id)

	user := user{}
	row.Scan(&user.Id, &user.Name, &user.Age, &user.CreateAt)

	if user.Id == 0 {
		respondWithJSON(w, http.StatusNotFound, "user not found")

		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := server.db.Query("SELECT * FROM users")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	defer rows.Close()

	var users []user

	for rows.Next() {
		user := user{}
		err = rows.Scan(&user.Id, &user.Name, &user.Age, &user.CreateAt)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}
		users = append(users, user)
	}

	respondWithJSON(w, http.StatusOK, users)
}

func createUser(w http.ResponseWriter, r *http.Request) {

	user := user{}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)

		return
	}

	res, err := server.db.Exec("INSERT INTO users (name,age,created_at) VALUES(?,?,?)", user.Name, user.Age, time.Now())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Failed to insert user: %v", err)
		return
	}
	id, _ := res.LastInsertId()
	user.Id = int(id)

	respondWithJSON(w, http.StatusOK, user)
}
