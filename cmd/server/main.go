package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/database"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/pkg"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/pulsar"
	"github.com/yaninyzwitty/pulsar-outbox-products-service/sonyflake"
)

var (
	cfg      pkg.Config
	password string
	token    string
)

func main() {

	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := cfg.LoadConfig(file); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}
	if s := os.Getenv("DB_PASSWORD"); s != "" {
		password = s

	}

	if t := os.Getenv("PULSAR_TOKEN"); t != "" {
		token = t
	}

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

	err = sonyflake.InitSonyFlake()
	if err != nil {
		slog.Error("failed to init sonyflake", "error", err)
		os.Exit(1)
	}

	sonyflakId, err := sonyflake.GenerateID()
	if err != nil {
		slog.Error("failed to generate sonyflake id", "error", err)
		os.Exit(1)
	}

	slog.Info("sonyflake id", "id", sonyflakId)

	if err := dbConfig.Ping(ctx, pool, 30); err != nil {
		slog.Error("failed to ping db", "error", err)
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

}
