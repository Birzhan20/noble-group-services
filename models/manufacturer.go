package models

type Manufacturer struct {
	ID   string  `db:"id" json:"id"`
	Name string  `db:"name" json:"name"`
	Slug string  `db:"slug" json:"slug"`
	Logo *string `db:"logo" json:"logo,omitempty"`
}
