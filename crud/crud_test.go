package crud

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"noble-group-services/core"
	"noble-group-services/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}
	if core.DB == nil {
		if err := core.InitDB(dsn); err != nil {
			t.Fatalf("Failed to initialize database: %v", err)
		}
		SetDB(core.DB)
	}
}

// ================== Categories Unit Tests ==================

func TestCategoriesHandler_Get(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/categories", nil)
	w := httptest.NewRecorder()

	CategoriesHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var categories []models.Category
	err := json.NewDecoder(w.Body).Decode(&categories)
	assert.NoError(t, err)
	assert.NotNil(t, categories)
}

func TestCategoriesHandler_Post_ValidData(t *testing.T) {
	setupTestDB(t)

	category := models.Category{
		Name: "Test Category",
		Slug: "test-category-unit",
	}
	body, _ := json.Marshal(category)

	req := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CategoriesHandler(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Response body: %s", w.Body.String())

	var created models.Category
	err := json.NewDecoder(w.Body).Decode(&created)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Test Category", created.Name)

	// Cleanup
	_, _ = db.Exec("DELETE FROM categories WHERE id = $1", created.ID)
}

func TestCategoriesHandler_Post_InvalidData(t *testing.T) {
	setupTestDB(t)

	// Missing required fields
	category := models.Category{
		Name: "",
		Slug: "",
	}
	body, _ := json.Marshal(category)

	req := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CategoriesHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoriesHandler_MethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodPut, "/products/categories", nil)
	w := httptest.NewRecorder()

	CategoriesHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestCategoryItemHandler_Get(t *testing.T) {
	setupTestDB(t)

	// First create a category
	category := models.Category{
		Name: "Test Get Category",
		Slug: "test-get-category",
	}
	body, _ := json.Marshal(category)
	createReq := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	CategoriesHandler(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var created models.Category
	json.NewDecoder(createW.Body).Decode(&created)

	// Test Get
	req := httptest.NewRequest(http.MethodGet, "/products/categories/"+created.ID, nil)
	w := httptest.NewRecorder()
	CategoryItemHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetched models.Category
	json.NewDecoder(w.Body).Decode(&fetched)
	assert.Equal(t, created.ID, fetched.ID)

	// Cleanup
	_, _ = db.Exec("DELETE FROM categories WHERE id = $1", created.ID)
}

func TestCategoryItemHandler_GetNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/categories/non-existent-id", nil)
	w := httptest.NewRecorder()
	CategoryItemHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoryItemHandler_Update(t *testing.T) {
	setupTestDB(t)

	// Create
	category := models.Category{Name: "Before Update", Slug: "before-update"}
	body, _ := json.Marshal(category)
	createReq := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	CategoriesHandler(createW, createReq)

	var created models.Category
	json.NewDecoder(createW.Body).Decode(&created)

	// Update
	created.Name = "After Update"
	updateBody, _ := json.Marshal(created)
	updateReq := httptest.NewRequest(http.MethodPut, "/products/categories/"+created.ID, bytes.NewBuffer(updateBody))
	updateW := httptest.NewRecorder()
	CategoryItemHandler(updateW, updateReq)

	assert.Equal(t, http.StatusOK, updateW.Code)

	var updated models.Category
	json.NewDecoder(updateW.Body).Decode(&updated)
	assert.Equal(t, "After Update", updated.Name)

	// Cleanup
	_, _ = db.Exec("DELETE FROM categories WHERE id = $1", created.ID)
}

func TestCategoryItemHandler_Delete(t *testing.T) {
	setupTestDB(t)

	// Create
	category := models.Category{Name: "To Delete", Slug: "to-delete"}
	body, _ := json.Marshal(category)
	createReq := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	CategoriesHandler(createW, createReq)

	var created models.Category
	json.NewDecoder(createW.Body).Decode(&created)

	// Delete
	deleteReq := httptest.NewRequest(http.MethodDelete, "/products/categories/"+created.ID, nil)
	deleteW := httptest.NewRecorder()
	CategoryItemHandler(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)

	// Verify deleted
	getReq := httptest.NewRequest(http.MethodGet, "/products/categories/"+created.ID, nil)
	getW := httptest.NewRecorder()
	CategoryItemHandler(getW, getReq)
	assert.Equal(t, http.StatusNotFound, getW.Code)
}

// ================== Manufacturers Unit Tests ==================

func TestManufacturersHandler_Get(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/manufacturers", nil)
	w := httptest.NewRecorder()

	ManufacturersHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var manufacturers []models.Manufacturer
	err := json.NewDecoder(w.Body).Decode(&manufacturers)
	assert.NoError(t, err)
}

func TestManufacturersHandler_Post_ValidData(t *testing.T) {
	setupTestDB(t)

	manufacturer := models.Manufacturer{
		Name: "Test Manufacturer",
		Slug: "test-manufacturer-unit",
	}
	body, _ := json.Marshal(manufacturer)

	req := httptest.NewRequest(http.MethodPost, "/products/manufacturers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ManufacturersHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created models.Manufacturer
	json.NewDecoder(w.Body).Decode(&created)
	assert.NotEmpty(t, created.ID)

	// Cleanup
	_, _ = db.Exec("DELETE FROM manufacturers WHERE id = $1", created.ID)
}

func TestManufacturersHandler_Post_InvalidData(t *testing.T) {
	setupTestDB(t)

	manufacturer := models.Manufacturer{Name: "", Slug: ""}
	body, _ := json.Marshal(manufacturer)

	req := httptest.NewRequest(http.MethodPost, "/products/manufacturers", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ManufacturersHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestManufacturerItemHandler_CRUD(t *testing.T) {
	setupTestDB(t)

	// Create
	manufacturer := models.Manufacturer{Name: "CRUD Test", Slug: "crud-test-mfr"}
	body, _ := json.Marshal(manufacturer)
	createReq := httptest.NewRequest(http.MethodPost, "/products/manufacturers", bytes.NewBuffer(body))
	createW := httptest.NewRecorder()
	ManufacturersHandler(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var created models.Manufacturer
	json.NewDecoder(createW.Body).Decode(&created)

	// Get
	getReq := httptest.NewRequest(http.MethodGet, "/products/manufacturers/"+created.ID, nil)
	getW := httptest.NewRecorder()
	ManufacturerItemHandler(getW, getReq)
	assert.Equal(t, http.StatusOK, getW.Code)

	// Update
	created.Name = "CRUD Test Updated"
	updateBody, _ := json.Marshal(created)
	updateReq := httptest.NewRequest(http.MethodPut, "/products/manufacturers/"+created.ID, bytes.NewBuffer(updateBody))
	updateW := httptest.NewRecorder()
	ManufacturerItemHandler(updateW, updateReq)
	assert.Equal(t, http.StatusOK, updateW.Code)

	// Delete
	deleteReq := httptest.NewRequest(http.MethodDelete, "/products/manufacturers/"+created.ID, nil)
	deleteW := httptest.NewRecorder()
	ManufacturerItemHandler(deleteW, deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteW.Code)
}

// ================== Products Unit Tests ==================

func TestProductsHandler_Get(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	ProductsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	err := json.NewDecoder(w.Body).Decode(&products)
	assert.NoError(t, err)
}

func TestProductsHandler_GetWithFilters(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		queryParams string
	}{
		{"with category filter", "?category=smartfony"},
		{"with manufacturer filter", "?manufacturer=apple"},
		{"with search", "?search=iphone"},
		{"with inStockOnly", "?inStockOnly=true"},
		{"with pagination", "?page=1&limit=10"},
		{"with all filters", "?category=smartfony&manufacturer=apple&search=iphone&inStockOnly=true&page=1&limit=5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			ProductsHandler(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestProductsHandler_Post_ValidData(t *testing.T) {
	setupTestDB(t)

	// Get existing manufacturer and category
	var m models.Manufacturer
	err := db.Get(&m, "SELECT id FROM manufacturers LIMIT 1")
	require.NoError(t, err)

	var c models.Category
	err = db.Get(&c, "SELECT id FROM categories LIMIT 1")
	require.NoError(t, err)

	product := models.Product{
		Name:           "Test Product Unit",
		Slug:           "test-product-unit",
		ManufacturerID: m.ID,
		CategoryID:     c.ID,
		Price:          10000,
		Stock:          5,
		Availability:   "in_stock",
		SKU:            "TEST-UNIT-SKU",
	}
	body, _ := json.Marshal(product)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ProductsHandler(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())

	var created models.Product
	json.NewDecoder(w.Body).Decode(&created)
	assert.NotEmpty(t, created.ID)

	// Cleanup
	_, _ = db.Exec("DELETE FROM products WHERE id = $1", created.ID)
}

func TestProductsHandler_Post_MissingRequiredFields(t *testing.T) {
	setupTestDB(t)

	product := models.Product{Name: "Incomplete Product"}
	body, _ := json.Marshal(product)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	ProductsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestProductItemHandler_Get(t *testing.T) {
	setupTestDB(t)

	// Get existing product
	var p models.Product
	err := db.Get(&p, "SELECT id FROM products LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	req := httptest.NewRequest(http.MethodGet, "/products/"+p.ID, nil)
	w := httptest.NewRecorder()

	ProductItemHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fetched models.Product
	json.NewDecoder(w.Body).Decode(&fetched)
	assert.Equal(t, p.ID, fetched.ID)
}

func TestProductItemHandler_GetNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/products/non-existent-product-id", nil)
	w := httptest.NewRecorder()

	ProductItemHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
