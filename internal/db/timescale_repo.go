package db

import (
	"blocsy/internal/types"
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

const (
	swapLogTable = "swap_log"
	blocksTable  = "processed_block"
)

type TimescaleRepository struct {
	db *sqlx.DB
}

func NewTimescaleRepository(db *sqlx.DB) *TimescaleRepository {
	return &TimescaleRepository{
		db: db,
	}
}

//=============================================== Block Table Functions  ===============================================

func (repo *TimescaleRepository) MarkBlockProcessed(ctx context.Context, blockNumber int) error {
	var query = fmt.Sprintf(`INSERT INTO "%s" ("blockNumber") VALUES ($1) ON CONFLICT DO NOTHING;`, blocksTable)

	if _, err := repo.db.ExecContext(ctx, query, blockNumber); err != nil {
		return fmt.Errorf("cannot mark block as processed: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) FindMissingBlocks(ctx context.Context) ([][]int, error) {
	var query = fmt.Sprintf(`SELECT DISTINCT "blockNumber" FROM "%s" ORDER BY "blockNumber";`, blocksTable)

	var blockNumbers []int
	if err := repo.db.SelectContext(ctx, &blockNumbers, query); err != nil {
		return nil, fmt.Errorf("cannot get block numbers: %w", err)
	}

	if len(blockNumbers) == 0 {
		return nil, nil
	}

	if blockNumbers[0] == 0 {
		blockNumbers = blockNumbers[1:]
	}

	var missingBlockRanges [][]int
	for i := 1; i < len(blockNumbers); i++ {
		gap := blockNumbers[i] - blockNumbers[i-1] - 1
		if gap > 0 {
			missingBlockRanges = append(missingBlockRanges, []int{
				blockNumbers[i-1] + 1,
				blockNumbers[i] - 1,
			})
		}
	}

	return missingBlockRanges, nil
}

//=============================================== Swap Table Functions  ================================================

func (repo *TimescaleRepository) InsertSwaps(ctx context.Context, swaps []types.SwapLog) error {
	if len(swaps) == 0 {
		return nil
	}

	columns := []string{
		`"id"`,
		`"wallet"`,
		`"source"`,
		`"blockNumber"`,
		`"timestamp"`,
		`"amountOut"`,
		`"amountIn"`,
		`"action"`,
		`"pair"`,
		`"token"`,
		`"processed"`,
	}

	query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, swapLogTable, strings.Join(columns, ", "))

	valueStrings := []string{}
	valueArgs := []interface{}{}

	for i, swap := range swaps {
		base := i*len(columns) + 1
		placeholders := []string{}
		for j := 0; j < len(columns); j++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", base+j))
		}
		valueStrings = append(valueStrings, "("+strings.Join(placeholders, ", ")+")")

		valueArgs = append(valueArgs,
			swap.ID,
			swap.Wallet,
			swap.Source,
			swap.BlockNumber,
			swap.Timestamp.UTC(),
			swap.AmountOut,
			swap.AmountIn,
			swap.Action,
			swap.Pair,
			swap.Token,
			swap.Processed,
		)
	}

	query += strings.Join(valueStrings, ", ") + fmt.Sprintf(` ON CONFLICT (id, pair, action, "amountOut", "amountIn", timestamp, "blockNumber") DO NOTHING;`)

	_, err := repo.db.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("cannot insert swaps batch: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) DeleteSwapsUsingTx(ctx context.Context, signature string) error {
	var query = fmt.Sprintf(`DELETE FROM "%s" WHERE id = '%s'`, swapLogTable, signature)
	log.Println("Deleting swaps using tx: ", signature, query)

	if _, err := repo.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("cannot delete swap: %w", err)
	}

	return nil

}

func (repo *TimescaleRepository) GetAllWalletSwaps(ctx context.Context, wallet string) ([]types.SwapLog, error) {

	var query = fmt.Sprintf(`SELECT * FROM "%s" WHERE wallet = $1;`, swapLogTable)

	var swaps []types.SwapLog

	if err := repo.db.SelectContext(ctx, &swaps, query, wallet); err != nil {
		return nil, fmt.Errorf("cannot get all wallet swaps: %w", err)
	}

	return swaps, nil
}

func (repo *TimescaleRepository) GetSwapsOnDate(ctx context.Context, wallet string, startDate time.Time) ([]types.SwapLog, error) {
	formattedStartDate := startDate.Format("2006-01-02")

	var query = fmt.Sprintf(`
		SELECT * FROM "%s" 
		WHERE "wallet" = $1 
		AND DATE(timestamp) >= $2
		ORDER BY "timestamp" ASC;
	`, swapLogTable)

	var swaps []types.SwapLog
	if err := repo.db.SelectContext(ctx, &swaps, query, wallet, formattedStartDate); err != nil {
		return nil, fmt.Errorf("cannot get swaps: %w", err)
	}

	return swaps, nil
}

func (repo *TimescaleRepository) FindSwap(ctx context.Context, timestamp int64, token string, amount float64) (*types.SwapLog, error) {
	var query = fmt.Sprintf(`SELECT * FROM "%s" 
WHERE token = ?
AND ABS("amountOut"-?) <= 0.001
AND ABS(EXTRACT(EPOCH FROM timestamp) - ?) <= 3 
AND EXTRACT(EPOCH FROM timestamp) < ?
ORDER BY timestamp DESC LIMIT 1;`, swapLogTable)

	query, args, err := sqlx.In(query, token, amount, timestamp, timestamp)
	if err != nil {
		return nil, fmt.Errorf("cannot build query: %w", err)
	}

	query = repo.db.Rebind(query)

	var swap types.SwapLog
	err = repo.db.GetContext(ctx, &swap, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			// No swap found
			return nil, nil
		}
		return nil, fmt.Errorf("cannot get swap: %w", err)
	}

	return &swap, nil
}

