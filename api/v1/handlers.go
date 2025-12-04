package v1

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"noble-group-services/crud"
)

// SetupRoutes sets up the API routes.
// It accepts a mux to register handlers.
func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/products", crud.ProductsHandler)
	mux.HandleFunc("/products/categories", crud.CategoriesHandler)
	mux.HandleFunc("/products/categories/", crud.CategoryItemHandler)
	mux.HandleFunc("/products/manufacturers", crud.ManufacturersHandler)
	mux.HandleFunc("/products/manufacturers/", crud.ManufacturerItemHandler)
	mux.HandleFunc("/products/", crud.ProductItemHandler)
	mux.HandleFunc("/cart", crud.CartHandler)
	mux.HandleFunc("/cart/", crud.CartItemHandler)
	mux.HandleFunc("/orders", crud.OrdersHandler)
	mux.HandleFunc("/orders/", crud.OrderItemHandler)

	// Swagger documentation
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
