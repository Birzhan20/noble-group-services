package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Product represents a product in the marketplace.
type Product struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Manufacturer string   `json:"manufacturer"`
	Availability string   `json:"availability"`
	Price        int      `json:"price"`
	OldPrice     *int     `json:"oldPrice,omitempty"`
	Description  string   `json:"description"`
	Features     []string `json:"features"`
	Image        string   `json:"image"`
	Stock        int      `json:"stock"`
	Rating       float64  `json:"rating"`
	Reviews      int      `json:"reviews"`
	SKU          string   `json:"sku"`
}

// CartItem represents an item in the shopping cart.
type CartItem struct {
	Product
	Quantity int `json:"quantity"`
}

// CheckoutForm represents the checkout form data.
type CheckoutForm struct {
	CustomerType string `json:"customerType"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	CompanyName  *string `json:"companyName,omitempty"`
	BIN          *string `json:"bin,omitempty"`
	Address      string `json:"address"`
	Comment      *string `json:"comment,omitempty"`
}

// Using a map for mock product data
var products = map[string]Product{
	"1": {ID: "1", Name: "Респиратор 3M 7502", Category: "Респираторы", Manufacturer: "3M", Availability: "in-stock", Price: 12500, Stock: 100, Rating: 4.8, Reviews: 124, SKU: "3M-7502", Image: "https://noble-group.vercel.app/images/product1.jpg"},
	"2": {ID: "2", Name: "Перчатки Ansell HyFlex", Category: "Перчатки", Manufacturer: "Ansell", Availability: "in-stock", Price: 5000, Stock: 200, Rating: 4.5, Reviews: 80, SKU: "AN-11-800", Image: "https://noble-group.vercel.app/images/product2.jpg"},
	"3": {ID: "3", Name: "Каска MSA V-Gard", Category: "Каски", Manufacturer: "MSA Safety", Availability: "pre-order", Price: 8000, Stock: 0, Rating: 4.7, Reviews: 50, SKU: "MSA-VGARD", Image: "https://noble-group.vercel.app/images/product3.jpg"},
	"4": {ID: "4", Name: "Очки Uvex Skyper", Category: "Очки", Manufacturer: "Uvex", Availability: "in-stock", Price: 3500, Stock: 150, Rating: 4.6, Reviews: 95, SKU: "UV-9195", Image: "https://noble-group.vercel.app/images/product4.jpg"},
	"5": {ID: "5", Name: "Костюм Tyvek Classic", Category: "Спецодежда", Manufacturer: "Tyvek", Availability: "in-stock", Price: 15000, Stock: 50, Rating: 4.9, Reviews: 200, SKU: "TY-CLASSIC", Image: "https://noble-group.vercel.app/images/product5.jpg"},
}

func main() {
	http.HandleFunc("/api/v1/products", productsHandler)
	http.HandleFunc("/api/v1/products/categories", categoriesHandler)
	http.HandleFunc("/api/v1/products/manufacturers", manufacturersHandler)
	http.HandleFunc("/api/v1/products/", productByIDHandler)
	http.HandleFunc("/api/v1/cart", cartHandler)
	http.HandleFunc("/api/v1/cart/", cartItemHandler) // Note the trailing slash
	http.HandleFunc("/api/v1/orders", ordersHandler)

	http.ListenAndServe(":3000", nil)
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	productsChan := make(chan []Product)

	go func() {
		query := r.URL.Query()
		category := query.Get("category")
		manufacturer := query.Get("manufacturer")
		search := query.Get("search")
		inStockOnly := query.Get("inStockOnly") == "true"

		var filteredProducts []Product

		for _, p := range products {
			if category != "" && p.Category != category {
				continue
			}
			if manufacturer != "" && p.Manufacturer != manufacturer {
				continue
			}
			if search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(search)) {
				continue
			}
			if inStockOnly && p.Availability != "in-stock" {
				continue
			}
			filteredProducts = append(filteredProducts, p)
		}

		// Pagination
		page, _ := strconv.Atoi(query.Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(query.Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		startIndex := (page - 1) * limit
		endIndex := page * limit
		if startIndex >= len(filteredProducts) {
			productsChan <- []Product{}
			return
		}
		if endIndex > len(filteredProducts) {
			endIndex = len(filteredProducts)
		}

		productsChan <- filteredProducts[startIndex:endIndex]
	}()

	productList := <-productsChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(productList)
}

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	categoriesChan := make(chan map[string][]string)

	go func() {
		categories := map[string][]string{"categories": {"Все", "Респираторы", "Перчатки", "Каски", "Очки", "Спецодежда"}}
		categoriesChan <- categories
	}()

	categories := <-categoriesChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func manufacturersHandler(w http.ResponseWriter, r *http.Request) {
	manufacturersChan := make(chan map[string][]string)

	go func() {
		manufacturers := map[string][]string{"manufacturers": {"3M", "Honeywell", "Ansell", "MSA Safety", "Tyvek", "Uvex"}}
		manufacturersChan <- manufacturers
	}()

	manufacturers := <-manufacturersChan

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manufacturers)
}

func productByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")

	productChan := make(chan Product)

	go func() {
		product, ok := products[id]
		if !ok {
			productChan <- Product{}
			return
		}
		productChan <- product
	}()

	product := <-productChan

	if product.ID == "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
		w.Header().Set("X-Session-ID", sessionID)
	}

	cart := getCart(sessionID)

	switch r.Method {
	case http.MethodGet:
		// Handled by returning the cart at the end
	case http.MethodPost:
		var reqBody struct {
			ProductID string `json:"productId"`
			Quantity  int    `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if reqBody.Quantity <= 0 {
			reqBody.Quantity = 1
		}

		product, ok := products[reqBody.ProductID]
		if !ok {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}

		// Check if item is already in cart
		var found bool
		for i := range cart.Items {
			if cart.Items[i].ID == reqBody.ProductID {
				cart.Items[i].Quantity += reqBody.Quantity
				found = true
				break
			}
		}

		if !found {
			cart.Items = append(cart.Items, CartItem{Product: product, Quantity: reqBody.Quantity})
		}

	case http.MethodDelete:
		cart.Items = []CartItem{}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cart.calculateTotals()
	updateCart(sessionID, cart)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func cartItemHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID header is required", http.StatusBadRequest)
		return
	}

	productID := strings.TrimPrefix(r.URL.Path, "/api/v1/cart/")

	cart := getCart(sessionID)

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

	cart.calculateTotals()
	updateCart(sessionID, cart)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	// Mock data
	order := map[string]interface{}{
		"success":     true,
		"orderId":     "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
		"orderNumber": "ORD-2025-000123",
		"total":       12500,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
