package models

type CartItem struct {
	Product
	Quantity int `json:"quantity"`
}
