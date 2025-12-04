package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"noble-group-services/core"
	"noble-group-services/crud"
	"noble-group-services/models"

	"github.com/stretchr/testify/assert"
)

func TestProductRetrieval(t *testing.T) {
	// Database connection string
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}

	// Initialize Database
	if err := core.InitDB(dsn); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer core.CloseDB()

	crud.SetDB(core.DB)

	// Test fetching products
	// We'll use a direct query or call the handler logic if possible,
	// but here we want to test if the model mapping works.
	var products []models.Product
	err := core.DB.Select(&products, `
		SELECT 
			p.id, p.name, p.slug, p.price, p.old_price, p.description, p.features, p.image, 
			p.stock, p.rating, p.reviews_count, p.sku, p.availability,
			m.id AS "manufacturer.id", m.name AS "manufacturer.name", m.slug AS "manufacturer.slug", m.logo AS "manufacturer.logo",
			c.id AS "category.id", c.name AS "category.name", c.slug AS "category.slug"
		FROM products p
		LEFT JOIN manufacturers m ON p.manufacturer_id = m.id
		LEFT JOIN categories c ON p.category_id = c.id
		LIMIT 1
	`)

	if err != nil {
		t.Fatalf("Failed to select products: %v", err)
	}

	if len(products) == 0 {
		t.Log("No products found, skipping validation")
		return
	}

	product := products[0]
	t.Logf("Fetched product: %+v", product)

	assert.NotEmpty(t, product.ID)
	assert.NotEmpty(t, product.Name)
	// Check if Features and Image are populated correctly
	assert.NotEmpty(t, product.Features)
	assert.NotEmpty(t, product.Image)
}

func TestOrderCreation(t *testing.T) {
	// Setup DB
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}
	if err := core.InitDB(dsn); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer core.CloseDB()
	crud.SetDB(core.DB)

	// Create a session ID
	sessionID := "test-session-id"

	// 1. Add item to cart
	// We need a valid product ID. Let's fetch one.
	var product models.Product
	err := core.DB.Get(&product, "SELECT id FROM products LIMIT 1")
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	// Manually add to cart (since CartHandler uses global map, we can access it via HTTP or just trust it works if we test via HTTP)
	// Let's use httptest to call the handler.

	// ... wait, CartHandler uses `carts` map which is package private in `crud`.
	// But `crud.CartHandler` is exported.

	// Construct request to add item
	cartBody := map[string]interface{}{
		"productId": product.ID,
		"quantity":  1,
	}
	cartJSON, _ := json.Marshal(cartBody)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(cartJSON))
	req.Header.Set("X-Session-ID", sessionID)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	crud.CartHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Place order
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	orderBody := models.CheckoutForm{
		Name:         "Test User",
		Phone:        "+77001234567",
		Email:        email,
		Address:      "Test Address 123",
		CustomerType: "individual",
	}
	orderJSON, _ := json.Marshal(orderBody)
	reqOrder := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	reqOrder.Header.Set("X-Session-ID", sessionID)
	reqOrder.Header.Set("Content-Type", "application/json")
	wOrder := httptest.NewRecorder()
	crud.OrdersHandler(wOrder, reqOrder)

	if wOrder.Code != http.StatusCreated {
		t.Logf("Order creation failed: %s", wOrder.Body.String())
	}
	assert.Equal(t, http.StatusCreated, wOrder.Code)

	// Verify order in DB
	var count int
	err = core.DB.Get(&count, "SELECT count(*) FROM orders WHERE customer_email = $1", email)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestOrderCreationWithCarts(t *testing.T) {
	// Setup DB
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}
	if err := core.InitDB(dsn); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer core.CloseDB()
	crud.SetDB(core.DB)

	// Fetch a product
	var product models.Product
	err := core.DB.Get(&product, "SELECT id FROM products LIMIT 1")
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	// Create order with carts field
	email := fmt.Sprintf("test-carts-%d@example.com", time.Now().UnixNano())
	orderBody := models.CheckoutForm{
		Name:         "Test User Carts",
		Phone:        "+77001234567",
		Email:        email,
		Address:      "Test Address Carts",
		CustomerType: "individual",
		Company:      false, // Don't try to send email in test
		Carts: []models.CartItemRequest{
			{ProductID: product.ID, Quantity: 2},
		},
	}
	orderJSON, _ := json.Marshal(orderBody)
	reqOrder := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	reqOrder.Header.Set("X-Session-ID", "dummy-session") // Still required by handler check
	reqOrder.Header.Set("Content-Type", "application/json")
	wOrder := httptest.NewRecorder()
	crud.OrdersHandler(wOrder, reqOrder)

	if wOrder.Code != http.StatusCreated {
		t.Logf("Order creation failed: %s", wOrder.Body.String())
	}
	assert.Equal(t, http.StatusCreated, wOrder.Code)

	// Verify order items count
	var orderID string
	err = core.DB.Get(&orderID, "SELECT id FROM orders WHERE customer_email = $1", email)
	assert.NoError(t, err)

	var itemCount int
	err = core.DB.Get(&itemCount, "SELECT count(*) FROM order_items WHERE order_id = $1", orderID)
	assert.NoError(t, err)
	assert.Equal(t, 1, itemCount)
}

