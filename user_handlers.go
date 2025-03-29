package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := server.db.Query("SELECT id, username FROM users")
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
		if err := rows.Scan(&u.Id, &u.Username); err != nil {
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

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Username is required",
		})
		return
	}

	var u userResponse
	err := server.db.QueryRow("SELECT id, username FROM users WHERE username = ?", username).
		Scan(&u.Id, &u.Username)
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
		return
	}

	if u.Username == "" || u.Password == "" {
		respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_input",
			Message: "Username and password are required",
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
	if exists {
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

	query := getInsertQuery([]string{"id", "username", "password"})

	u.Id = uuid.New().String()

	_, err = server.db.Exec(query, u.Id, u.Username, hashedPassword)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to create user",
		})
		log.Printf("Failed to insert user: %v", err)
		return
	}

	respondWithJSON(w, http.StatusOK, ToUserResponse(u))
}

func doesUserExist(u *user) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE username = ?`
	err := server.db.QueryRow(query, u.Username).Scan(&count)
	return count > 0, err
}

func ToUserResponse(u user) userResponse {
	return userResponse{
		Id:       u.Id,
		Username: u.Username,
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func isPasswordMatch(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))

	return err == nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	loginCredentials := user{}
	err := json.NewDecoder(r.Body).Decode(&loginCredentials)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{Message: "Failed to decode user"})
		log.Println("Failed to decode user")
		return
	}

	userFromDb, err := getUserFromDb(&loginCredentials)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{Message: "Username or password does not exist"})
		log.Printf("Failed to get user from Db -- %#v -- %v", loginCredentials, err)

		return
	}

	if !isPasswordMatch(loginCredentials.Password, []byte(userFromDb.Password)) {
		respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{Message: "Username or password does not exist"})
		log.Printf("Wrong password -- %#v", loginCredentials)

		return
	}

}

func getUserFromDb(u *user) (user, error) {
	res := server.db.QueryRow("SELECT * FROM users WHERE username = ?", u.Username)
	dbUser := user{}
	err := res.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.CreatedAt, &dbUser.UpdatedAt)

	return dbUser, err
}
