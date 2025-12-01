package models

// Product represents a product in the marketplace.
type Product struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Manufacturer string   `json:"manufacturer"`
	Availability string   `json:"availability"`
	Price        int      `json:"price"`
	OldPrice     *int     `json:"oldPrice,omitempty"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	Image        string   `json:"image"`
	Stock        int      `json:"stock"`
	Rating       float64  `json:"rating"`
	Reviews      int      `json:"reviews"`
	SKU          string   `json:"sku"`
}
