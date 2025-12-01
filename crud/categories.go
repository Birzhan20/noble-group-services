package crud

import (
	"encoding/json"
	"net/http"
)

// CategoriesHandler handles requests for product categories.
// @Summary Get a list of product categories
// @Description Get a list of all available product categories
// @Tags products
// @Produce  json
// @Success 200 {object} map[string][]string
// @Router /products/categories [get]
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories := map[string][]string{
		"categories": {"Все", "Респираторы", "Перчатки", "Каски", "Очки", "Спецодежда"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
