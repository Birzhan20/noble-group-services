package crud

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// Using a map for mock product data
var products = map[string]models.Product{
	"1": {ID: "1", Name: "Респиратор 3M 7502", Category: "Респираторы", Manufacturer: "3M", Availability: "in-stock", Price: 12500, Stock: 100, Rating: 4.8, Reviews: 124, SKU: "3M-7502", Image: "https://noble-group.vercel.app/images/product1.jpg"},
	"2": {ID: "2", Name: "Перчатки Ansell HyFlex", Category: "Перчатки", Manufacturer: "Ansell", Availability: "in-stock", Price: 5000, Stock: 200, Rating: 4.5, Reviews: 80, SKU: "AN-11-800", Image: "https://noble-group.vercel.app/images/product2.jpg"},
	"3": {ID: "3", Name: "Каска MSA V-Gard", Category: "Каски", Manufacturer: "MSA Safety", Availability: "pre-order", Price: 8000, Stock: 0, Rating: 4.7, Reviews: 50, SKU: "MSA-VGARD", Image: "https://noble-group.vercel.app/images/product3.jpg"},
	"4": {ID: "4", Name: "Очки Uvex Skyper", Category: "Очки", Manufacturer: "Uvex", Availability: "in-stock", Price: 3500, Stock: 150, Rating: 4.6, Reviews: 95, SKU: "UV-9195", Image: "https://noble-group.vercel.app/images/product4.jpg"},
	"5": {ID: "5", Name: "Костюм Tyvek Classic", Category: "Спецодежда", Manufacturer: "Tyvek", Availability: "in-stock", Price: 15000, Stock: 50, Rating: 4.9, Reviews: 200, SKU: "TY-CLASSIC", Image: "https://noble-group.vercel.app/images/product5.jpg"},
}

// ProductsHandler handles product-related requests.
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
	productsChan := make(chan []models.Product)

	go func() {
		query := r.URL.Query()
		category := query.Get("category")
		manufacturer := query.Get("manufacturer")
		search := query.Get("search")
		inStockOnly := query.Get("inStockOnly") == "true"

		var filteredProducts []models.Product

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
			productsChan <- []models.Product{}
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

// ProductByIDHandler handles requests for a single product by its ID.
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
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")

	productChan := make(chan models.Product)

	go func() {
		product, ok := products[id]
		if !ok {
			productChan <- models.Product{}
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
