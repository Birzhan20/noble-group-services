package models

// Order represents a customer order.
type Order struct {
	ID          string `json:"id"`
	OrderNumber string `json:"orderNumber"`
	Total       int    `json:"total"`
}
