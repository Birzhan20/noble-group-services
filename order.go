package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Order represents a customer order.
type Order struct {
	ID         string `json:"id"`
	OrderNumber string `json:"orderNumber"`
	Total      int    `json:"total"`
}

// generateOrderNumber creates a new order number.
func generateOrderNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), rand.Intn(100000))
}
