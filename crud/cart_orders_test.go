package crud

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"noble-group-services/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ================== Cart Unit Tests ==================

func TestCartHandler_Get_NewSession(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/cart", nil)
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Empty(t, response.Items)
	assert.Equal(t, 0, response.Total)
	assert.Equal(t, 0, response.Count)
}

func TestCartHandler_Get_NoSessionID_GeneratesNew(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/cart", nil)
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should set X-Session-ID header
	assert.NotEmpty(t, w.Header().Get("X-Session-ID"))
}

func TestCartHandler_Post_AddItem(t *testing.T) {
	setupTestDB(t)

	// Get a product ID
	var p models.Product
	err := db.Get(&p, "SELECT id, price, stock FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()
	body := map[string]interface{}{
		"productId": p.ID,
		"quantity":  1,
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, 1, len(response.Items))
	assert.Equal(t, 1, response.Count)
}

func TestCartHandler_Post_AddSameItemTwice(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id, price, stock FROM products WHERE stock >= 5 LIMIT 1")
	if err != nil {
		t.Skip("No products with enough stock")
	}

	sessionID := uuid.New().String()
	body := map[string]interface{}{
		"productId": p.ID,
		"quantity":  2,
	}
	bodyJSON, _ := json.Marshal(body)

	// Add first time
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()
	CartHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Add second time
	req = httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w = httptest.NewRecorder()
	CartHandler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	json.NewDecoder(w.Body).Decode(&response)
	// Should have 4 total (2 + 2)
	assert.Equal(t, 4, response.Count)
	assert.Equal(t, 1, len(response.Items)) // Still one item, just higher quantity
}

func TestCartHandler_Post_InvalidProductID(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	body := map[string]interface{}{
		"productId": "non-existent-product",
		"quantity":  1,
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCartHandler_Post_MissingProductID(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	body := map[string]interface{}{
		"quantity": 1,
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCartHandler_Post_NotEnoughStock(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id, stock FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()
	body := map[string]interface{}{
		"productId": p.ID,
		"quantity":  p.Stock + 100, // More than available
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCartHandler_Delete_ClearCart(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()

	// Add item
	body := map[string]interface{}{"productId": p.ID, "quantity": 1}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()
	CartHandler(w, req)

	// Clear cart
	req = httptest.NewRequest(http.MethodDelete, "/cart", nil)
	req.Header.Set("X-Session-ID", sessionID)
	w = httptest.NewRecorder()
	CartHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Empty(t, response.Items)
	assert.Equal(t, 0, response.Count)
}

func TestCartItemHandler_Patch_UpdateQuantity(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()

	// Add item
	body := map[string]interface{}{"productId": p.ID, "quantity": 1}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	CartHandler(httptest.NewRecorder(), req)

	// Update quantity
	updateBody := map[string]interface{}{"quantity": 5}
	updateJSON, _ := json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPatch, "/cart/"+p.ID, bytes.NewBuffer(updateJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()
	CartItemHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, 5, response.Count)
}

func TestCartItemHandler_Patch_NoSessionID(t *testing.T) {
	setupTestDB(t)

	updateBody := map[string]interface{}{"quantity": 5}
	updateJSON, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPatch, "/cart/some-id", bytes.NewBuffer(updateJSON))
	w := httptest.NewRecorder()

	CartItemHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCartItemHandler_Patch_InvalidQuantity(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	updateBody := map[string]interface{}{"quantity": 0}
	updateJSON, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPatch, "/cart/some-id", bytes.NewBuffer(updateJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartItemHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCartItemHandler_Patch_ItemNotInCart(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	updateBody := map[string]interface{}{"quantity": 5}
	updateJSON, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPatch, "/cart/non-existent", bytes.NewBuffer(updateJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartItemHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCartItemHandler_Delete_RemoveItem(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()

	// Add item
	body := map[string]interface{}{"productId": p.ID, "quantity": 3}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	CartHandler(httptest.NewRecorder(), req)

	// Remove item
	req = httptest.NewRequest(http.MethodDelete, "/cart/"+p.ID, nil)
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()
	CartItemHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CartResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Empty(t, response.Items)
}

func TestCartItemHandler_Delete_ItemNotInCart(t *testing.T) {
	setupTestDB(t)

	sessionID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/cart/non-existent", nil)
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	CartItemHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCartItemHandler_MethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/cart/some-id", nil)
	req.Header.Set("X-Session-ID", "test")
	w := httptest.NewRecorder()

	CartItemHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// ================== Orders Unit Tests ==================

func TestOrdersHandler_Post_ValidOrder(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id, price FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	sessionID := uuid.New().String()

	// Add item to cart first
	body := map[string]interface{}{"productId": p.ID, "quantity": 1}
	bodyJSON, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/cart", bytes.NewBuffer(bodyJSON))
	req.Header.Set("X-Session-ID", sessionID)
	CartHandler(httptest.NewRecorder(), req)

	// Create order
	orderForm := models.CheckoutForm{
		Name:         "Test Customer",
		Phone:        "+77001234567",
		Email:        "test@example.com",
		Address:      "Test Address 123, City, Country",
		CustomerType: "individual",
	}
	orderJSON, _ := json.Marshal(orderForm)
	req = httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("X-Session-ID", sessionID)
	w := httptest.NewRecorder()

	OrdersHandler(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.True(t, response["success"].(bool))
	assert.NotEmpty(t, response["orderId"])
	assert.NotEmpty(t, response["orderNumber"])

	// Cleanup
	orderID := response["orderId"].(string)
	_, _ = db.Exec("DELETE FROM orders WHERE id = $1", orderID)
}

func TestOrdersHandler_Post_WithCartsField(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id, price FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	orderForm := models.CheckoutForm{
		Name:         "Test Customer Carts",
		Phone:        "+77001234567",
		Email:        "test-carts@example.com",
		Address:      "Test Address with Carts Field",
		CustomerType: "individual",
		Carts: []models.CartItemRequest{
			{ProductID: p.ID, Quantity: 2},
		},
	}
	orderJSON, _ := json.Marshal(orderForm)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("X-Session-ID", uuid.New().String())
	w := httptest.NewRecorder()

	OrdersHandler(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())

	var response map[string]interface{}
	json.NewDecoder(w.Body).Decode(&response)
	assert.True(t, response["success"].(bool))

	// Cleanup
	orderID := response["orderId"].(string)
	_, _ = db.Exec("DELETE FROM orders WHERE id = $1", orderID)
}

func TestOrdersHandler_Post_LegalEntityValidation(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	// Legal entity without companyName and BIN
	orderForm := models.CheckoutForm{
		Name:         "Legal Test",
		Phone:        "+77001234567",
		Email:        "legal@example.com",
		Address:      "Legal Address 123",
		CustomerType: "legal",
		Carts: []models.CartItemRequest{
			{ProductID: p.ID, Quantity: 1},
		},
	}
	orderJSON, _ := json.Marshal(orderForm)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	w := httptest.NewRecorder()

	OrdersHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ValidationErrorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.NotEmpty(t, response.Details)
}

func TestOrdersHandler_Post_ValidationErrors(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name        string
		form        models.CheckoutForm
		expectField string
	}{
		{
			name:        "name too short",
			form:        models.CheckoutForm{Name: "A", Phone: "+77001234567", Email: "test@test.com", Address: "Valid Address Here"},
			expectField: "name",
		},
		{
			name:        "phone too short",
			form:        models.CheckoutForm{Name: "Valid Name", Phone: "123", Email: "test@test.com", Address: "Valid Address Here"},
			expectField: "phone",
		},
		{
			name:        "invalid email",
			form:        models.CheckoutForm{Name: "Valid Name", Phone: "+77001234567", Email: "invalid-email", Address: "Valid Address Here"},
			expectField: "email",
		},
		{
			name:        "address too short",
			form:        models.CheckoutForm{Name: "Valid Name", Phone: "+77001234567", Email: "test@test.com", Address: "Short"},
			expectField: "address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderJSON, _ := json.Marshal(tt.form)
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
			w := httptest.NewRecorder()

			OrdersHandler(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response ValidationErrorResponse
			json.NewDecoder(w.Body).Decode(&response)
			assert.Equal(t, "VALIDATION_ERROR", response.Error)

			foundField := false
			for _, detail := range response.Details {
				if detail.Field == tt.expectField {
					foundField = true
					break
				}
			}
			assert.True(t, foundField, "Expected validation error for field: %s", tt.expectField)
		})
	}
}

func TestOrdersHandler_Post_EmptyCart(t *testing.T) {
	setupTestDB(t)

	orderForm := models.CheckoutForm{
		Name:         "Empty Cart Test",
		Phone:        "+77001234567",
		Email:        "empty@test.com",
		Address:      "Valid Address Here",
		CustomerType: "individual",
	}
	orderJSON, _ := json.Marshal(orderForm)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	req.Header.Set("X-Session-ID", uuid.New().String()) // New session, empty cart
	w := httptest.NewRecorder()

	OrdersHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrdersHandler_MethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	w := httptest.NewRecorder()

	OrdersHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestOrderItemHandler_Delete(t *testing.T) {
	setupTestDB(t)

	var p models.Product
	err := db.Get(&p, "SELECT id FROM products WHERE stock > 0 LIMIT 1")
	if err != nil {
		t.Skip("No products in database")
	}

	// Create an order first
	orderForm := models.CheckoutForm{
		Name:         "Delete Test",
		Phone:        "+77001234567",
		Email:        "delete@test.com",
		Address:      "Delete Test Address",
		CustomerType: "individual",
		Carts: []models.CartItemRequest{
			{ProductID: p.ID, Quantity: 1},
		},
	}
	orderJSON, _ := json.Marshal(orderForm)
	createReq := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(orderJSON))
	createW := httptest.NewRecorder()
	OrdersHandler(createW, createReq)
	require.Equal(t, http.StatusCreated, createW.Code)

	var createResponse map[string]interface{}
	json.NewDecoder(createW.Body).Decode(&createResponse)
	orderID := createResponse["orderId"].(string)

	// Delete order
	deleteReq := httptest.NewRequest(http.MethodDelete, "/orders/"+orderID, nil)
	deleteW := httptest.NewRecorder()
	OrderItemHandler(deleteW, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteW.Code)
}

func TestOrderItemHandler_DeleteNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodDelete, "/orders/non-existent-order-id", nil)
	w := httptest.NewRecorder()

	OrderItemHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOrderItemHandler_MethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/orders/some-id", nil)
	w := httptest.NewRecorder()

	OrderItemHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
