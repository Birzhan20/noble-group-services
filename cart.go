package main

import (
	"sync"
)

// Cart represents a user's shopping cart.
type Cart struct {
	Items    []CartItem `json:"items"`
	Total    int        `json:"total"`
	Count    int        `json:"count"`
}

// In-memory store for carts, mapping session ID to cart
var (
	carts = make(map[string]*Cart)
	cartsMutex = &sync.RWMutex{}
)

func getCart(sessionID string) *Cart {
	cartsMutex.RLock()
	defer cartsMutex.RUnlock()

	if cart, ok := carts[sessionID]; ok {
		return cart
	}
	return &Cart{Items: []CartItem{}}
}

func updateCart(sessionID string, cart *Cart) {
	cartsMutex.Lock()
	defer cartsMutex.Unlock()

	carts[sessionID] = cart
}

func (c *Cart) calculateTotals() {
	var total, count int
	for _, item := range c.Items {
		total += item.Price * item.Quantity
		count += item.Quantity
	}
	c.Total = total
	c.Count = count
}
