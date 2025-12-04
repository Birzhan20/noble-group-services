package crud

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"noble-group-services/models"
)

// CategoriesHandler handles GET /categories and POST /categories
func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCategories(w, r)
	case http.MethodPost:
		CreateCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CategoryItemHandler handles GET, PUT, DELETE /categories/{id}
func CategoryItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCategory(w, r)
	case http.MethodPut:
		UpdateCategory(w, r)
	case http.MethodDelete:
		DeleteCategory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetCategories godoc
// @Summary Get all categories
// @Description Get a list of all product categories
// @Tags categories
// @Produce json
// @Success 200 {array} models.Category
// @Router /products/categories [get]
func GetCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	err := db.Select(&categories, `SELECT id, name, slug, parent_id AS "parent_id", image FROM categories ORDER BY name`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CreateCategory godoc
// @Summary Create a category
// @Description Create a new product category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body models.Category true "Category"
// @Success 201 {object} models.Category
// @Failure 400 {string} string "Invalid request"
// @Router /products/categories [post]
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var c models.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if c.Name == "" || c.Slug == "" {
		http.Error(w, "Name and Slug are required", http.StatusBadRequest)
		return
	}

	c.ID = uuid.New().String()

	_, err := db.Exec(`INSERT INTO categories (id, name, slug, parent_id, image) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, c.Name, c.Slug, c.ParentID, c.Image)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get details of a specific category
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} models.Category
// @Failure 404 {string} string "Category not found"
// @Router /products/categories/{id} [get]
func GetCategory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/categories/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var c models.Category
	err := db.Get(&c, `SELECT id, name, slug, parent_id AS "parent_id", image FROM categories WHERE id = $1`, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body models.Category true "Category"
// @Success 200 {object} models.Category
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Category not found"
// @Router /products/categories/{id} [put]
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/categories/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var c models.Category
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	c.ID = id

	_, err := db.Exec(`UPDATE categories SET name = $1, slug = $2, parent_id = $3, image = $4 WHERE id = $5`,
		c.Name, c.Slug, c.ParentID, c.Image, c.ID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {string} string "Category not found"
// @Router /products/categories/{id} [delete]
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/categories/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`DELETE FROM categories WHERE id = $1`, id)
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
