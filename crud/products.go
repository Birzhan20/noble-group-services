package crud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/core"
	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
	"github.com/lib/pq"
)

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
	productsChan := make(chan []models.Product)
	errChan := make(chan error)

	go func() {
		query := r.URL.Query()
		category := query.Get("category")
		manufacturer := query.Get("manufacturer")
		search := strings.ToLower(query.Get("search"))
		inStockOnly := query.Get("inStockOnly") == "true"
		page, _ := strconv.Atoi(query.Get("page"))
		limit, _ := strconv.Atoi(query.Get("limit"))

		if page < 1 {
			page = 1
		}
		if limit <= 0 {
			limit = 20
		}
		offset := (page - 1) * limit

		sqlQuery := `SELECT id, name, category, manufacturer, availability, price, old_price, description, features, image, stock, rating, reviews, sku FROM products WHERE 1=1`
		var args []interface{}
		argCount := 1

		if category != "" {
			sqlQuery += fmt.Sprintf(" AND category = $%d", argCount)
			args = append(args, category)
			argCount++
		}
		if manufacturer != "" {
			sqlQuery += fmt.Sprintf(" AND manufacturer = $%d", argCount)
			args = append(args, manufacturer)
			argCount++
		}
		if search != "" {
			sqlQuery += fmt.Sprintf(" AND LOWER(name) LIKE $%d", argCount)
			args = append(args, "%"+search+"%")
			argCount++
		}
		if inStockOnly {
			sqlQuery += fmt.Sprintf(" AND availability = 'in-stock'")
		}

		sqlQuery += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, limit, offset)

		rows, err := core.DB.Query(sqlQuery, args...)
		if err != nil {
			errChan <- err
			return
		}
		defer rows.Close()

		var products []models.Product
		for rows.Next() {
			var p models.Product
			var features []string
			// Using pq.Array to scan features array
			err := rows.Scan(
				&p.ID, &p.Name, &p.Category, &p.Manufacturer, &p.Availability, &p.Price, &p.OldPrice,
				&p.Description, pq.Array(&features), &p.Image, &p.Stock, &p.Rating, &p.Reviews, &p.SKU,
			)
			if err != nil {
				errChan <- err
				return
			}
			p.Features = features
			products = append(products, p)
		}

		// Ensure we send an empty slice, not nil
		if products == nil {
			products = []models.Product{}
		}

		productsChan <- products
	}()

	select {
	case products := <-productsChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(products)
	case err := <-errChan:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")

	productChan := make(chan *models.Product)
	errChan := make(chan error)

	go func() {
		var p models.Product
		var features []string
		err := core.DB.QueryRow(`
			SELECT id, name, category, manufacturer, availability, price, old_price, description, features, image, stock, rating, reviews, sku
			FROM products WHERE id = $1
		`, id).Scan(
			&p.ID, &p.Name, &p.Category, &p.Manufacturer, &p.Availability, &p.Price, &p.OldPrice,
			&p.Description, pq.Array(&features), &p.Image, &p.Stock, &p.Rating, &p.Reviews, &p.SKU,
		)
		if err != nil {
			errChan <- err
			return
		}
		p.Features = features
		productChan <- &p
	}()

	select {
	case p := <-productChan:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	case err := <-errChan:
		if strings.Contains(err.Error(), "no rows") {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
