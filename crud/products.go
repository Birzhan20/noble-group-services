package crud

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"noble-group-services/models"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB // будет проинициализировано в main.go

// ProductsHandler handles GET /products and POST /products
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetProducts(w, r)
	case http.MethodPost:
		CreateProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ProductItemHandler handles GET, PUT, DELETE /products/{id}
func ProductItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetProduct(w, r)
	case http.MethodPut:
		UpdateProduct(w, r)
	case http.MethodDelete:
		DeleteProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetProducts godoc
// @Summary Get list of products
// @Description Get a list of products with optional filtering
// @Tags products
// @Produce json
// @Param category query string false "Category Slug"
// @Param manufacturer query string false "Manufacturer Slug"
// @Param search query string false "Search term"
// @Param inStockOnly query bool false "Only in stock"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {array} models.Product
// @Router /products [get]
func GetProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	categorySlug := query.Get("category")
	manufacturerSlug := query.Get("manufacturer")
	search := strings.ToLower(query.Get("search"))
	inStockOnly := query.Get("inStockOnly") == "true"

	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var products []models.Product

	q := `
		SELECT 
			p.id, p.name, p.slug, p.price, p.old_price, p.description, p.features, p.image, 
			p.stock, p.rating, p.reviews_count, p.sku, p.availability,
			m.id AS "manufacturer.id", m.name AS "manufacturer.name", m.slug AS "manufacturer.slug", m.logo AS "manufacturer.logo",
			c.id AS "category.id", c.name AS "category.name", c.slug AS "category.slug"
		FROM products p
		LEFT JOIN manufacturers m ON p.manufacturer_id = m.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE true
	`

	args := []interface{}{}
	argID := 1

	if categorySlug != "" {
		q += ` AND c.slug = $` + strconv.Itoa(argID)
		args = append(args, categorySlug)
		argID++
	}
	if manufacturerSlug != "" {
		q += ` AND m.slug = $` + strconv.Itoa(argID)
		args = append(args, manufacturerSlug)
		argID++
	}
	if search != "" {
		q += ` AND LOWER(p.name) LIKE $` + strconv.Itoa(argID)
		args = append(args, "%"+search+"%")
		argID++
	}
	if inStockOnly {
		q += ` AND p.stock > 0 AND p.availability = 'in_stock'`
	}

	q += ` ORDER BY p.name LIMIT $` + strconv.Itoa(argID) + ` OFFSET $` + strconv.Itoa(argID+1)
	args = append(args, limit, offset)

	err := db.Select(&products, q, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// CreateProduct godoc
// @Summary Create a product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body models.Product true "Product"
// @Success 201 {object} models.Product
// @Failure 400 {string} string "Invalid request"
// @Router /products [post]
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if p.Name == "" || p.Slug == "" || p.ManufacturerID == "" || p.CategoryID == "" {
		http.Error(w, "Name, Slug, ManufacturerID, and CategoryID are required", http.StatusBadRequest)
		return
	}

	p.ID = uuid.New().String()
	if p.Features == nil {
		p.Features = models.JSONStringArray{}
	}
	if p.Image == nil {
		p.Image = models.JSONStringArray{}
	}

	_, err := db.Exec(`
		INSERT INTO products (
			id, name, slug, manufacturer_id, category_id, price, old_price, 
			description, features, image, stock, rating, reviews_count, sku, availability
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, p.ID, p.Name, p.Slug, p.ManufacturerID, p.CategoryID, p.Price, p.OldPrice,
		p.Description, p.Features, p.Image, p.Stock, p.Rating, p.ReviewsCount, p.SKU, p.Availability)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get details of a specific product
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [get]
func GetProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	var product models.Product
	err := db.Get(&product, `
		SELECT 
			p.*, 
			m.id AS "manufacturer.id", m.name AS "manufacturer.name", m.slug AS "manufacturer.slug", m.logo AS "manufacturer.logo",
			c.id AS "category.id", c.name AS "category.name", c.slug AS "category.slug"
		FROM products p
		LEFT JOIN manufacturers m ON p.manufacturer_id = m.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`, id)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body models.Product true "Product"
// @Success 200 {object} models.Product
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [put]
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	p.ID = id
	if p.Features == nil {
		p.Features = models.JSONStringArray{}
	}
	if p.Image == nil {
		p.Image = models.JSONStringArray{}
	}

	_, err := db.Exec(`
		UPDATE products SET 
			name=$1, slug=$2, manufacturer_id=$3, category_id=$4, price=$5, old_price=$6, 
			description=$7, features=$8, image=$9, stock=$10, rating=$11, reviews_count=$12, 
			sku=$13, availability=$14
		WHERE id=$15
	`, p.Name, p.Slug, p.ManufacturerID, p.CategoryID, p.Price, p.OldPrice,
		p.Description, p.Features, p.Image, p.Stock, p.Rating, p.ReviewsCount, p.SKU, p.Availability, p.ID)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {string} string "Product not found"
// @Router /products/{id} [delete]
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" {
		http.Error(w, "ID required", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`DELETE FROM products WHERE id = $1`, id)
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
