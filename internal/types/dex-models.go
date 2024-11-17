package types

import (
	"fmt"
	"github.com/blocto/solana-go-sdk/common"
	bin "github.com/gagliardetto/binary"
)

type RaydiumV4Layout struct {
	Status                 bin.Uint64
	Nonce                  bin.Uint64
	MaxOrder               bin.Uint64
	Depth                  bin.Uint64
	BaseDecimal            bin.Uint64
	QuoteDecimal           bin.Uint64
	State                  bin.Uint64
	ResetFlag              bin.Uint64
	MinSize                bin.Uint64
	VolMaxCutRatio         bin.Uint64
	AmountWaveRatio        bin.Uint64
	BaseLotSize            bin.Uint64
	QuoteLotSize           bin.Uint64
	MinPriceMultiplier     bin.Uint64
	MaxPriceMultiplier     bin.Uint64
	SystemDecimalValue     bin.Uint64
	MinSeparateNumerator   bin.Uint64
	MinSeparateDenominator bin.Uint64
	TradeFeeNumerator      bin.Uint64
	TradeFeeDenominator    bin.Uint64
	PnlNumerator           bin.Uint64
	PnlDenominator         bin.Uint64
	SwapFeeNumerator       bin.Uint64
	SwapFeeDenominator     bin.Uint64
	BaseNeedTakePnl        bin.Uint64
	QuoteNeedTakePnl       bin.Uint64
	QuoteTotalPnl          bin.Uint64
	BaseTotalPnl           bin.Uint64
	QuoteTotalDeposited    bin.Uint128
	BaseTotalDeposited     bin.Uint128
	SwapBaseInAmount       bin.Uint128
	SwapQuoteOutAmount     bin.Uint128
	SwapBase2QuoteFee      bin.Uint64
	SwapQuoteInAmount      bin.Uint128
	SwapBaseOutAmount      bin.Uint128
	SwapQuote2BaseFee      bin.Uint64
	BaseVault              common.PublicKey
	QuoteVault             common.PublicKey
	BaseMint               common.PublicKey
	QuoteMint              common.PublicKey
	LpMint                 common.PublicKey
	OpenOrders             common.PublicKey
	MarketID               common.PublicKey
	MarketProgramID        common.PublicKey
	TargetOrders           common.PublicKey
	WithdrawQueue          common.PublicKey
	LpVault                common.PublicKey
	Owner                  common.PublicKey
	LpReserve              bin.Uint64
	Padding                [3]byte `json:"-"`
}

