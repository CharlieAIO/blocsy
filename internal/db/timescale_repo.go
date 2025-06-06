package db

import (
	"blocsy/internal/types"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	swapLogTable = "swap_log"
	blocksTable  = "processed_block"
	tokensTable  = "token"
	pairsTable   = "pair"
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

	query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES`, swapLogTable, strings.Join(columns, ", "))

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

func (repo *TimescaleRepository) GetAllWalletSwaps(ctx context.Context, wallet string, limit int64, offset int64) ([]types.SwapLog, error) {

	var query = fmt.Sprintf(`
		SELECT sl.*, t.symbol AS "tokenSymbol"
		FROM "%s" sl
		JOIN token t ON sl.token = t.address
		WHERE sl.wallet = $1
		ORDER BY sl.timestamp DESC
		LIMIT %d OFFSET %d;`, swapLogTable, limit, offset)

	var swaps []types.SwapLog

	if err := repo.db.SelectContext(ctx, &swaps, query, wallet); err != nil {
		return nil, fmt.Errorf("cannot get all wallet swaps: %w", err)
	}

	return swaps, nil
}

func (repo *TimescaleRepository) GetSwapsOnDate(ctx context.Context, wallet string, startDate time.Time) ([]types.SwapLog, error) {
	formattedStartDate := startDate.Format("2006-01-02")

	var query = fmt.Sprintf(`
		SELECT 
			sl.*, 
			t.symbol AS "tokenSymbol", 
			COALESCE(qt.symbol, 'SOL') AS "quoteTokenSymbol"
		FROM 
			"%s" sl
		JOIN 
			token t ON sl.token = t.address
		LEFT JOIN 
			pair pr ON sl.pair = pr.address
		LEFT JOIN 
			token qt ON pr."quoteToken" = qt.address
		WHERE 
			sl.wallet = $1
			AND DATE(sl.timestamp) >= $2
		ORDER BY sl.timestamp ASC;`, swapLogTable)

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
AND (action = 'BUY' OR action = 'SELL')
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
			    WHEN action = 'TRANSFER' THEN "amountIn" 
				WHEN action = 'SELL' THEN -"amountOut" 
			    WHEN action = 'RECEIVE' THEN -"amountOut" 

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

func (repo *TimescaleRepository) FindTopTraders(ctx context.Context, token string, limit int64) ([]string, error) {
	var query = fmt.Sprintf(`
		WITH trader_pnl AS (
			SELECT 
				wallet,
				SUM(CASE 
					WHEN action = 'BUY' THEN -"amountOut"
					WHEN action = 'SELL' THEN "amountIn"
					WHEN action = 'RECEIVE' THEN "amountIn"
					WHEN action = 'TRANSFER' THEN -"amountOut"
					ELSE 0 
				END) as pnl
			FROM "%s"
			WHERE token = $1
			GROUP BY wallet
		)
		SELECT wallet
		FROM trader_pnl
		ORDER BY pnl DESC
		LIMIT %d;`, swapLogTable, limit)

	var traders []string
	if err := repo.db.SelectContext(ctx, &traders, query, token); err != nil {
		return nil, fmt.Errorf("cannot get top traders: %w", err)
	}

	return traders, nil
}

func (repo *TimescaleRepository) FindTopRecentTokens(ctx context.Context) ([]types.TopRecentToken, error) {
	var query = fmt.Sprintf(`
WITH recent_swaps AS (
  SELECT *
  FROM "%s"
  ORDER BY timestamp DESC
  LIMIT 2000
),
ranked_swaps AS (
  SELECT 
    token,
    timestamp,
    CASE
      WHEN action = 'BUY' THEN "amountOut" / "amountIn"
      WHEN action = 'SELL' THEN "amountIn" / "amountOut"
      ELSE 0
    END AS price,
    ROW_NUMBER() OVER (PARTITION BY token ORDER BY timestamp DESC) as rn
  FROM recent_swaps
),
latest_swaps AS (
  SELECT token, price
  FROM ranked_swaps
  WHERE rn = 1
),
tokens AS (
  SELECT address, supply
  FROM "%s"
  WHERE "createdTimestamp" >= NOW() - INTERVAL '1 hour'
)
SELECT 
  s.token, 
  (s.price * t.supply) as market_cap
FROM latest_swaps s
JOIN tokens t ON s.token = t.address
ORDER BY market_cap DESC;
`, swapLogTable, tokensTable)

	var results []types.TopRecentToken
	if err := repo.db.SelectContext(ctx, &results, query); err != nil {
		return nil, fmt.Errorf("cannot get top traders: %w", err)
	}

	return results, nil
}

func (repo *TimescaleRepository) QueryAll(ctx context.Context, searchQuery string) ([]types.QueryAll, error) {
	var query = fmt.Sprintf(`
WITH pair_res AS (
  SELECT
    pair
  FROM "%s"
  WHERE pair = $1
  LIMIT 1
),
wallet_res AS (
  SELECT
    wallet
  FROM "%s"  
  WHERE wallet = $1
    AND NOT EXISTS (SELECT 1 FROM pair_res)
  LIMIT 1
),
token_res AS (
  SELECT
    address  AS token,
    name,
    symbol
  FROM "%s"       
  WHERE name    = $1
     OR symbol  = $1
     OR address = $1
  LIMIT 1
)

SELECT
  'pair'      AS source,
  NULL::text  AS wallet,
  NULL::text  AS token,
  NULL::text  AS name,
  NULL::text  AS symbol,
  pair        AS pair
FROM pair_res

UNION ALL

SELECT
  'wallet'    AS source,
  wallet      AS wallet,
  NULL::text  AS token,
  NULL::text  AS name,
  NULL::text  AS symbol,
  NULL::text  AS pair
FROM wallet_res

UNION ALL

SELECT
  'token'     AS source,
  NULL::text  AS wallet,
  token       AS token,
  name        AS name,
  symbol      AS symbol,
  NULL::text  AS pair
FROM token_res;
`, swapLogTable, swapLogTable, tokensTable)

	var queryAll []types.QueryAll
	if err := repo.db.SelectContext(ctx, &queryAll, query, searchQuery); err != nil {
		return nil, fmt.Errorf("cannot search: %w", err)
	}
	return queryAll, nil
}

// =============================================== Token Table Functions  ================================================
func (repo *TimescaleRepository) InsertToken(ctx context.Context, token types.Token) error {
	var query = fmt.Sprintf(`INSERT INTO "%s" ("address", "name", "symbol", "decimals", "supply", "createdBlock", "createdTimestamp", "deployer", "metadata", "network") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, tokensTable)

	var createdBlock interface{}
	if token.CreatedBlock == 0 {
		createdBlock = 0
	} else {
		createdBlock = token.CreatedBlock
	}

	//var createdTimestamp sql.NullTime
	//if token.CreatedTimestamp == 0 {
	//	createdTimestamp = sql.NullTime{Time: time.Unix(0, 0), Valid: true}
	//} else {
	//	createdTimestamp = sql.NullTime{
	//		Time:  time.Unix(int64(token.CreatedTimestamp), 0),
	//		Valid: true,
	//	}
	//}

	var deployer interface{}
	if token.Deployer == nil || *token.Deployer == "" {
		deployer = nil
	} else {
		deployer = *token.Deployer
	}

	var metadata interface{}
	if token.Metadata == nil || *token.Metadata == "" {
		metadata = nil
	} else {
		metadata = *token.Metadata
	}

	if _, err := repo.db.ExecContext(ctx, query, token.Address, token.Name, token.Symbol, token.Decimals, token.Supply, createdBlock, token.CreatedTimestamp, deployer, metadata, token.Network); err != nil {
		return fmt.Errorf("cannot insert token: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) FindToken(ctx context.Context, address string) (*types.Token, error) {
	var query = fmt.Sprintf(`SELECT
    address, 
	name, 
	symbol, 
	decimals, 
	CASE 
		WHEN supply = TRUNC(supply) THEN CAST(supply AS TEXT)
		ELSE TRIM(TRAILING '0' FROM CAST(supply AS TEXT))
	END as supply,
	"createdBlock",
	"createdTimestamp",
	deployer,
	metadata,
	network 
	FROM "%s" WHERE address = $1`, tokensTable)

	var token types.Token
	if err := repo.db.GetContext(ctx, &token, query, address); err != nil {
		return nil, fmt.Errorf("cannot get token: %w", err)
	}

	return &token, nil
}

func (repo *TimescaleRepository) UpdateTokenInfo(ctx context.Context, address string, metadata *types.Metadata) error {
	var query = fmt.Sprintf(`UPDATE "%s" SET name = $1, symbol = $2, metadata = $3 WHERE address = $4`, tokensTable)
	if _, err := repo.db.ExecContext(ctx, query, metadata.Name, metadata.Symbol, metadata.URI, address); err != nil {
		return fmt.Errorf("cannot update token info: %w", err)
	}
	return nil
}

func (repo *TimescaleRepository) UpdateTokenDecimals(ctx context.Context, address string, decimals int) error {
	var query = fmt.Sprintf(`UPDATE "%s" SET decimals = $1 WHERE address = $2`, tokensTable)
	if _, err := repo.db.ExecContext(ctx, query, decimals, address); err != nil {
		return fmt.Errorf("cannot update token decimals: %w", err)
	}
	return nil
}

func (repo *TimescaleRepository) UpdateTokenSupply(ctx context.Context, address string, changeAmount string, action string) error {
	amount, err := strconv.ParseFloat(changeAmount, 64)
	if err != nil {
		return fmt.Errorf("invalid change amount: %w", err)
	}

	var query string
	switch action {
	case "mint":
		query = fmt.Sprintf(`UPDATE "%s" SET supply = supply + $1::float8 WHERE address = $2`, tokensTable)
	case "burn":
		query = fmt.Sprintf(`UPDATE "%s" SET supply = supply - $1::float8 WHERE address = $2`, tokensTable)
	default:
		return fmt.Errorf("invalid action: %s, must be either 'mint' or 'burn'", action)
	}

	if _, err := repo.db.ExecContext(ctx, query, amount, address); err != nil {
		return fmt.Errorf("cannot update token supply: %w", err)
	}

	return nil
}

//=============================================== Pair Table Functions  ================================================

func (repo *TimescaleRepository) InsertPair(ctx context.Context, pair types.Pair) error {
	var createdBlock interface{}
	if pair.CreatedBlock == 0 {
		createdBlock = 0
	} else {
		createdBlock = pair.CreatedBlock
	}

	var query = fmt.Sprintf(`INSERT INTO "%s" ("address", "token", "quoteToken", "createdBlock", "createdTimestamp","exchange", "network") VALUES ($1,$2,$3,$4,$5,$6,$7)`, pairsTable)
	if _, err := repo.db.ExecContext(ctx, query, pair.Address, pair.Token, pair.QuoteToken.Address, createdBlock, pair.CreatedTimestamp, pair.Exchange, pair.Network); err != nil {
		return fmt.Errorf("cannot insert pair: %w", err)
	}

	return nil
}

func (repo *TimescaleRepository) FindPair(ctx context.Context, address string) (*types.Pair, error) {
	var query = fmt.Sprintf(`SELECT * FROM "%s" WHERE address = $1`, pairsTable)

	var pair types.Pair
	if err := repo.db.GetContext(ctx, &pair, query, address); err != nil {
		return nil, fmt.Errorf("cannot get pair: %w", err)
	}

	return &pair, nil
}

func (repo *TimescaleRepository) FindPairsByToken(ctx context.Context, token string) ([]*types.Pair, error) {
	var query = fmt.Sprintf(`SELECT * FROM "%s" WHERE token = $1`, pairsTable)

	var pairs []*types.Pair
	if err := repo.db.SelectContext(ctx, &pairs, query, token); err != nil {
		return nil, fmt.Errorf("cannot get pairs: %w", err)
	}

	return pairs, nil
}

//=============================================== Create Tables  =======================================================

func CreateTokenTable(ctx context.Context, db *sqlx.DB) {
	var query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
    "address" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "symbol" TEXT NOT NULL,
    "decimals" INT NOT NULL,
    "supply" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "createdBlock" INT NOT NULL DEFAULT 0,
    "createdTimestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deployer" TEXT,
    "metadata" TEXT,
    "network" TEXT NOT NULL,
    PRIMARY KEY ("address")
);`, tokensTable)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

}

func CreatePairTable(ctx context.Context, db *sqlx.DB) {
	var query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
    "address" TEXT NOT NULL,
    "token" TEXT NOT NULL,
    "quoteToken" TEXT NOT NULL,
    "createdBlock" INT NOT NULL DEFAULT 0,
    "createdTimestamp" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "exchange" TEXT NOT NULL,
    "network" TEXT NOT NULL,
    PRIMARY KEY ("address")
);`, pairsTable)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Fatalf("Error creating table: %v", err)
	}

}

func CreateProcessedBlocksTable(ctx context.Context, db *sqlx.DB) {
	var query = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (
    "blockNumber" INT NOT NULL DEFAULT 0,
	"solUsdPrice" DOUBLE PRECISION NOT NULL DEFAULT 0,
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

	// Create indexes on the table so that queries are faster
	//indexes := []string{
	//	`("timestamp" DESC)`,
	//	`("pair","timestamp" DESC)`,
	//	`("token","timestamp" DESC)`,
	//	`("wallet")`,
	//	`("token")`,
	//	`("wallet","pair")`,
	//	`("wallet","token")`,
	//}
	//
	//for _, index := range indexes {
	//	_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE INDEX ON "%s" %s;`, swapLogTable, index))
	//	if err != nil {
	//		log.Printf("Error creating index '%s': %v", index, err)
	//	}
	//}
}

func ConvertHyperTable(ctx context.Context, db *sqlx.DB, tableName string) {
	query := fmt.Sprintf(`SELECT create_hypertable('%s', 'timestamp');`, tableName)

	if _, err := db.ExecContext(ctx, query); err != nil {
		log.Printf("Error converting table to hypertable: %v", err)
	}
}
