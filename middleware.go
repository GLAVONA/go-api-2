package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

var middlewares = []func(http.Handler) http.Handler{
	LoggingMiddleware,
}

func initMiddleware() {
	for _, m := range middlewares {
		server.router.Use(m)
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v %v\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const userIDKey contextKey = "userID"

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{
					Error:   "unauthorized",
					Message: "No session cookie provided",
				})
				log.Println("Authentication failed: No session cookie")
				return
			}
			respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "bad_request",
				Message: "Error reading session cookie",
			})
			log.Printf("Authentication failed: Error reading cookie: %v", err)
			return
		}

		sessionToken := sessionCookie.Value
		var userID string

		query := "SELECT user_id FROM logins WHERE session_token = ?"
		err = server.db.QueryRow(query, sessionToken).Scan(&userID)

		if err != nil {
			if err == sql.ErrNoRows {
				respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{
					Error:   "unauthorized",
					Message: "Invalid or expired session token",
				})
				log.Printf("Authentication failed: Invalid session token: %s", sessionToken)
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				return
			}
			respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to validate session",
			})
			log.Printf("Authentication failed: Database error validating session: %v", err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

		// Get the CSRF token from the request header or form
		csrfTokenFromRequest := r.Header.Get("X-CSRF-Token")
		if csrfTokenFromRequest == "" && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete) {
			if err := r.ParseForm(); err != nil {
				respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
					Error:   "bad_request",
					Message: "Failed to parse form",
				})
				log.Printf("CSRF check failed: Failed to parse form: %v", err)
				return
			}
			csrfTokenFromRequest = r.FormValue("csrf_token")
		}

		if csrfTokenFromRequest == "" {
			respondWithJSON(w, http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "CSRF token is missing",
			})
			log.Println("CSRF check failed: CSRF token is missing")
			return
		}

		sessionCookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{
					Error:   "unauthorized",
					Message: "No session cookie provided for CSRF check",
				})
				log.Println("CSRF check failed: No session cookie")
				return
			}
			respondWithJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "bad_request",
				Message: "Error reading session cookie for CSRF check",
			})
			log.Printf("CSRF check failed: Error reading session cookie: %v", err)
			return
		}
		sessionToken := sessionCookie.Value

		var csrfTokenFromSession string
		query := "SELECT csrf_token FROM logins WHERE session_token = ?"
		err = server.db.QueryRow(query, sessionToken).Scan(&csrfTokenFromSession)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithJSON(w, http.StatusUnauthorized, ErrorResponse{
					Error:   "unauthorized",
					Message: "Invalid or expired session token for CSRF check",
				})
				log.Printf("CSRF check failed: Invalid session token: %s", sessionToken)
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				return
			}
			respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Failed to validate CSRF token",
			})
			log.Printf("CSRF check failed: Database error validating CSRF token: %v", err)
			return
		}

		if csrfTokenFromRequest != csrfTokenFromSession {
			respondWithJSON(w, http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Message: "Invalid CSRF token",
			})
			log.Println("CSRF check failed: Invalid CSRF token")
			return
		}

		log.Println("CSRF PASSED")
		next.ServeHTTP(w, r)
	})
}
