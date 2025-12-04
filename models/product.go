package models

type Product struct {
	ID             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	Slug           string `db:"slug" json:"slug"`
	ManufacturerID string `db:"manufacturer_id" json:"manufacturerId"`
	CategoryID     string `db:"category_id" json:"categoryId"`

	Manufacturer Manufacturer `db:"manufacturer" json:"manufacturer"`
	Category     Category     `db:"category" json:"category"`

	Price        int             `db:"price" json:"price"`
	OldPrice     *int            `db:"old_price" json:"oldPrice,omitempty"`
	Description  string          `db:"description" json:"description"`
	Features     JSONStringArray `db:"features" json:"features"`
	Image        JSONStringArray `db:"image" json:"image"`
	Stock        int             `db:"stock" json:"stock"`
	Rating       float64         `db:"rating" json:"rating"`
	ReviewsCount int             `db:"reviews_count" json:"reviews"`
	SKU          string          `db:"sku" json:"sku"`
	Availability string          `db:"availability" json:"availability"`
}
