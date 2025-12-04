// crud/cart.go
package crud

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"

	"noble-group-services/models"
)

// Глобальное хранилище корзин (в памяти, для гостей)
var carts = make(map[string]*models.Cart)
var cartMu sync.RWMutex

// CartResponse — ответ для фронта
type CartResponse struct {
	Items []models.CartItem `json:"items"`
	Total int               `json:"total"`
	Count int               `json:"count"` // ← теперь общее количество товаров!
}

// CartHandler — GET /cart, POST /cart, DELETE /cart
func CartHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetCart(w, r)
	case http.MethodPost:
		AddToCart(w, r)
	case http.MethodDelete:
		ClearCart(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetCart godoc
// @Summary Get current cart
// @Description Get the current session's cart
// @Tags cart
// @Produce json
// @Param X-Session-ID header string false "Session ID"
// @Success 200 {object} CartResponse
// @Router /cart [get]
func GetCart(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(w, r)

	cartMu.RLock()
	cart := getCartUnsafe(sessionID)
	response := CartResponse{
		Items: cart.Items,
		Total: cart.FinalTotal,
		Count: cart.Count,
	}
	cartMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add a product to the cart
// @Tags cart
// @Accept json
// @Produce json
// @Param X-Session-ID header string false "Session ID"
// @Param request body object{productId=string,quantity=int} true "Product ID and Quantity"
// @Success 200 {object} CartResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Product not found"
// @Router /cart [post]
func AddToCart(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(w, r)

	var req struct {
		ProductID string `json:"productId"`
		Quantity  *int   `json:"quantity,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.ProductID == "" {
		http.Error(w, "productId is required", http.StatusBadRequest)
		return
	}

	qty := 1
	if req.Quantity != nil && *req.Quantity > 0 {
		qty = *req.Quantity
	}

	// Загружаем товар из БД
	var product models.Product
	err := db.Get(&product, `
		SELECT 
			p.id, p.name, p.slug, p.price, p.old_price, p.description, 
			p.features, p.image, p.stock, p.sku, p.availability,
			m.id AS "manufacturer.id", m.name AS "manufacturer.name", m.slug AS "manufacturer.slug", m.logo AS "manufacturer.logo",
			c.id AS "category.id", c.name AS "category.name", c.slug AS "category.slug"
		FROM products p
		LEFT JOIN manufacturers m ON p.manufacturer_id = m.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`, req.ProductID)

	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	if product.Stock < qty {
		http.Error(w, "Not enough stock", http.StatusBadRequest)
		return
	}

	cartMu.Lock()
	cart := getCartUnsafe(sessionID)
	found := false
	for i := range cart.Items {
		if cart.Items[i].ID == req.ProductID {
			cart.Items[i].Quantity += qty
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, models.CartItem{
			Product:  product,
			Quantity: qty,
		})
	}
	cart.CalculateTotals()
	cartMu.Unlock()

	respondCart(w, cart)
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from the cart
// @Tags cart
// @Produce json
// @Param X-Session-ID header string false "Session ID"
// @Success 200 {object} CartResponse
// @Router /cart [delete]
func ClearCart(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(w, r)

	cartMu.Lock()
	cart := getCartUnsafe(sessionID)
	cart.Items = []models.CartItem{}
	cart.CalculateTotals()
	cartMu.Unlock()

	respondCart(w, cart)
}

func getSessionID(w http.ResponseWriter, r *http.Request) string {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
		w.Header().Set("X-Session-ID", sessionID)
	}
	return sessionID
}

func respondCart(w http.ResponseWriter, cart *models.Cart) {
	cartMu.RLock()
	response := CartResponse{
		Items: cart.Items,
		Total: cart.FinalTotal,
		Count: cart.Count,
	}
	cartMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CartItemHandler — PATCH /cart/{id}, DELETE /cart/{id}
func CartItemHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		UpdateCartItem(w, r)
	case http.MethodDelete:
		RemoveCartItem(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UpdateCartItem godoc
// @Summary Update cart item quantity
// @Description Update the quantity of a product in the cart
// @Tags cart
// @Accept json
// @Produce json
// @Param X-Session-ID header string false "Session ID"
// @Param id path string true "Product ID"
// @Param request body object{quantity=int} true "New Quantity"
// @Success 200 {object} CartResponse
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Item not found"
// @Router /cart/{id} [patch]
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID required", http.StatusBadRequest)
		return
	}

	productID := strings.TrimPrefix(r.URL.Path, "/cart/")
	if productID == "" {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	cartMu.Lock()
	cart := getCartUnsafe(sessionID)
	cartMu.Unlock()

	var req struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Quantity <= 0 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	cartMu.Lock()
	found := false
	for i := range cart.Items {
		if cart.Items[i].ID == productID {
			cart.Items[i].Quantity = req.Quantity
			found = true
			break
		}
	}
	if !found {
		cartMu.Unlock()
		http.Error(w, "Item not in cart", http.StatusNotFound)
		return
	}
	cart.CalculateTotals()
	cartMu.Unlock()

	respondCart(w, cart)
}

// RemoveCartItem godoc
// @Summary Remove item from cart
// @Description Remove a product from the cart
// @Tags cart
// @Produce json
// @Param X-Session-ID header string false "Session ID"
// @Param id path string true "Product ID"
// @Success 200 {object} CartResponse
// @Failure 404 {string} string "Item not found"
// @Router /cart/{id} [delete]
func RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID required", http.StatusBadRequest)
		return
	}

	productID := strings.TrimPrefix(r.URL.Path, "/cart/")
	if productID == "" {
		http.Error(w, "Product ID required", http.StatusBadRequest)
		return
	}

	cartMu.Lock()
	cart := getCartUnsafe(sessionID)

	found := false
	for i, item := range cart.Items {
		if item.ID == productID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		cartMu.Unlock()
		http.Error(w, "Item not in cart", http.StatusNotFound)
		return
	}

	cart.CalculateTotals()
	cartMu.Unlock()

	respondCart(w, cart)
}

// Вспомогательные функции
func getCartUnsafe(sessionID string) *models.Cart {
	if _, ok := carts[sessionID]; !ok {
		carts[sessionID] = &models.Cart{
			Items: []models.CartItem{},
		}
	}
	return carts[sessionID]
}
