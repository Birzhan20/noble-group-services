package models

type Category struct {
	ID       string  `db:"id" json:"id"`
	Name     string  `db:"name" json:"name"`
	Slug     string  `db:"slug" json:"slug"`
	ParentID *string `db:"parent_id" json:"parentId,omitempty"`
	Image    *string `db:"image" json:"image,omitempty"`
}
