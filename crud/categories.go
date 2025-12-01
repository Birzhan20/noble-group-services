package crud

import (
	"encoding/json"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
)

// CategoriesHandler handles requests for product categories.
// @Summary Get a list of product categories
// @Description Get a list of all available product categories
// @Tags products
// @Produce  json
// @Success 200 {object} map[string][]string
// @Router /products/categories [get]
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// The frontend expects "Все" (All) to be present, usually as the first item.
	// We will fetch unique categories from DB and prepend "Все".
	rows, err := core.DB.Query("SELECT DISTINCT category FROM products ORDER BY category")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cats := []string{"Все"}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cats = append(cats, c)
	}

	response := map[string][]string{
		"categories": cats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
