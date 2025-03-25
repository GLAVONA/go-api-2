package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type user struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type userResponse struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := server.db.Query("SELECT id, username, name, age FROM users")
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve users",
		})
		log.Printf("Failed to query users: %v", err)
		return
	}
	defer rows.Close()

	var users []userResponse
	for rows.Next() {
		var u userResponse
		if err := rows.Scan(&u.Id, &u.Username, &u.Name, &u.Age); err != nil {
			respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to scan user data",
			})
			log.Printf("Error scanning row: %v", err)
			return
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Error processing user data",
		})
		log.Printf("Error iterating rows: %v", err)
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	var u userResponse
	err := server.db.QueryRow("SELECT id, username, name, age FROM users WHERE id = ?", id).
		Scan(&u.Id, &u.Username, &u.Name, &u.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w, http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
		} else {
			respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to retrieve user",
			})
			log.Printf("Failed to scan user: %v", err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
		return
	}

	if u.Username == "" || u.Name == "" {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_input",
			Message: "Username and name are required",
		})
		return
	}

	exists, err := doesUserExist(&u)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to check username availability",
		})
		log.Printf("Failed to check username: %v", err)
		return
	}
	if exists > 0 {
		respondWithJSON(w, http.StatusConflict, ErrorResponse{
			Error:   "username_taken",
			Message: "The username already exists",
		})
		return
	}

	hashedPassword, err := HashPassword(u.Password)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to hash password",
		})
		log.Printf("Failed to hash password: %v", err)
		return
	}

	timeNow := time.Now()
	query := `
        INSERT INTO users (username, password, name, age, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	res, err := server.db.Exec(query, u.Username, hashedPassword, u.Name, u.Age, timeNow.Format(time.RFC3339), timeNow.Format(time.RFC3339))
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
		log.Printf("Failed to insert user: %v", err)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to retrieve new user ID",
		})
		log.Printf("Failed to get last insert ID: %v", err)
		return
	}
	u.Id = int(id)

	respondWithJSON(w, http.StatusOK, ToUserResponse(u))
}

func doesUserExist(u *user) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = ?`
	err := server.db.QueryRow(query, u.Username).Scan(&count)
	return count, err
}

func ToUserResponse(u user) userResponse {
	return userResponse{
		Id:       u.Id,
		Username: u.Username,
		Name:     u.Name,
		Age:      u.Age,
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
