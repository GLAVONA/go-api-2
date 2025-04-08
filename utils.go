package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

func getInsertQuery(table string, fields []string) string {
	joinedFields := strings.Join(fields, ",")
	placeholders := strings.Repeat("?,", len(fields)-1) + "?"
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, joinedFields, placeholders)
}

func getUserFromDb(u *user) (user, error) {
	res := server.db.QueryRow("SELECT * FROM users WHERE username = ?", u.Username)
	dbUser := user{}
	err := res.Scan(&dbUser.Id, &dbUser.Username, &dbUser.Password, &dbUser.CreatedAt, &dbUser.UpdatedAt)

	return dbUser, err
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

func generateToken(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func getClientLimiter(ip string, clients *Clients) *rate.Limiter {
	clients.mu.Lock()
	defer clients.mu.Unlock()

	if client, exists := clients.cMap[ip]; exists {

		return client.limiter
	}

	limiter := rate.NewLimiter(1, 2)
	clients.cMap[ip] = &Client{limiter: limiter}
	return limiter
}
