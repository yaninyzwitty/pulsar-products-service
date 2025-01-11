package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/helpers"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/models"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/sonyflake"
)

type ProductsController interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
}

type productController struct {
	pool *pgxpool.Pool
}

func NewProductsController(pool *pgxpool.Pool) ProductsController {
	return &productController{pool: pool}
}

func (c *productController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	var ctx = r.Context()

	// Decode the request body into the product struct
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		handleError(w, http.StatusBadRequest, fmt.Sprintf("Failed to decode product data: %v", err))
		return
	}

	// Begin a new transaction
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer tx.Rollback(ctx)

	// Generate Snowflake ID for the product
	sfId, err := sonyflake.GenerateID()
	if err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to generate Snowflake ID")
		return
	}

	// Insert product data into the database
	query := `INSERT INTO products (id, name, description, price, stock, category)
				VALUES ($1, $2, $3, $4, $5, $6) 
				RETURNING id, name, description, price, stock, category, created_at, updated_at`
	err = tx.QueryRow(ctx, query, sfId, product.Name, product.Description, product.Price, product.Stock, product.Category).
		Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to insert product: %v", err))
		return
	}

	// Prepare outbox event for message queue
	event := models.Outbox{
		EventType: "product_created",
		Payload: fmt.Sprintf(`{"id": %d, "name": "%s", "description": "%s", "price": %.2f, "stock": %d, "category": "%s"}`,
			product.ID, product.Name, product.Description, product.Price, product.Stock, product.Category),
		Status: "pending",
	}

	// Insert outbox event into the database
	outboxQuery := `INSERT INTO outbox (event_type, payload, status)
					VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, outboxQuery, event.EventType, event.Payload, event.Status)
	if err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to insert outbox event: %v", err))
		return
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to commit transaction: %v", err))
		return
	}

	// Convert and send the response
	if err := helpers.ConvertStructToJson(w, http.StatusOK, product); err != nil {
		handleError(w, http.StatusInternalServerError, err.Error())
	}
}

// handleError is a helper function to send error responses with appropriate status codes
func handleError(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
}
