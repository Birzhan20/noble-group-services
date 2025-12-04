package crud

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"noble-group-services/models"
)

// ManufacturersHandler handles GET /manufacturers and POST /manufacturers
func ManufacturersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetManufacturers(w, r)
	case http.MethodPost:
		CreateManufacturer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ManufacturerItemHandler handles GET, PUT, DELETE /manufacturers/{id}
func ManufacturerItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetManufacturer(w, r)
	case http.MethodPut:
		UpdateManufacturer(w, r)
	case http.MethodDelete:
		DeleteManufacturer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetManufacturers godoc
// @Summary Get all manufacturers
// @Description Get a list of all manufacturers
// @Tags manufacturers
// @Produce json
// @Success 200 {array} models.Manufacturer
// @Router /products/manufacturers [get]
func GetManufacturers(w http.ResponseWriter, r *http.Request) {
	var manufacturers []models.Manufacturer
	err := db.Select(&manufacturers, `SELECT id, name, slug, logo FROM manufacturers ORDER BY name`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manufacturers)
}

// CreateManufacturer godoc
// @Summary Create a manufacturer
// @Description Create a new manufacturer
// @Tags manufacturers
// @Accept json
// @Produce json
// @Param manufacturer body models.Manufacturer true "Manufacturer"
// @Success 201 {object} models.Manufacturer
// @Failure 400 {string} string "Invalid request"
// @Router /products/manufacturers [post]
func CreateManufacturer(w http.ResponseWriter, r *http.Request) {
	var m models.Manufacturer
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.Name == "" || m.Slug == "" {
		http.Error(w, "Name and Slug are required", http.StatusBadRequest)
		return
	}

	m.ID = uuid.New().String()

	_, err := db.Exec(`INSERT INTO manufacturers (id, name, slug, logo) VALUES ($1, $2, $3, $4)`,
		m.ID, m.Name, m.Slug, m.Logo)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// GetManufacturer godoc
// @Summary Get manufacturer by ID
// @Description Get details of a specific manufacturer
// @Tags manufacturers
// @Produce json
// @Param id path string true "Manufacturer ID"
// @Success 200 {object} models.Manufacturer
// @Failure 404 {string} string "Manufacturer not found"
// @Router /products/manufacturers/{id} [get]
func GetManufacturer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/manufacturers/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var m models.Manufacturer
	err := db.Get(&m, `SELECT id, name, slug, logo FROM manufacturers WHERE id = $1`, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// UpdateManufacturer godoc
// @Summary Update manufacturer
// @Description Update an existing manufacturer
// @Tags manufacturers
// @Accept json
// @Produce json
// @Param id path string true "Manufacturer ID"
// @Param manufacturer body models.Manufacturer true "Manufacturer"
// @Success 200 {object} models.Manufacturer
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Manufacturer not found"
// @Router /products/manufacturers/{id} [put]
func UpdateManufacturer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/manufacturers/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var m models.Manufacturer
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	m.ID = id // Ensure ID matches path

	_, err := db.Exec(`UPDATE manufacturers SET name = $1, slug = $2, logo = $3 WHERE id = $4`,
		m.Name, m.Slug, m.Logo, m.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// DeleteManufacturer godoc
// @Summary Delete manufacturer
// @Description Delete a manufacturer
// @Tags manufacturers
// @Produce json
// @Param id path string true "Manufacturer ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {string} string "Manufacturer not found"
// @Router /products/manufacturers/{id} [delete]
func DeleteManufacturer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/manufacturers/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`DELETE FROM manufacturers WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
