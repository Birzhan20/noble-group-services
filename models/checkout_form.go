package models

type CartItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CheckoutForm struct {
	CustomerType string            `json:"customerType"`
	Name         string            `json:"name"`
	Phone        string            `json:"phone"`
	Email        string            `json:"email"`
	CompanyName  *string           `json:"companyName,omitempty"`
	BIN          *string           `json:"bin,omitempty"`
	Address      string            `json:"address"`
	Comment      *string           `json:"comment,omitempty"`
	Company      bool              `json:"company"` // New field
	Carts        []CartItemRequest `json:"carts"`   // New field
}
