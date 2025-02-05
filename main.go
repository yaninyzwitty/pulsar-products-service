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

	"github.com/yaninyzwitty/pulsar-outbox-products-service/controller"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/database"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/helpers"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/pkg"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/pulsar"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/router"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/sonyflake"
)

var (
	cfg      pkg.Config
	password string
	token    string
)

func main() {
	// Load configuration
	file, err := os.Open("my_config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := cfg.LoadConfig(file); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// if err := godotenv.Load(); err != nil {
	// 	slog.Error("failed to load .env", "error", err)
	// 	os.Exit(1)
	// }
	password = os.Getenv("DB_PASSWORD")
	token = os.Getenv("PULSAR_TOKEN")

	// Initialize dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbConfig := database.DbConfig{
		Host:     cfg.Database.Host,
		Port:     5432,
		User:     cfg.Database.User,
		Password: password,
		DbName:   cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
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
		URI:       cfg.Pulsar.URI,
		Token:     token,
		TopicName: cfg.Pulsar.TopicName,
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
	router := router.NewRouter(productController)

	// Start the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("SERVER starting", "port", cfg.Server.Port)
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
