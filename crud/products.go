package crud

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// products holds our mock database of products.
// Since this is a static map and we only read from it, it is thread-safe for concurrent reads.
var products = map[string]models.Product{
	"1": {ID: "1", Name: "Респиратор 3M 7502", Category: "Респираторы", Manufacturer: "3M", Availability: "in-stock", Price: 12500, Stock: 100, Rating: 4.8, Reviews: 124, SKU: "3M-7502", Image: "https://noble-group.vercel.app/images/product1.jpg"},
	"2": {ID: "2", Name: "Перчатки Ansell HyFlex", Category: "Перчатки", Manufacturer: "Ansell", Availability: "in-stock", Price: 5000, Stock: 200, Rating: 4.5, Reviews: 80, SKU: "AN-11-800", Image: "https://noble-group.vercel.app/images/product2.jpg"},
	"3": {ID: "3", Name: "Каска MSA V-Gard", Category: "Каски", Manufacturer: "MSA Safety", Availability: "pre-order", Price: 8000, Stock: 0, Rating: 4.7, Reviews: 50, SKU: "MSA-VGARD", Image: "https://noble-group.vercel.app/images/product3.jpg"},
	"4": {ID: "4", Name: "Очки Uvex Skyper", Category: "Очки", Manufacturer: "Uvex", Availability: "in-stock", Price: 3500, Stock: 150, Rating: 4.6, Reviews: 95, SKU: "UV-9195", Image: "https://noble-group.vercel.app/images/product4.jpg"},
	"5": {ID: "5", Name: "Костюм Tyvek Classic", Category: "Спецодежда", Manufacturer: "Tyvek", Availability: "in-stock", Price: 15000, Stock: 50, Rating: 4.9, Reviews: 200, SKU: "TY-CLASSIC", Image: "https://noble-group.vercel.app/images/product5.jpg"},
}

// ProductsHandler handles requests to list products.
// It supports filtering by category, manufacturer, search term, and availability.
// It also supports pagination.
//
// @Summary Get a list of products
// @Description Get a list of products with optional filters
// @Tags products
// @Accept  json
// @Produce  json
// @Param category query string false "Category to filter by"
// @Param manufacturer query string false "Manufacturer to filter by"
// @Param search query string false "Search term to filter by name"
// @Param inStockOnly query boolean false "Filter for products in stock only"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of items per page"
// @Success 200 {array} models.Product
// @Router /products [get]
func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	// We use a channel to communicate the result from the goroutine back to the main handler.
	productsChan := make(chan []models.Product)

	// Run the filtering logic in a separate goroutine.
	// This demonstrates concurrent processing, keeping the main thread free until the result is ready.
	go func() {
		// 1. Parse Query Parameters
		query := r.URL.Query()
		category := query.Get("category")
		manufacturer := query.Get("manufacturer")
		search := strings.ToLower(query.Get("search"))
		inStockOnly := query.Get("inStockOnly") == "true"

		// 2. Filter Products
		var filteredProducts []models.Product

		for _, p := range products {
			// Check Category
			if category != "" && p.Category != category {
				continue
			}
			// Check Manufacturer
			if manufacturer != "" && p.Manufacturer != manufacturer {
				continue
			}
			// Check Search Term (Name)
			if search != "" && !strings.Contains(strings.ToLower(p.Name), search) {
				continue
			}
			// Check Availability
			if inStockOnly && p.Availability != "in-stock" {
				continue
			}

			// If it passes all checks, add to list
			filteredProducts = append(filteredProducts, p)
		}

		// 3. Handle Pagination
		page, _ := strconv.Atoi(query.Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(query.Get("limit"))
		if limit <= 0 {
			limit = 20
		}

		startIndex := (page - 1) * limit
		endIndex := startIndex + limit

		// Adjust indices to be within bounds
		total := len(filteredProducts)
		if startIndex >= total {
			// Page is out of range, return empty list
			productsChan <- []models.Product{}
			return
		}
		if endIndex > total {
			endIndex = total
		}

		// Send the slice of products back to the main thread
		productsChan <- filteredProducts[startIndex:endIndex]
	}()

	// Wait for the result from the goroutine
	productList := <-productsChan

	// Return the result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(productList)
}

// ProductByIDHandler handles requests for a single product by its ID.
//
// @Summary Get a single product by its ID
// @Description Get a single product by its ID
// @Tags products
// @Accept  json
// @Produce  json
// @Param id path string true "Product ID"
// @Success 200 {object} models.Product
// @Failure 404 {string} string "Not Found"
// @Router /products/{id} [get]
func ProductByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")

	// Channel to receive the found product (or empty if not found)
	productChan := make(chan *models.Product)

	// Look up the product in a goroutine
	go func() {
		if p, ok := products[id]; ok {
			productChan <- &p
		} else {
			productChan <- nil
		}
	}()

	// Wait for result
	product := <-productChan

	// If product is nil, it wasn't found
	if product == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
