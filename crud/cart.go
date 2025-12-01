package crud

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// Using a map for mock cart data
var carts = make(map[string]*models.Cart)

// cartMu protects access to the carts map AND the individual cart structs.
// For a high-concurrency production app, we would use per-cart locks,
// but for a simple intern-friendly example, a global lock for all cart operations is safest and easiest to reason about.
var cartMu sync.RWMutex

// CartResponse represents the JSON response for the cart.
type CartResponse struct {
	Items []models.CartItem `json:"items"`
	Total int              `json:"total"`
	Count int              `json:"count"`
}

// CartHandler handles cart-related requests.
// @Summary Get or update the cart
// @Description Get the current cart, add an item, or clear the cart.
// @Tags cart
// @Accept  json
// @Produce  json
// @Param body body object{productId=string,quantity=integer} false "Add to cart request"
// @Success 200 {object} CartResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Router /cart [get]
// @Router /cart [post]
// @Router /cart [delete]
func CartHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
		w.Header().Set("X-Session-ID", sessionID)
	}

	// Lock the entire operation to ensure thread safety
	cartMu.Lock()
	defer cartMu.Unlock()

	cart := getCartUnsafe(sessionID)

	switch r.Method {
	case http.MethodGet:
		// Handled by returning the cart at the end
	case http.MethodPost:
		var reqBody struct {
			ProductID string `json:"productId"`
			Quantity  *int   `json:"quantity,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		quantity := 1
		if reqBody.Quantity != nil && *reqBody.Quantity > 0 {
			quantity = *reqBody.Quantity
		}

		// 'products' map is read-only safe, so no lock needed for it
		// but we are holding the global lock anyway.
		product, ok := products[reqBody.ProductID]
		if !ok {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		var found bool
		for i := range cart.Items {
			if cart.Items[i].ID == reqBody.ProductID {
				cart.Items[i].Quantity += quantity
				found = true
				break
			}
		}

		if !found {
			cart.Items = append(cart.Items, models.CartItem{Product: product, Quantity: quantity})
		}

	case http.MethodDelete:
		cart.Items = []models.CartItem{}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cart.CalculateTotals()
	// updateCart is not needed because we are modifying the pointer directly, and map already holds it.
	// But if we replaced the struct, we would need it. 'cart' is a pointer.
	// So just modifying fields is enough.

	response := CartResponse{
		Items: cart.Items,
		Total: cart.FinalTotal,
		Count: len(cart.Items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CartItemHandler handles requests for a single item in the cart.
// @Summary Update or delete an item in the cart
// @Description Update the quantity of an item or delete it from the cart
// @Tags cart
// @Accept  json
// @Produce  json
// @Param productId path string true "Product ID"
// @Param body body object{quantity=integer} false "Update quantity request"
// @Success 200 {object} CartResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Not Found"
// @Router /cart/{productId} [patch]
// @Router /cart/{productId} [delete]
func CartItemHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID header is required", http.StatusBadRequest)
		return
	}

	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/cart/")

	cartMu.Lock()
	defer cartMu.Unlock()

	cart := getCartUnsafe(sessionID)

	switch r.Method {
	case http.MethodPatch:
		var reqBody struct {
			Quantity int `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if reqBody.Quantity <= 0 {
			http.Error(w, "Quantity must be greater than 0", http.StatusBadRequest)
			return
		}

		var found bool
		for i := range cart.Items {
			if cart.Items[i].ID == productID {
				cart.Items[i].Quantity = reqBody.Quantity
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Product not in cart", http.StatusNotFound)
			return
		}

	case http.MethodDelete:
		var found bool
		for i, item := range cart.Items {
			if item.ID == productID {
				cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Product not in cart", http.StatusNotFound)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cart.CalculateTotals()

	response := CartResponse{
		Items: cart.Items,
		Total: cart.FinalTotal,
		Count: len(cart.Items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getCart retrieves the cart for a given session ID.
// This version is exported for use in other handlers (like orders), but they should respect locking.
// Actually, since this package handles the lock, we should export a thread-safe way or make OrdersHandler use the lock.
// Since OrdersHandler is in this package, it can access `cartMu`.
func getCart(sessionID string) *models.Cart {
	cartMu.Lock()
	defer cartMu.Unlock()
	return getCartUnsafe(sessionID)
}

// getCartUnsafe retrieves the cart without locking. Caller must hold cartMu.
func getCartUnsafe(sessionID string) *models.Cart {
	if _, ok := carts[sessionID]; !ok {
		carts[sessionID] = &models.Cart{}
	}
	return carts[sessionID]
}

// updateCart is removed/internalized because we modify the pointer in place under lock.
// For compatibility if needed:
func updateCart(sessionID string, cart *models.Cart) {
	cartMu.Lock()
	defer cartMu.Unlock()
	carts[sessionID] = cart
}
