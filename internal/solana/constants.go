package solana

const (
	JUPITER_V6_AGGREGATOR = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"

	PUMPFUN = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"

	PUMPFUN_AMM = "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"

	METEORA_DLMM_PROGRAM  = "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"
	METEORA_POOLS_PROGRAM = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"

	RAYDIUM_AMM_ROUTING      = "routeUGWgWzqBWFcrCfv8tritsqukccJPu3q5GPP3xS"
	RAYDIUM_LIQ_POOL_V4      = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	RAYDIUM_CONCENTRATED_LIQ = "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"
	RAYDIUM_CPMM             = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"

	LIFINITY_SWAP_V2 = "2wT8Yq49kHgDzXuPxZSaeLaH1qbmGXtEyPy64bL7aD3c"

	FLUXBEAM_PROGRAM = "FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X"

	ORCA_WHIRL_PROGRAM_ID = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"
	ORCA_SWAP_V2          = "MfDuWeqSHEqTFVYZ7LoexgAK9dxk7cy4DFJWjWMGVWa"
	ORCA_SWAP             = "DjVE6JNiYqPL2QXyCUUh8rNjHrbz9hXHNYt99MQ59qw1"

	PHOENIX = "PhoeNiXZ8ByJGLkxNfZRnkUfjvmuYqLR89jjFHGqdXY"

	TOKEN_PROGRAM            = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	ASSOCIATED_TOKEN_PROGRAM = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	SYSTEM_PROGRAM           = "11111111111111111111111111111111"
	METAPLEX_TOKEN_METDATA   = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
)

var QuoteTokens = map[string]string{
	"So11111111111111111111111111111111111111112":  "SOL",
	"Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": "USDT",
	"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": "USDC",
}

var IgnoreToUsers = map[string]bool{
	"CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM": true,
	"CpoD6tWAsMDeyvVG2q2rD1JbDY6d4AujnvAn2NdrhZV2": true,
	"JD25qVdtd65FoiXNmR89JjmoJdYk9sjYQeSTZAALFiMy": true,
	"GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL": true,
	"45ruCyfdRkWpRNGEqWzjCiXRHkZs8WXCLQ67Pnpye7Hp": true,
	"EQ4iGa81Uyp8WuqZ3yXQGqVH9jWfKT9nubUUGrQmRL3a": true,
}

var IgnoreTokens = map[string]bool{
	"J1toso1uCk3RLmjorhTtrVwY9HJ7X8V9yYac6Y7kGCPn": true, //JitoSOL
	"mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So":  true, //mSOL
	"27G8MtK7VtTcCHkpASjSDdkWWYfoqT6ggEuKidVJidD4": true, //JLP
}

var Programs = map[string]string{
	JUPITER_V6_AGGREGATOR:    "JUPITER_V6_AGGREGATOR",
	PUMPFUN:                  "PUMPFUN",
	PUMPFUN_AMM:              "PUMPFUN_AMM",
	METEORA_DLMM_PROGRAM:     "METEORA_DLMM_PROGRAM",
	METEORA_POOLS_PROGRAM:    "METEORA_POOLS_PROGRAM",
	RAYDIUM_AMM_ROUTING:      "RAYDIUM_AMM_ROUTING",
	RAYDIUM_LIQ_POOL_V4:      "RAYDIUM_LIQ_POOL_V4",
	RAYDIUM_CPMM:             "RAYDIUM_CPMM",
	RAYDIUM_CONCENTRATED_LIQ: "RAYDIUM_CONCENTRATED_LIQ",
	LIFINITY_SWAP_V2:         "LIFINITY_SWAP_V2",
	FLUXBEAM_PROGRAM:         "FLUXBEAM_PROGRAM",
	ORCA_WHIRL_PROGRAM_ID:    "ORCA_WHIRL_PROGRAM_ID",
	ORCA_SWAP_V2:             "ORCA_SWAP_V2",
	ORCA_SWAP:                "ORCA_SWAP",
	PHOENIX:                  "PHOENIX",
}

var IgnorePrograms = map[string]bool{
	PHOENIX:                  true,
	LIFINITY_SWAP_V2:         true,
	ORCA_SWAP:                true,
	ORCA_SWAP_V2:             true,
	RAYDIUM_CONCENTRATED_LIQ: true,
	RAYDIUM_CPMM:             true,
	METEORA_POOLS_PROGRAM:    true,
}
