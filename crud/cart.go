package crud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

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

	switch r.Method {
	case http.MethodGet:
		// Just fetching, handled at end
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

		// Check if product exists using DB
		var exists bool
		err := core.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", reqBody.ProductID).Scan(&exists)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		// Insert or Update logic:
		// ON CONFLICT (session_id, product_id) DO UPDATE SET quantity = cart_items.quantity + excluded.quantity
		_, err = core.DB.Exec(`
			INSERT INTO cart_items (session_id, product_id, quantity)
			VALUES ($1, $2, $3)
			ON CONFLICT (session_id, product_id)
			DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()
		`, sessionID, reqBody.ProductID, quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case http.MethodDelete:
		_, err := core.DB.Exec("DELETE FROM cart_items WHERE session_id = $1", sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cart, err := getCartFromDB(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

		res, err := core.DB.Exec(`
			UPDATE cart_items SET quantity = $1, updated_at = NOW()
			WHERE session_id = $2 AND product_id = $3
		`, reqBody.Quantity, sessionID, productID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Product not in cart", http.StatusNotFound)
			return
		}

	case http.MethodDelete:
		res, err := core.DB.Exec(`
			DELETE FROM cart_items
			WHERE session_id = $1 AND product_id = $2
		`, sessionID, productID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Product not in cart", http.StatusNotFound)
			return
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cart, err := getCartFromDB(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CartResponse{
		Items: cart.Items,
		Total: cart.FinalTotal,
		Count: len(cart.Items),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getCartFromDB retrieves the cart for a given session ID from the database.
func getCartFromDB(sessionID string) (*models.Cart, error) {
	// Join cart_items with products to get full product details
	rows, err := core.DB.Query(`
		SELECT
			p.id, p.name, p.category, p.manufacturer, p.availability, p.price, p.old_price,
			p.description, p.features, p.image, p.stock, p.rating, p.reviews, p.sku,
			ci.quantity
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.session_id = $1
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query cart items: %w", err)
	}
	defer rows.Close()

	cart := &models.Cart{
		Items: []models.CartItem{},
	}

	for rows.Next() {
		var p models.Product
		var features []string
		var quantity int

		err := rows.Scan(
			&p.ID, &p.Name, &p.Category, &p.Manufacturer, &p.Availability, &p.Price, &p.OldPrice,
			&p.Description, pq.Array(&features), &p.Image, &p.Stock, &p.Rating, &p.Reviews, &p.SKU,
			&quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		p.Features = features
		cart.Items = append(cart.Items, models.CartItem{Product: p, Quantity: quantity})
	}

	cart.CalculateTotals()
	return cart, nil
}
