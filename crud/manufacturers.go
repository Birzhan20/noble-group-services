package crud

import (
	"encoding/json"
	"net/http"
)

// ManufacturersHandler handles requests for product manufacturers.
// @Summary Get a list of product manufacturers
// @Description Get a list of all available product manufacturers
// @Tags products
// @Produce  json
// @Success 200 {object} map[string][]string
// @Router /products/manufacturers [get]
func ManufacturersHandler(w http.ResponseWriter, r *http.Request) {
	manufacturers := map[string][]string{
		"manufacturers": {"3M", "Honeywell", "Ansell", "MSA Safety", "Tyvek", "Uvex"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manufacturers)
}
