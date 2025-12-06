package v1

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"noble-group-services/crud"
)

// SetupRoutes sets up the API routes.
// It accepts a mux to register handlers.
// IMPORTANT: More specific routes MUST be registered before less specific ones
// because http.ServeMux uses longest-prefix matching.
func SetupRoutes(mux *http.ServeMux) {
	// Categories routes (more specific, must come before /products/)
	mux.HandleFunc("/products/categories/", crud.CategoryItemHandler)
	mux.HandleFunc("/products/categories", crud.CategoriesHandler)

	// Manufacturers routes (more specific, must come before /products/)
	mux.HandleFunc("/products/manufacturers/", crud.ManufacturerItemHandler)
	mux.HandleFunc("/products/manufacturers", crud.ManufacturersHandler)

	// Products routes (less specific)
	mux.HandleFunc("/products/", crud.ProductItemHandler)
	mux.HandleFunc("/products", crud.ProductsHandler)

	// Cart routes
	mux.HandleFunc("/cart/", crud.CartItemHandler)
	mux.HandleFunc("/cart", crud.CartHandler)

	// Orders routes
	mux.HandleFunc("/orders/", crud.OrderItemHandler)
	mux.HandleFunc("/orders", crud.OrdersHandler)

	// Swagger documentation
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
