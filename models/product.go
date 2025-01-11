package models

import "time"

// Product represents an item in an e-commerce or inventory system.
type Product struct {
	ID          uint64    `json:"id"`          // Unique identifier for the product
	Name        string    `json:"name"`        // Name of the product
	Description string    `json:"description"` // Detailed description of the product
	Price       float64   `json:"price"`       // Price of the product
	Stock       int       `json:"stock"`       // Number of items available in stock
	Category    string    `json:"category"`    // Category the product belongs to
	CreatedAt   time.Time `json:"created_at"`  // Timestamp for when the product was created
	UpdatedAt   time.Time `json:"updated_at"`  // Timestamp for when the product was last updated
}
