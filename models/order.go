package models

// Order represents a placed order.
type Order struct {
	ID          string `json:"id"`
	OrderNumber string `json:"orderNumber"`
	Total       int    `json:"total"`
}
