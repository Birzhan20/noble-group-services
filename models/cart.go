package models

// Cart represents a shopping cart.
type Cart struct {
	Items      []CartItem `json:"items"`
	Total      int        `json:"total"`
	Discount   int        `json:"discount"`
	FinalTotal int        `json:"finalTotal"`
}

// CalculateTotals calculates the total, discount, and final total for the cart.
func (c *Cart) CalculateTotals() {
	c.Total = 0
	for _, item := range c.Items {
		c.Total += item.Price * item.Quantity
	}

	// Placeholder for discount logic
	c.Discount = 0
	c.FinalTotal = c.Total - c.Discount
}
