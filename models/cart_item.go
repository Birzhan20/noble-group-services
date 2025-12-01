package models

// CartItem represents an item in the shopping cart.
type CartItem struct {
	Product
	Quantity int `json:"quantity"`
}
