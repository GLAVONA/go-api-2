package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	query := getInsertQuery("users", []string{"id", "username", "password"})

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

func logInHandler(w http.ResponseWriter, r *http.Request) {
	loginCredentials := user{}
	err := json.NewDecoder(r.Body).Decode(&loginCredentials)
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{Message: "Failed to decode user"})
		log.Println("Failed to decode user")
		return
	}

	userFromDb, err := getUserFromDb(&loginCredentials)
	if err != nil {
		respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{Message: "Username or password are wrong"})
		log.Printf("Failed to get user from Db -- %#v -- %v", loginCredentials, err)
		return
	}

	if !isPasswordMatch(loginCredentials.Password, []byte(userFromDb.Password)) {
		respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{Message: "Username or password are wrong"})
		log.Printf("Wrong password -- %#v", loginCredentials)
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	loginUuid := uuid.New()

	insertQuery := getInsertQuery("logins", []string{"id", "user_id", "session_token", "csrf_token"})

	_, err = server.db.Exec(insertQuery, loginUuid, userFromDb.Id, sessionToken, csrfToken)

	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{Message: "Something went wrong"})
		log.Printf("Couldn't save session -- %v -- %v", loginCredentials, err)
		return
	}

	respondWithJSON(w, http.StatusOK, "Authenticated!")

}

func logOutHandler(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Message: "No active session found to log out.",
			})
			log.Println("Logout failed: No session cookie found (middleware should have prevented this)")
		} else {
			respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "bad_request",
				Message: "Could not read session cookie.",
			})
			log.Printf("Logout failed: Error reading session cookie: %v", err)
		}
		return
	}

	sessionToken := sessionCookie.Value

	query := "DELETE FROM logins WHERE session_token = ?"
	result, err := server.db.ExecContext(r.Context(), query, sessionToken) // Use ExecContext for cancellation propagation
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Failed to terminate session.",
		})
		log.Printf("Logout failed: Could not delete session token '%s' from database: %v", sessionToken, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Logout warning: Could not verify session deletion for token '%s': %v", sessionToken, err)
	} else if rowsAffected == 0 {
		log.Printf("Logout warning: Session token '%s' was not found in the database during logout.", sessionToken)
	} else {
		log.Printf("Logout successful: Deleted session for token '%s' from database.", sessionToken)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   r.TLS != nil,
	})

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)

	if !ok {
		respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "server_error",
			Message: "Could not retrieve user ID from context",
		})
		log.Println("Error: userID not found in context for protected handler")
		return
	}

	log.Printf("User %s accessed protected route", userID)

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Welcome authenticated user!", "userID": userID})
}
