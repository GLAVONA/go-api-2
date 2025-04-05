package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const csrfTokenCtxKey contextKey = "csrfToken"

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
		var csrfToken string

		query := "SELECT csrf_token FROM logins WHERE session_token = ?"
		err = server.db.QueryRow(query, sessionToken).Scan(&csrfToken)

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

		ctx := context.WithValue(r.Context(), csrfTokenCtxKey, csrfToken)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}

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

		csrfTokenFromSession, ok := r.Context().Value(csrfTokenCtxKey).(string)
		if !ok {
			respondWithJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "server_error",
				Message: "Could not retrieve csrfToken from context",
			})
			log.Println("Error: CSRF Token no found in context of CSRF Protection Middleware")
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

		next.ServeHTTP(w, r)
	})
}
