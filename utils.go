package main

import (
	"fmt"
	"strings"
)

func getInsertQuery(fields []string) string {
	joinedFields := strings.Join(fields, ",")
	placeholders := strings.Repeat("?,", len(fields)-1) + "?"
	return fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", joinedFields, placeholders)
}
