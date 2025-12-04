package models

import "time"

type Order struct {
	ID            string    `db:"id" json:"id"`
	OrderNumber   string    `db:"order_number" json:"orderNumber"`
	CustomerName  string    `db:"customer_name" json:"customerName"`
	CustomerPhone string    `db:"customer_phone" json:"customerPhone"`
	CustomerEmail string    `db:"customer_email" json:"customerEmail"`
	Address       string    `db:"address" json:"address"`
	CustomerType  string    `db:"customer_type" json:"customerType"`
	CompanyName   *string   `db:"company_name" json:"companyName,omitempty"`
	BIN           *string   `db:"bin" json:"bin,omitempty"`
	Comment       *string   `db:"comment" json:"comment,omitempty"`
	Total         int       `db:"total" json:"total"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
}

type OrderItem struct {
	ID        string `db:"id" json:"id"`
	OrderID   string `db:"order_id" json:"orderId"`
	ProductID string `db:"product_id" json:"productId"`
	Quantity  int    `db:"quantity" json:"quantity"`
	Price     int    `db:"price" json:"price"`
}
