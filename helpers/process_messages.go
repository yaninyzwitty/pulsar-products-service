package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/models"
)

// ProcessMessages fetches pending messages from the outbox table, publishes them to Pulsar, and marks them as processed.
func ProcessMessages(ctx context.Context, pool *pgxpool.Pool, producer pulsar.Producer) error {
	// Define the query to select pending events
	query := `
		SELECT 
			id, event_type, payload, status, created_at 
		FROM 
			outbox 
		WHERE 
			status = $1
		FOR UPDATE SKIP LOCKED
	`

	// Execute the query to fetch pending messages
	rows, err := pool.Query(ctx, query, "pending")
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()

	// Process each row
	for rows.Next() {
		var outbox models.Outbox
		if err := rows.Scan(&outbox.ID, &outbox.EventType, &outbox.Payload, &outbox.Status, &outbox.CreatedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Begin transaction for updating the message status
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Defer rollback only if transaction fails
		rollback := true
		defer func() {
			if rollback {
				tx.Rollback(ctx)
			}
		}()

		// Unmarshal the payload into a product struct
		var product models.Product
		if err := json.Unmarshal([]byte(outbox.Payload), &product); err != nil {
			log.Printf("Warning: failed to unmarshal payload for message ID %v: %v", outbox.ID, err)
			continue // Skip this message and continue with the next one
		}

		// Marshal the entire outbox message
		messageData, err := json.Marshal(outbox)
		if err != nil {
			log.Printf("Warning: failed to marshal outbox for message ID %v: %v", outbox.ID, err)
			continue
		}

		// Create message channel for async response
		messageChan := make(chan error, 1)

		// Send message asynchronously
		producer.SendAsync(
			ctx,
			&pulsar.ProducerMessage{
				Key:     fmt.Sprintf("%s:%v", outbox.EventType, product.ID),
				Payload: messageData,
			},
			func(msgID pulsar.MessageID, msg *pulsar.ProducerMessage, err error) {
				messageChan <- err
				if err == nil {
					log.Printf("Message ID %v successfully published with Pulsar ID %v", outbox.ID, msgID)
				}
			},
		)

		// Wait for async operation to complete or context to cancel
		select {
		case err := <-messageChan:
			if err != nil {
				log.Printf("Failed to publish message ID %v: %v", outbox.ID, err)
				continue
			}
		case <-ctx.Done():
			return fmt.Errorf("context canceled while publishing message ID %v: %w", outbox.ID, ctx.Err())
		}

		// Update message status to processed
		updateQuery := `
			UPDATE outbox 
			SET status = $1 
			WHERE id = $2
		`

		if _, err = tx.Exec(ctx, updateQuery, "processed", outbox.ID); err != nil {
			return fmt.Errorf("failed to update status for ID %v: %w", outbox.ID, err)
		}

		// Commit the transaction
		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction for ID %v: %w", outbox.ID, err)
		}

		// Mark transaction as successful
		rollback = false
	}

	// Check for errors after iterating through rows
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return nil
}