func (m *RaydiumV4Layout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type RewardInfo struct {
	RewardState           uint8            `json:"rewardState"`
	Padding1              [7]byte          `json:"-"` // Padding to align next field to 8 bytes
	OpenTime              uint64           `json:"openTime"`
	EndTime               uint64           `json:"endTime"`
	LastUpdateTime        uint64           `json:"lastUpdateTime"`
	EmissionsPerSecondX64 uint64           `json:"emissionsPerSecondX64"`
	RewardTotalEmissioned uint64           `json:"rewardTotalEmissioned"`
	RewardClaimed         uint64           `json:"rewardClaimed"`
	TokenMint             common.PublicKey `json:"tokenMint"`
	TokenVault            common.PublicKey `json:"tokenVault"`
	Authority             common.PublicKey `json:"authority"`
	RewardGrowthGlobalX64 uint64           `json:"rewardGrowthGlobalX64"`
}

type RaydiumConcentratedLayout struct {
	Bump                      [1]uint8         `json:"bump"`
	AmmConfig                 common.PublicKey `json:"ammConfig"`
	Owner                     common.PublicKey `json:"owner"`
	TokenMint0                common.PublicKey `json:"tokenMint0"`
	TokenMint1                common.PublicKey `json:"tokenMint1"`
	TokenVault0               common.PublicKey `json:"tokenVault0"`
	TokenVault1               common.PublicKey `json:"tokenVault1"`
	ObservationKey            common.PublicKey `json:"observationKey"`
	MintDecimals0             uint8            `json:"mintDecimals0"`
	MintDecimals1             uint8            `json:"mintDecimals1"`
	TickSpacing               uint16           `json:"tickSpacing"`
	Liquidity                 bin.Uint128      `json:"liquidity"`
	SqrtPriceX64              bin.Uint128      `json:"sqrtPriceX64"`
	TickCurrent               int32            `json:"tickCurrent"`
	ObservationIndex          uint16           `json:"observationIndex"`
	ObservationUpdateDuration uint16           `json:"observationUpdateDuration"`
	FeeGrowthGlobal0X64       bin.Uint128      `json:"feeGrowthGlobal0X64"`
	FeeGrowthGlobal1X64       bin.Uint128      `json:"feeGrowthGlobal1X64"`
	ProtocolFeesToken0        uint64           `json:"protocolFeesToken0"`
	ProtocolFeesToken1        uint64           `json:"protocolFeesToken1"`
	SwapInAmountToken0        bin.Uint128      `json:"swapInAmountToken0"`
	SwapOutAmountToken1       bin.Uint128      `json:"swapOutAmountToken1"`
	SwapInAmountToken1        bin.Uint128      `json:"swapInAmountToken1"`
	SwapOutAmountToken0       bin.Uint128      `json:"swapOutAmountToken0"`
	Status                    uint8            `json:"status"`
	Padding                   [7]byte          `json:"padding"`
	RewardInfos               [3]RewardInfo    `json:"rewardInfos"`
	TickArrayBitmap           [16]uint64       `json:"tickArrayBitmap"`
	TotalFeesToken0           uint64           `json:"totalFeesToken0"`
	TotalFeesClaimedToken0    uint64           `json:"totalFeesClaimedToken0"`
	TotalFeesToken1           uint64           `json:"totalFeesToken1"`
	TotalFeesClaimedToken1    uint64           `json:"totalFeesClaimedToken1"`
	FundFeesToken0            uint64           `json:"fundFeesToken0"`
	FundFeesToken1            uint64           `json:"fundFeesToken1"`
	OpenTime                  uint64           `json:"openTime"`
	Padding1                  [25]uint64       `json:"padding1"`
	Padding2                  [32]uint64       `json:"padding2"`
}

func (m *RaydiumConcentratedLayout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type MeteoraLayout struct {
	Padding1                  [8]byte          `json:"padding1"`
	StaticParametersPadding   [32]byte         `json:"staticParametersPadding"`
	VariableParametersPadding [32]byte         `json:"variableParametersPadding"`
	BumpSeed                  [1]byte          `json:"bumpSeed"`
	BinStepSeed               [2]byte          `json:"binStepSeed"`
	PairType                  uint8            `json:"pairType"`
	ActiveId                  int32            `json:"activeId"`
	BinStep                   uint16           `json:"binStep"`
	Status                    uint8            `json:"status"`
	RequireBaseFactorSeed     uint8            `json:"requireBaseFactorSeed"`
	BaseFactorSeed            [2]byte          `json:"baseFactorSeed"`
	Padding2                  [2]byte          `json:"padding2"`
	TokenXMint                common.PublicKey `json:"tokenXMint"`
	TokenYMint                common.PublicKey `json:"tokenYMint"`
	ReserveX                  common.PublicKey `json:"reserveX"`
	ReserveY                  common.PublicKey `json:"reserveY"`
	ProtocolFeePadding        [16]byte         `json:"protocolFeePadding"`
	FeeOwner                  common.PublicKey `json:"feeOwner"`
	RewardInfosPadding        [2][64]byte      `json:"rewardInfosPadding"`
	Oracle                    common.PublicKey `json:"oracle"`
	BinArrayBitmap            [16]uint64       `json:"binArrayBitmap"`
	LastUpdatedAt             int64            `json:"lastUpdatedAt"`
	WhitelistedWallet         common.PublicKey `json:"whitelistedWallet"`
	PreActivationSwapAddress  common.PublicKey `json:"preActivationSwapAddress"`
	BaseKey                   common.PublicKey `json:"baseKey"`
	ActivationSlot            uint64           `json:"activationSlot"`
	PreActivationSlotDuration uint64           `json:"preActivationSlotDuration"`
	Padding3                  [8]byte          `json:"padding3"`
	LockDurationsInSlot       uint64           `json:"lockDurationsInSlot"`
	Creator                   common.PublicKey `json:"creator"`
	Reserved                  [24]byte         `json:"reserved"`
}

func (m *MeteoraLayout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type MeteoraPoolsLayout struct {
	Padding        [8]byte
	LpMint         common.PublicKey `json:"lpMint"`
	TokenAMint     common.PublicKey `json:"tokenAMint"`
	TokenBMint     common.PublicKey `json:"tokenBMint"`
	AVault         common.PublicKey `json:"aVault"`
	BVault         common.PublicKey `json:"bVault"`
	AVaultLp       common.PublicKey `json:"aVaultLp"`
	BVaultLp       common.PublicKey `json:"bVaultLp"`
	AVaultLpBump   uint8            `json:"aVaultLpBump"`
	Enabled        bool             `json:"enabled"`
	Padding1       [6]byte          `json:"padding1"`
	AdminTokenAFee common.PublicKey `json:"adminTokenAFee"`
	AdminTokenBFee common.PublicKey `json:"adminTokenBFee"`
	Admin          common.PublicKey `json:"admin"`
	Padding2       [32]byte         `json:"padding2"` // Add padding to fill the space of PoolFees struct
	PoolType       struct {
		Permissionless interface{} `json:"permissionless"`
	} `json:"poolType"`
	Padding3      [7]byte          `json:"padding3"`
	Stake         common.PublicKey `json:"stake"`
	TotalLockedLp uint64           `json:"totalLockedLp"`
	Padding4      [8]byte          `json:"padding4"`
	Padding5      [128]byte        `json:"padding5"` // Add padding to fill the space of CurveType struct
}

func (m *MeteoraPoolsLayout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type OrcaWhirlpool struct {
	Padding          [8]byte
	WhirlpoolsConfig common.PublicKey // 32
	WhirlpoolBump    [1]uint8         // 1 - 33
	TickSpacing      uint16
	TickSpacingSeed  [2]uint8 // 2 - 37
	// Padding [45]byte
	FeeRate uint16 // 4 - 41
	// Padding         [47]byte
	ProtocolFeeRate uint16 // 4 - 45
	// Padding   [49]byte
	Liquidity bin.Uint128 // 16 - 57
	// Padding   [65]byte
	SqrtPrice bin.Uint128 // 16 - 69
	// Padding          [81]byte
	TickCurrentIndex int32 // 4 - 73
	// Padding          [85]byte
	ProtocolFeeOwedA uint64 // 8 - 81
	// Padding          [93]byte
	ProtocolFeeOwedB uint64 // 8 - 89
	// Padding                    [101]byte
	TokenMintA                 common.PublicKey
	TokenVaultA                common.PublicKey
	FeeGrowthGlobalA           bin.Uint128
	TokenMintB                 common.PublicKey
	TokenVaultB                common.PublicKey
	FeeGrowthGlobalB           bin.Uint128
	RewardLastUpdatedTimestamp uint64
	RewardInfos                [3]struct {
		Mint                  common.PublicKey
		Vault                 common.PublicKey
		Authority             common.PublicKey
		EmissionsPerSecondX64 bin.Uint128
		GrowthGlobalX64       bin.Uint128
	}
}

func (m *OrcaWhirlpool) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type FluxBeamPool struct {
	Version                     uint8            `json:"version"`
	IsInitialized               bool             `json:"isInitialized"`
	BumpSeed                    uint8            `json:"bumpSeed"`
	PoolTokenProgramId          common.PublicKey `json:"poolTokenProgramId"`
	TokenAccountA               common.PublicKey `json:"tokenAccountA"`
	TokenAccountB               common.PublicKey `json:"tokenAccountB"`
	TokenPool                   common.PublicKey `json:"tokenPool"`
	MintA                       common.PublicKey `json:"mintA"`
	MintB                       common.PublicKey `json:"mintB"`
	FeeAccount                  common.PublicKey `json:"feeAccount"`
	TradeFeeNumerator           uint64           `json:"tradeFeeNumerator"`
	TradeFeeDenominator         uint64           `json:"tradeFeeDenominator"`
	OwnerTradeFeeNumerator      uint64           `json:"ownerTradeFeeNumerator"`
	OwnerTradeFeeDenominator    uint64           `json:"ownerTradeFeeDenominator"`
	OwnerWithdrawFeeNumerator   uint64           `json:"ownerWithdrawFeeNumerator"`
	OwnerWithdrawFeeDenominator uint64           `json:"ownerWithdrawFeeDenominator"`
	HostFeeNumerator            uint64           `json:"hostFeeNumerator"`
	HostFeeDenominator          uint64           `json:"hostFeeDenominator"`
	CurveType                   uint8            `json:"curveType"`
	CurveParameters             [32]byte         `json:"curveParameters"`
}

func (m *FluxBeamPool) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type MarketSizeParams struct {
	BidsSize uint64 `json:"bidsSize"`
	AsksSize uint64 `json:"asksSize"`
	NumSeats uint64 `json:"numSeats"`
}

type TokenParams struct {
	Decimals  uint8            `json:"decimals"`
	VaultBump uint8            `json:"vaultBump"`
	MintKey   common.PublicKey `json:"mintKey"`
	VaultKey  common.PublicKey `json:"vaultKey"`
}

type Header struct {
	Discriminator                   [8]byte          `json:"discriminator"`
	Status                          uint8            `json:"status"`
	MarketSizeParams                MarketSizeParams `json:"marketSizeParams"`
	TokenParams                     TokenParams      `json:"tokenParams"`
	BaseLotSize                     uint64           `json:"baseLotSize"`
	QuoteLotSize                    uint64           `json:"quoteLotSize"`
	TickSizeInQuoteAtomsPerBaseUnit uint64           `json:"tickSizeInQuoteAtomsPerBaseUnit"`
	Authority                       common.PublicKey `json:"authority"`
	FeeRecipient                    common.PublicKey `json:"feeRecipient"`
	MarketSequenceNumber            uint64           `json:"marketSequenceNumber"`
	Successor                       common.PublicKey `json:"successor"`
	RawBaseUnitsPerBaseUnit         uint64           `json:"rawBaseUnitsPerBaseUnit"`
	Padding1                        uint8            `json:"padding1"`
	Padding2                        [32]uint8        `json:"padding2"`
}

type PhoenixMarketLayout struct {
	Header                      Header `json:"header"`
	BaseLotsPerBaseUnit         uint64 `json:"baseLotsPerBaseUnit"`
	QuoteLotsPerBaseUnitPerTick uint64 `json:"quoteLotsPerBaseUnitPerTick"`
	OrderSequenceNumber         uint64 `json:"orderSequenceNumber"`
	TakerFeeBps                 uint16 `json:"takerFeeBps"`
	CollectedQuoteLotFees       uint64 `json:"collectedQuoteLotFees"`
	UnclaimedQuoteLotFees       uint64 `json:"unclaimedQuoteLotFees"`
	// Assuming you have a fixed-size order book (modify as necessary)
	Bids [1024]OrderEntry `json:"bids"`
	Asks [1024]OrderEntry `json:"asks"`
	// Additional fields, like trader mappings, can be added as needed
}

type OrderEntry struct {
	PriceInTicks                    uint64 `json:"priceInTicks"`
	OrderSequenceNumber             uint64 `json:"orderSequenceNumber"`
	TraderIndex                     uint64 `json:"traderIndex"`
	NumBaseLots                     uint64 `json:"numBaseLots"`
	LastValidSlot                   uint64 `json:"lastValidSlot"`
	LastValidUnixTimestampInSeconds uint64 `json:"lastValidUnixTimestampInSeconds"`
}

func (m *PhoenixMarketLayout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type AmmFees struct {
	TradeFeeNumerator           uint64 `json:"tradeFeeNumerator"`
	TradeFeeDenominator         uint64 `json:"tradeFeeDenominator"`
	OwnerTradeFeeNumerator      uint64 `json:"ownerTradeFeeNumerator"`
	OwnerTradeFeeDenominator    uint64 `json:"ownerTradeFeeDenominator"`
	OwnerWithdrawFeeNumerator   uint64 `json:"ownerWithdrawFeeNumerator"`
	OwnerWithdrawFeeDenominator uint64 `json:"ownerWithdrawFeeDenominator"`
	HostFeeNumerator            uint64 `json:"hostFeeNumerator"`
	HostFeeDenominator          uint64 `json:"hostFeeDenominator"`
}

type AmmCurve struct {
	CurveType       uint8   `json:"curveType"`
	Padding1        [7]byte // Padding to align next field to 8 bytes
	CurveParameters uint64  `json:"curveParameters"`
}

type AmmConfig struct {
	LastPrice                uint64 `json:"lastPrice"`
	LastBalancedPrice        uint64 `json:"lastBalancedPrice"`
	ConfigDenominator        uint64 `json:"configDenominator"`
	VolumeX                  uint64 `json:"volumeX"`
	VolumeY                  uint64 `json:"volumeY"`
	VolumeXInY               uint64 `json:"volumeXInY"`
	DepositCap               uint64 `json:"depositCap"`
	RegressionTarget         uint64 `json:"regressionTarget"`
	OracleType               uint64 `json:"oracleType"`
	OracleStatus             uint64 `json:"oracleStatus"`
	OracleMainSlotLimit      uint64 `json:"oracleMainSlotLimit"`
	OracleSubConfidenceLimit uint64 `json:"oracleSubConfidenceLimit"`
	OracleSubSlotLimit       uint64 `json:"oracleSubSlotLimit"`
	OraclePcConfidenceLimit  uint64 `json:"oraclePcConfidenceLimit"`
	OraclePcSlotLimit        uint64 `json:"oraclePcSlotLimit"`
	StdSpread                uint64 `json:"stdSpread"`
	StdSpreadBuffer          uint64 `json:"stdSpreadBuffer"`
	SpreadCoefficient        uint64 `json:"spreadCoefficient"`
	PriceBufferCoin          int64  `json:"priceBufferCoin"`
	PriceBufferPc            int64  `json:"priceBufferPc"`
	RebalanceRatio           uint64 `json:"rebalanceRatio"`
	FeeTrade                 uint64 `json:"feeTrade"`
	FeePlatform              uint64 `json:"feePlatform"`
	ConfigTemp3              uint64 `json:"configTemp3"`
	ConfigTemp4              uint64 `json:"configTemp4"`
	ConfigTemp5              uint64 `json:"configTemp5"`
	ConfigTemp6              uint64 `json:"configTemp6"`
	ConfigTemp7              uint64 `json:"configTemp7"`
	ConfigTemp8              uint64 `json:"configTemp8"`
}

type LfinitySwapV2Layout struct {
	InitializerKey                 common.PublicKey `json:"initializerKey"`
	InitializerDepositTokenAccount common.PublicKey `json:"initializerDepositTokenAccount"`
	InitializerReceiveTokenAccount common.PublicKey `json:"initializerReceiveTokenAccount"`
	InitializerAmount              uint64           `json:"initializerAmount"`
	TakerAmount                    uint64           `json:"takerAmount"`
	IsInitialized                  uint8            `json:"isInitialized"`
	BumpSeed                       uint8            `json:"bumpSeed"`
	FreezeTrade                    uint8            `json:"freezeTrade"`
	FreezeDeposit                  uint8            `json:"freezeDeposit"`
	FreezeWithdraw                 uint8            `json:"freezeWithdraw"`
	BaseDecimals                   uint8            `json:"baseDecimals"`
	Padding1                       [3]byte          // Padding to align the next field to 8 bytes
	TokenProgramId                 common.PublicKey `json:"tokenProgramId"`
	TokenAAccount                  common.PublicKey `json:"tokenAAccount"`
	TokenBAccount                  common.PublicKey `json:"tokenBAccount"`
	PoolMint                       common.PublicKey `json:"poolMint"`
	TokenAMint                     common.PublicKey `json:"tokenAMint"`
	TokenBMint                     common.PublicKey `json:"tokenBMint"`
	FeeAccount                     common.PublicKey `json:"feeAccount"`
	OracleMainAccount              common.PublicKey `json:"oracleMainAccount"`
	OracleSubAccount               common.PublicKey `json:"oracleSubAccount"`
	OraclePcAccount                common.PublicKey `json:"oraclePcAccount"`
	Fees                           AmmFees          `json:"fees"`
	Curve                          AmmCurve         `json:"curve"`
	Config                         AmmConfig        `json:"config"`
	AmmPTemp1                      common.PublicKey `json:"ammPTemp1"`
	AmmPTemp2                      common.PublicKey `json:"ammPTemp2"`
	AmmPTemp3                      common.PublicKey `json:"ammPTemp3"`
	AmmPTemp4                      common.PublicKey `json:"ammPTemp4"`
	AmmPTemp5                      common.PublicKey `json:"ammPTemp5"`
}

func (m *LfinitySwapV2Layout) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}

type PumpFunSwap struct {
	Mint                 common.PublicKey `json:"mint"`
	SolAmount            uint64           `json:"solAmount,string"`
	TokenAmount          uint64           `json:"tokenAmount,string"`
	IsBuy                bool             `json:"isBuy"`
	User                 common.PublicKey `json:"user"`
	Timestamp            int64            `json:"timestamp,string"`
	VirtualSolReserves   uint64           `json:"virtualSolReserves,string"`
	VirtualTokenReserves uint64           `json:"virtualTokenReserves,string"`
}

func (m *PumpFunSwap) Decode(in []byte) error {
	decoder := bin.NewBinDecoder(in)
	err := decoder.Decode(&m)
	if err != nil {
		return fmt.Errorf("unpack: %w", err)
	}
	return nil
}