func TestManufacturersCRUD(t *testing.T) {
	setupTestDB(t)

	// 1. Create
	mName := fmt.Sprintf("Test Manufacturer %d", time.Now().UnixNano())
	mSlug := fmt.Sprintf("test-manufacturer-%d", time.Now().UnixNano())
	mBody := models.Manufacturer{Name: mName, Slug: mSlug}
	mJSON, _ := json.Marshal(mBody)

	req := httptest.NewRequest(http.MethodPost, "/products/manufacturers", bytes.NewBuffer(mJSON))
	w := httptest.NewRecorder()
	crud.ManufacturersHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdM models.Manufacturer
	json.NewDecoder(w.Body).Decode(&createdM)
	assert.NotEmpty(t, createdM.ID)
	assert.Equal(t, mName, createdM.Name)

	// 2. Get All
	req = httptest.NewRequest(http.MethodGet, "/products/manufacturers", nil)
	w = httptest.NewRecorder()
	crud.ManufacturersHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var listM []models.Manufacturer
	json.NewDecoder(w.Body).Decode(&listM)
	assert.NotEmpty(t, listM)

	// 3. Get One
	req = httptest.NewRequest(http.MethodGet, "/products/manufacturers/"+createdM.ID, nil)
	w = httptest.NewRecorder()
	crud.ManufacturerItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var getM models.Manufacturer
	json.NewDecoder(w.Body).Decode(&getM)
	assert.Equal(t, createdM.ID, getM.ID)

	// 4. Update
	newName := mName + " Updated"
	createdM.Name = newName
	updateJSON, _ := json.Marshal(createdM)
	req = httptest.NewRequest(http.MethodPut, "/products/manufacturers/"+createdM.ID, bytes.NewBuffer(updateJSON))
	w = httptest.NewRecorder()
	crud.ManufacturerItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var updatedM models.Manufacturer
	json.NewDecoder(w.Body).Decode(&updatedM)
	assert.Equal(t, newName, updatedM.Name)

	// 5. Delete
	req = httptest.NewRequest(http.MethodDelete, "/products/manufacturers/"+createdM.ID, nil)
	w = httptest.NewRecorder()
	crud.ManufacturerItemHandler(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify Delete
	req = httptest.NewRequest(http.MethodGet, "/products/manufacturers/"+createdM.ID, nil)
	w = httptest.NewRecorder()
	crud.ManufacturerItemHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCategoriesCRUD(t *testing.T) {
	setupTestDB(t)

	// 1. Create
	cName := fmt.Sprintf("Test Category %d", time.Now().UnixNano())
	cSlug := fmt.Sprintf("test-category-%d", time.Now().UnixNano())
	cBody := models.Category{Name: cName, Slug: cSlug}
	cJSON, _ := json.Marshal(cBody)

	req := httptest.NewRequest(http.MethodPost, "/products/categories", bytes.NewBuffer(cJSON))
	w := httptest.NewRecorder()
	crud.CategoriesHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdC models.Category
	json.NewDecoder(w.Body).Decode(&createdC)
	assert.NotEmpty(t, createdC.ID)

	// 2. Get All
	req = httptest.NewRequest(http.MethodGet, "/products/categories", nil)
	w = httptest.NewRecorder()
	crud.CategoriesHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Get One
	req = httptest.NewRequest(http.MethodGet, "/products/categories/"+createdC.ID, nil)
	w = httptest.NewRecorder()
	crud.CategoryItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Update
	newName := cName + " Updated"
	createdC.Name = newName
	updateJSON, _ := json.Marshal(createdC)
	req = httptest.NewRequest(http.MethodPut, "/products/categories/"+createdC.ID, bytes.NewBuffer(updateJSON))
	w = httptest.NewRecorder()
	crud.CategoryItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Delete
	req = httptest.NewRequest(http.MethodDelete, "/products/categories/"+createdC.ID, nil)
	w = httptest.NewRecorder()
	crud.CategoryItemHandler(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestProductsCRUD(t *testing.T) {
	setupTestDB(t)

	// Need Manufacturer and Category first
	var m models.Manufacturer
	core.DB.Get(&m, "SELECT id FROM manufacturers LIMIT 1")
	var c models.Category
	core.DB.Get(&c, "SELECT id FROM categories LIMIT 1")

	// 1. Create
	pName := fmt.Sprintf("Test Product %d", time.Now().UnixNano())
	pSlug := fmt.Sprintf("test-product-%d", time.Now().UnixNano())
	pBody := models.Product{
		Name:           pName,
		Slug:           pSlug,
		ManufacturerID: m.ID,
		CategoryID:     c.ID,
		Price:          1000,
		Stock:          10,
		Availability:   "in_stock",
	}
	pJSON, _ := json.Marshal(pBody)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(pJSON))
	w := httptest.NewRecorder()
	crud.ProductsHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdP models.Product
	json.NewDecoder(w.Body).Decode(&createdP)
	assert.NotEmpty(t, createdP.ID)

	// 2. Get All
	req = httptest.NewRequest(http.MethodGet, "/products", nil)
	w = httptest.NewRecorder()
	crud.ProductsHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Get One
	req = httptest.NewRequest(http.MethodGet, "/products/"+createdP.ID, nil)
	w = httptest.NewRecorder()
	crud.ProductItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Update
	newName := pName + " Updated"
	createdP.Name = newName
	updateJSON, _ := json.Marshal(createdP)
	req = httptest.NewRequest(http.MethodPut, "/products/"+createdP.ID, bytes.NewBuffer(updateJSON))
	w = httptest.NewRecorder()
	crud.ProductItemHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Delete
	req = httptest.NewRequest(http.MethodDelete, "/products/"+createdP.ID, nil)
	w = httptest.NewRecorder()
	crud.ProductItemHandler(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCartAndOrderFlow(t *testing.T) {
	setupTestDB(t)
	sessionID := "test-flow-session"

	// Get Product
	var p models.Product
	core.DB.Get(&p, "SELECT id FROM products LIMIT 1")

	// 1. Add to Cart
	cartBody := map[string]interface{}{"productId": p.ID, "quantity": 2}
	cartJSON, _ := json.Marshal(cartBody)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(cartJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()
	crud.CartHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Get Cart
	req = httptest.NewRequest(http.MethodGet, "/cart", nil)
	req.Header.Set("X-Session-ID", sessionID)
	w = httptest.NewRecorder()
	crud.CartHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var cartResp crud.CartResponse
	json.NewDecoder(w.Body).Decode(&cartResp)
	assert.Equal(t, 2, cartResp.Count)

	// 3. Create Order
	email := fmt.Sprintf("flow-%d@example.com", time.Now().UnixNano())
	orderBody := models.CheckoutForm{
		Name:         "Flow User",
		Phone:        "+77001112233",
		Email:        email,
		Address:      "Flow Address",
		CustomerType: "individual",
	}
	orderJSON, _ := json.Marshal(orderBody)
	req = httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w = httptest.NewRecorder()
	crud.OrdersHandler(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	orderID := orderResp["orderId"].(string)

	// 4. Delete Order
	req = httptest.NewRequest(http.MethodDelete, "/orders/"+orderID, nil)
	w = httptest.NewRecorder()
	crud.OrderItemHandler(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func setupTestDB(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/noble"
	}
	if err := core.InitDB(dsn); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	crud.SetDB(core.DB)
}
