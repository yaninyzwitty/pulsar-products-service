package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/controller"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/database"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/helpers"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/pulsar"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/router"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/sonyflake"
)

var (
	password  string
	token     string
	host      = "ep-dark-unit-a2z9twbi-pooler.eu-central-1.aws.neon.tech"
	dbName    = "neondb"
	SSLMode   = "require"
	pulsarURI = "pulsar+ssl://pulsar-aws-eucentral1.streaming.datastax.com:6651"
	TopicName = "persistent://witty-cluster/default/products_topic"
	port      = 3000
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}
	password = helpers.GetEnvOrDefault("DB_PASSWORD", "")
	token = helpers.GetEnvOrDefault("PULSAR_TOKEN", "")

	// Initialize dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbConfig := database.DbConfig{
		Host:     host,
		Port:     5432,
		User:     "neondb_owner",
		Password: password,
		DbName:   dbName, // Fixed typo (dnName → dbName)
		SSLMode:  SSLMode,
		MaxConn:  500,
	}

	pool, err := dbConfig.NewPgxPool(ctx, 30)
	if err != nil {
		slog.Error("failed to create pgx pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := dbConfig.Ping(ctx, pool, 30); err != nil {
		slog.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

	if err := sonyflake.InitSonyFlake(); err != nil {
		slog.Error("failed to init sonyflake", "error", err)
		os.Exit(1)
	}

	pulsarCfg := pulsar.PulsarConfig{
		URI:       pulsarURI, // Used hardcoded URI instead of uninitialized cfg.Pulsar.URI
		Token:     token,
		TopicName: TopicName, // Used hardcoded topic instead of uninitialized cfg.Pulsar.TopicName
	}

	pulsarClient, err := pulsarCfg.CreatePulsarConnection(ctx)
	if err != nil {
		slog.Error("failed to create pulsar connection", "error", err)
		os.Exit(1)
	}
	defer pulsarClient.Close()

	producer, err := pulsarCfg.CreatePulsarProducer(ctx, pulsarClient)
	if err != nil {
		slog.Error("failed to create pulsar producer", "error", err)
		os.Exit(1)
	}
	defer producer.Close()

	productController := controller.NewProductsController(pool)
	appRouter := router.NewRouter(productController) // Renamed to avoid conflict

	// Start the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port), // This assumes port is set elsewhere
		Handler: appRouter,
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("SERVER starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Process messages in a separate Goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := helpers.ProcessMessages(context.Background(), pool, producer); err != nil {
					slog.Error("failed to process messages", "error", err)
				}
			case <-stopCh:
				return
			}
		}
	}()

	// Graceful shutdown
	<-stopCh
	slog.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
	} else {
		slog.Info("server stopped gracefully")
	}
}
