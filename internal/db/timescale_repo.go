package db

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
)

type TimescaleRepository struct {
	db *sqlx.DB
}

func NewTimescaleRepository(db *sqlx.DB) *TimescaleRepository {
	return &TimescaleRepository{
		db: db,
	}
}

func CreateProcessedBlocksTable(ctx context.Context, db *sqlx.DB) {
	const query = `CREATE TABLE IF NOT EXISTS "processed_block" (
    "blockNumber" INT NOT NULL DEFAULT 0,
    PRIMARY KEY ("blockNumber")
);`

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

}

func CreateSwapsTable(ctx context.Context, db *sqlx.DB) {
	const query = `CREATE TABLE IF NOT EXISTS "swap_log_new" (
    "id" TEXT NOT NULL,
    "wallet" TEXT NOT NULL,
    "network" TEXT NOT NULL,
    "exchange" TEXT NOT NULL,
    "blockNumber" INT NOT NULL DEFAULT 0,
    "blockHash" TEXT NOT NULL DEFAULT '',
    "timestamp" TIMESTAMP NOT NULL,
    "type" TEXT NOT NULL,
    "amountOut" DOUBLE PRECISION NOT NULL DEFAULT 0, 
    "amountIn" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "price" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "pair" TEXT NOT NULL,
    "logIndex" TEXT NOT NULL DEFAULT '',
    "processed" BOOLEAN DEFAULT FALSE NOT NULL,
    PRIMARY KEY (id,pair,type, "amountIn","amountOut","blockNumber","timestamp")
);`

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	ConvertHyperTable(ctx, db, "swap_log_new")
}

func CreateDataPointTable(ctx context.Context, db *sqlx.DB) {
	const query = `CREATE TABLE IF NOT EXISTS "data_points" (
    "pair" TEXT NOT NULL,
    "timestamp" BIGINT NOT NULL,
    "openPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "closePrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "highPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "lowPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "totalVolume" FLOAT NOT NULL DEFAULT 0,
    "buyVolume" FLOAT NOT NULL DEFAULT 0,
    "sellVolume" FLOAT NOT NULL DEFAULT 0,
    "numTxns" INT NOT NULL DEFAULT 0,
    "numBuys" INT NOT NULL DEFAULT 0,
    "numSells" INT NOT NULL DEFAULT 0,
    "uniqueWallets" INT NOT NULL DEFAULT 0,
    "processed" BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY ("pair","timestamp")
);`

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	ConvertHyperTable(ctx, db, "data_points")
}

func ConvertHyperTable(ctx context.Context, db *sqlx.DB, tableName string) {
	query := fmt.Sprintf(`SELECT create_hypertable('%s', 'timestamp');`, tableName)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Printf("Error converting table to hypertable: %v", err)
	}
}
