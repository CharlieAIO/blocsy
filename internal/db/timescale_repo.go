package db

import (
	"context"
	"defi-intel/internal/types"
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

func (repo *TimescaleRepository) MarkBlockProcessed(ctx context.Context, blockNumber int) error {
	const query = `INSERT INTO "processed_block" ("blockNumber") VALUES ($1) ON CONFLICT DO NOTHING;`

	if _, err := repo.db.ExecContext(ctx, query, blockNumber); err != nil {
		return fmt.Errorf("cannot mark block as processed: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) InsertSwaps(ctx context.Context, swaps []types.SwapLog) error {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	const query = `INSERT INTO "swap_log_new" (id, wallet, network, exchange, "blockNumber", "blockHash", timestamp, type, "amountOut", "amountIn", price, pair, "logIndex", processed)
	VALUES (:id, :wallet, :network, :exchange, :blockNumber, :blockHash, :timestamp, :type, :amountOut, :amountIn, :price, :pair, :logIndex, :processed)
	ON CONFLICT (id, pair, type, "amountOut", "amountIn", timestamp, "blockNumber") DO NOTHING;`

	for _, swap := range swaps {
		if _, err := tx.NamedExecContext(ctx, query, swap); err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("cannot rollback transaction: %w", rbErr)
			}
			return fmt.Errorf("cannot insert swap: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit transaction: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) DeleteSwapsUsingTx(ctx context.Context, signature string) error {
	var query = fmt.Sprintf(`DELETE FROM "swap_log_new" WHERE id = '%s'`, signature)
	log.Println("Deleting swaps using tx: ", signature, query)

	if _, err := repo.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("cannot delete swap: %w", err)
	}

	return nil

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
