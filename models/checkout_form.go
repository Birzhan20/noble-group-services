package models

// CheckoutForm represents the data submitted by the user during checkout.
type CheckoutForm struct {
	CustomerType string  `json:"customerType"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`
	Email        string  `json:"email"`
	CompanyName  *string `json:"companyName,omitempty"`
	BIN          *string `json:"bin,omitempty"`
	Address      string  `json:"address"`
	Comment      *string `json:"comment,omitempty"`
}
