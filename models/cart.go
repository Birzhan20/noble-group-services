package models

// Cart represents a user's shopping cart.
type Cart struct {
	Items      []CartItem `json:"items"`
	Total      int        `json:"total"`
	Count      int        `json:"count"`
	FinalTotal int        `json:"finalTotal"`
}

// CalculateTotals calculates the total price and item count.
func (c *Cart) CalculateTotals() {
	var total, count int
	for _, item := range c.Items {
		total += item.Price * item.Quantity
		count += item.Quantity
	}
	c.Total = total
	c.FinalTotal = total
	c.Count = count
}
