package crud

import (
	"encoding/json"
	"net/http"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
)

// ManufacturersHandler handles requests for product manufacturers.
// @Summary Get a list of product manufacturers
// @Description Get a list of all available product manufacturers
// @Tags products
// @Produce  json
// @Success 200 {object} map[string][]string
// @Router /products/manufacturers [get]
func ManufacturersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := core.DB.Query("SELECT DISTINCT manufacturer FROM products ORDER BY manufacturer")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	mans := []string{}
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mans = append(mans, m)
	}

	response := map[string][]string{
		"manufacturers": mans,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
