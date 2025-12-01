package crud

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/GoogleCloudPlatform/golang-samples/run/helloworld/models"
)

// ValidationErrorDetail represents a single field validation error.
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents the structured error response.
type ValidationErrorResponse struct {
	Error   string                  `json:"error"`
	Details []ValidationErrorDetail `json:"details"`
}

// OrdersHandler handles order-related requests.
// @Summary Place a new order
// @Description Place a new order with the items in the cart
// @Tags orders
// @Accept  json
// @Produce  json
// @Param checkoutForm body models.CheckoutForm true "Checkout Form"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} ValidationErrorResponse
// @Router /orders [post]
func OrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "X-Session-ID header is required", http.StatusBadRequest)
		return
	}

	cartMu.Lock()
	defer cartMu.Unlock()

	cart := getCartUnsafe(sessionID)

	if len(cart.Items) == 0 {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}

	var form models.CheckoutForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation Logic
	var validationErrors []ValidationErrorDetail

	// Name: min 2 chars
	if utf8.RuneCountInString(form.Name) < 2 {
		validationErrors = append(validationErrors, ValidationErrorDetail{Field: "name", Message: "Имя должно содержать минимум 2 символа"})
	}

	// Phone: min 10 digits, starts with +7, 8, or 7
	// Remove non-digits first
	phoneDigits := regexp.MustCompile(`\D`).ReplaceAllString(form.Phone, "")
	if len(phoneDigits) < 10 {
		validationErrors = append(validationErrors, ValidationErrorDetail{Field: "phone", Message: "Номер телефона должен содержать минимум 10 цифр"})
	} else {
		match, _ := regexp.MatchString(`^(\+7|8|7)`, form.Phone)
		if !match {
			validationErrors = append(validationErrors, ValidationErrorDetail{Field: "phone", Message: "Номер телефона должен начинаться с +7, 7 или 8"})
		}
	}

	// Email: valid email
	// Simple regex
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(form.Email) {
		validationErrors = append(validationErrors, ValidationErrorDetail{Field: "email", Message: "Некорректный e-mail"})
	}

	// Address: min 10 chars
	if utf8.RuneCountInString(form.Address) < 10 {
		validationErrors = append(validationErrors, ValidationErrorDetail{Field: "address", Message: "Адрес должен содержать минимум 10 символов"})
	}

	// Legal entity checks
	if form.CustomerType == "legal" {
		if form.CompanyName == nil || strings.TrimSpace(*form.CompanyName) == "" {
			validationErrors = append(validationErrors, ValidationErrorDetail{Field: "companyName", Message: "Название компании обязательно для юридических лиц"})
		}
		if form.BIN == nil {
			validationErrors = append(validationErrors, ValidationErrorDetail{Field: "bin", Message: "БИН обязателен для юридических лиц"})
		} else {
			// BIN must be exactly 12 digits
			binClean := regexp.MustCompile(`\D`).ReplaceAllString(*form.BIN, "")
			if len(binClean) != 12 {
				validationErrors = append(validationErrors, ValidationErrorDetail{Field: "bin", Message: "БИН должен содержать ровно 12 цифр"})
			}
		}
	}

	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidationErrorResponse{
			Error:   "VALIDATION_ERROR",
			Details: validationErrors,
		})
		return
	}

	orderID := uuid.New().String()
	orderNumber := fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), rand.Intn(1000000))

	// Capture total before clearing
	finalTotal := cart.FinalTotal

	// Clear the cart
	cart.Items = []models.CartItem{}
	cart.CalculateTotals()

	response := map[string]interface{}{
		"success":     true,
		"orderId":     orderID,
		"orderNumber": orderNumber,
		"total":       finalTotal,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
