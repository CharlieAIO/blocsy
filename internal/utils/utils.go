package utils

import (
	"blocsy/internal/db"
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func LoadEnvironment() {
	ENV := os.Getenv("ENV")
	if ENV != "" {
		return
	}

	APP_ENV := strings.ToUpper(os.Getenv("APP_ENV"))
	var err error
	if APP_ENV == "PRODUCTION" {
		err = godotenv.Load(".env")
	} else if APP_ENV == "DEVELOPMENT" {
		log.Println("Loading development environment")
		err = godotenv.Load(".env.development")
	}

	if err != nil {
		log.Println("Error loading .env file", err)
	}
}

func GetMongoConnection(ctx context.Context) (*mongo.Client, error) {
	uri := os.Getenv("MONGO_CONN_STRING")

	if uri == "" {
		log.Fatalf("MONGO_CONN_STRING is not set")
	}

	opts := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetMaxPoolSize(300)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return client, nil
}

func GetDBConnection(ctx context.Context) (*sqlx.DB, error) {
	conn := os.Getenv("POSTGRES_CONN_STRING")
	if conn == "" {
		return nil, fmt.Errorf("POSTGRES_CONN_STRING is not set")
	}

	serviceName := os.Getenv("POSTGRES_SERVICE_NAME")
	if serviceName == "" {
		return nil, fmt.Errorf("POSTGRES_SERVICE_NAME is not set")
	}

	dbConn, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, fmt.Errorf("failed to open Postgres: %w", err)
	}

	dbx := sqlx.NewDb(dbConn, "postgres")
	dbx.SetMaxOpenConns(400)
	dbx.SetMaxIdleConns(100)
	dbx.SetConnMaxLifetime(5 * time.Minute)

	if err := dbx.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Postgres: %w", err)
	}

	if _, err := dbx.ExecContext(ctx, fmt.Sprintf("SET application_name = '%s'", serviceName)); err != nil {
		return nil, fmt.Errorf("failed to set application name: %w", err)
	}

	db.CreateSwapsTable(ctx, dbx)
	db.CreateProcessedBlocksTable(ctx, dbx)
	db.CreateTokenTable(ctx, dbx)
	db.CreatePairTable(ctx, dbx)
	return dbx, nil
}
