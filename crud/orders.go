package crud

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// init initializes the random number generator.
// Note: In Go 1.20+, the global random generator is automatically seeded,
// but for compatibility and clarity we can seed it here once.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// OrdersHandler handles order-related requests.
// @Summary Place a new order
// @Description Place a new order with the items in the cart
// @Tags orders
// @Accept  json
// @Produce  json
// @Param checkoutForm body models.CheckoutForm true "Checkout Form"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Router /orders [post]
func OrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID header is required", http.StatusBadRequest)
		return
	}

	// We need to lock the cart while reading and clearing it.
	cartMu.Lock()
	defer cartMu.Unlock()

	cart := getCartUnsafe(sessionID)

	// Check if cart is empty
	if len(cart.Items) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	var form models.CheckoutForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if form.Name == "" || form.Phone == "" || form.Email == "" || form.Address == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	if form.CustomerType == "legal" && (form.CompanyName == nil || form.BIN == nil || *form.CompanyName == "" || *form.BIN == "") {
		http.Error(w, "Company name and BIN are required for legal entities", http.StatusBadRequest)
		return
	}

	orderID := uuid.New().String()
	orderNumber := generateOrderNumber()

	order := models.Order{
		ID:          orderID,
		OrderNumber: orderNumber,
		Total:       cart.FinalTotal,
	}

	// In a real app, you would save the order to a database here.

	// Clear the cart
	cart.Items = []models.CartItem{}
	cart.CalculateTotals()
	// updateCart(sessionID, cart) // Not needed as we modified the pointer in place under lock

	response := map[string]interface{}{
		"success":     true,
		"orderId":     order.ID,
		"orderNumber": order.OrderNumber,
		"total":       order.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// generateOrderNumber generates a unique order number.
func generateOrderNumber() string {
	// Generate a random 6-digit number
	return fmt.Sprintf("№%06d", rand.Intn(1000000))
}
