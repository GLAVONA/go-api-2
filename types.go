package main

import (
	"sync"

	"golang.org/x/time/rate"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type user struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type userResponse struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type Client struct {
	limiter *rate.Limiter
}

type Clients struct {
	cMap map[string]*Client
	mu   sync.Mutex
}