func (repo *TimescaleRepository) FindFirstTokenSwaps(ctx context.Context, token string) ([]types.SwapLog, error) {
	var query = fmt.Sprintf(`SELECT * FROM "%s" 
WHERE token = $1
ORDER BY timestamp ASC 
LIMIT 100;`, swapLogTable)

	var swaps []types.SwapLog
	if err := repo.db.SelectContext(ctx, &swaps, query, token); err != nil {
		return nil, fmt.Errorf("cannot get swaps: %w", err)
	}

	return swaps, nil

}

func (repo *TimescaleRepository) FindLatestSwap(ctx context.Context, pair string) ([]types.SwapLog, error) {
	var query = fmt.Sprintf(`SELECT * FROM "%s" 
WHERE pair = $1
ORDER BY timestamp DESC
LIMIT 1;`, swapLogTable)
	var swaps []types.SwapLog
	if err := repo.db.SelectContext(ctx, &swaps, query, pair); err != nil {
		return nil, fmt.Errorf("cannot get swaps: %w", err)
	}

	return swaps, nil

}

func (repo *TimescaleRepository) FindWalletTokenHoldings(ctx context.Context, token string, wallet string) (float64, error) {
	var query = fmt.Sprintf(`
		SELECT COALESCE(SUM(
			CASE 
				WHEN action = 'BUY' THEN "amountIn" 
				WHEN action = 'SELL' THEN -"amountOut" 
				ELSE 0 
			END
		), 0) as totalTokens
		FROM "%s"
		WHERE token = $1
		AND wallet = $2;
	`, swapLogTable)

	var totalTokens float64
	if err := repo.db.GetContext(ctx, &totalTokens, query, token, wallet); err != nil {
		return 0, fmt.Errorf("cannot get total tokens: %w", err)
	}
	return totalTokens, nil
}

//=============================================== Create Tables  =======================================================

func CreateProcessedBlocksTable(ctx context.Context, db *sqlx.DB) {
	var query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
    "blockNumber" INT NOT NULL DEFAULT 0,
    PRIMARY KEY ("blockNumber")
);`, blocksTable)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

}

func CreateSwapsTable(ctx context.Context, db *sqlx.DB) {
	var query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
    "id" TEXT NOT NULL,
    "wallet" TEXT NOT NULL,
    "source" TEXT NOT NULL,
    "blockNumber" INT NOT NULL DEFAULT 0,
    "timestamp" TIMESTAMP NOT NULL,
    "amountOut" DOUBLE PRECISION NOT NULL DEFAULT 0, 
    "amountIn" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "action" TEXT,
    "pair" TEXT NOT NULL,
    "token" TEXT NOT NULL,
    "processed" BOOLEAN DEFAULT FALSE NOT NULL,
    PRIMARY KEY (id,pair,action,"amountIn","amountOut","blockNumber",timestamp)
);`, swapLogTable)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

	ConvertHyperTable(ctx, db, swapLogTable)
}

func ConvertHyperTable(ctx context.Context, db *sqlx.DB, tableName string) {
	query := fmt.Sprintf(`SELECT create_hypertable('%s', 'timestamp');`, tableName)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Printf("Error converting table to hypertable: %v", err)
	}
}
