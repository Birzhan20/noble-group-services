package v1

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/crud"
)

// SetupRoutes sets up the API routes.
func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/products", crud.ProductsHandler)
	mux.HandleFunc("/api/v1/products/categories", crud.CategoriesHandler)
	mux.HandleFunc("/api/v1/products/manufacturers", crud.ManufacturersHandler)
	mux.HandleFunc("/api/v1/products/", crud.ProductByIDHandler)
	mux.HandleFunc("/api/v1/cart", crud.CartHandler)
	mux.HandleFunc("/api/v1/cart/", crud.CartItemHandler)
	mux.HandleFunc("/api/v1/orders", crud.OrdersHandler)

	// Swagger documentation
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
}
