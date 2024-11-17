package types

import (
	"errors"
	"github.com/goccy/go-json"
	jlexer "github.com/mailru/easyjson/jlexer"
)

//easyjson:json
type BlockData struct {
	Transactions []SolanaTx `json:"transactions"`
	Timestamp    int64      `json:"blockTime"`
	Block        uint64     `json:"block"`
}

//easyjson:json
type SolanaBlockTx struct {
	Tx        SolanaTx `json:"tx"`
	Block     uint64   `json:"block"`
	Timestamp int64    `json:"timestamp"`
}

//easyjson:json
type SolanaTx struct {
	Meta        TransactionMeta `json:"meta"`
	Transaction TransactionData `json:"transaction"`
}

//easyjson:json
type TransactionMeta struct {
	Err               TransactionError   `json:"err"`
	Fee               int64              `json:"fee"`
	InnerInstructions []InnerInstruction `json:"innerInstructions,omitempty"`
	LogMessages       []string           `json:"logMessages,omitempty"`
	PostBalances      []int64            `json:"postBalances"`
	PostTokenBalances []TokenBalance     `json:"postTokenBalances,omitempty"`
	PreBalances       []int64            `json:"preBalances"`
	PreTokenBalances  []TokenBalance     `json:"preTokenBalances,omitempty"`
	LoadedAddresses   *struct {
		Readonly []string `json:"readonly"`
		Writable []string `json:"writable"`
	} `json:"loadedAddresses,omitempty"`
}

//easyjson:json
type InnerInstruction struct {
	Index        int           `json:"index"`
	Instructions []Instruction `json:"instructions"`
}

//easyjson:json
type TransactionError struct {
	InstructionError [2]interface{} `json:"InstructionError"`
}

//easyjson:json
type TokenBalance struct {
	AccountIndex  int           `json:"accountIndex"`
	Mint          string        `json:"mint"`
	Owner         string        `json:"owner"`
	ProgramId     string        `json:"programId"`
	UITokenAmount UITokenAmount `json:"uiTokenAmount"`
}

//easyjson:json
type UITokenAmount struct {
	Amount         string  `json:"amount"`
	Decimals       int     `json:"decimals"`
	UiAmount       float64 `json:"uiAmount"`
	UiAmountString string  `json:"uiAmountString"`
}

//easyjson:json
type TransactionStatus struct {
	Err interface{} `json:"Err"`
}

//easyjson:json
type TransactionData struct {
	Message    Message  `json:"message"`
	Signatures []string `json:"signatures"`
}

//easyjson:json
type Message struct {
	AccountKeys         []string             `json:"accountKeys"`
	AddressTableLookups []AddressTableLookup `json:"addressTableLookups"`
	Instructions        []Instruction        `json:"instructions"`
	RecentBlockhash     string               `json:"recentBlockhash"`
	Header              interface{}          `json:"header,omitempty"`
}

//easyjson:json
type AccountKey struct {
	Pubkey   string `json:"pubkey"`
	Signer   bool   `json:"signer"`
	Source   string `json:"source"`
	Writable bool   `json:"writable"`
}

//easyjson:json
type AddressTableLookup struct {
	AccountKey      string `json:"accountKey"`
	ReadonlyIndexes []int  `json:"readonlyIndexes"`
	WritableIndexes []int  `json:"writableIndexes"`
}

//easyjson:json
type Instruction struct {
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
	ProgramIdIndex int    `json:"programIdIndex"`
	//ProgramId      string      `json:"programId"`
	//Parsed  *ParsedData `json:"parsed"`
	//Program string `json:"program"`
}

//easyjson:skip
type ParsedData struct {
	Info ParsedDataInfo `json:"info"`
	Type *string        `json:"type,omitempty"`
}

func (pd *ParsedData) UnmarshalEasyJSON(w *jlexer.Lexer) {
	var parsedStruct struct {
		Info ParsedDataInfo `json:"info"`
		Type *string        `json:"type,omitempty"`
	}

	data := w.Raw()

	if err := json.Unmarshal(data, &parsedStruct); err == nil {
		*pd = ParsedData(parsedStruct)
		return
	}

	var parsedString string
	if err := json.Unmarshal(data, &parsedString); err == nil {
		*pd = ParsedData{}
		return
	}

	w.AddError(errors.New("failed to unmarshal ParsedData"))
}

//easyjson:json
type ParsedDataInfo struct {
	Amount      *string        `json:"amount,omitempty"`
	Destination *string        `json:"destination,omitempty"`
	Source      *string        `json:"source,omitempty"`
	Authority   *string        `json:"authority,omitempty"`
	Mint        *string        `json:"mint,omitempty"`
	Wallet      *string        `json:"wallet,omitempty"`
	TokenAmount *UITokenAmount `json:"tokenAmount,omitempty"`
	Account     *string        `json:"account,omitempty"`
	NewAccount  *string        `json:"newAccount,omitempty"`
	Owner       *string        `json:"owner,omitempty"`
	Lamports    *int64         `json:"lamports,omitempty"`
}

//easyjson:json
type PoolInit struct {
	Nonce          int64 `json:"nonce"`
	OpenTime       int64 `json:"open_time"`
	InitPcAmount   int64 `json:"init_pc_amount"`
	InitCoinAmount int64 `json:"init_coin_amount"`
}

type LogType uint8

const (
	INIT LogType = iota
	DEPOSIT
	WITHDRAW
	SWAP_BASE_IN
	SWAP_BASE_OUT
)

//easyjson:json
type SwapBaseInLog struct {
	LogType    LogType
	AmountIn   uint64
	MinimumOut uint64
	Direction  uint64
	UserSource uint64
	PoolCoin   uint64
	PoolPC     uint64
	OutAmount  uint64
}

//easyjson:json
type SwapBaseOutLog struct {
	LogType    LogType
	MaxIn      uint64
	AmountOut  uint64
	Direction  uint64
	UserSource uint64
	PoolCoin   uint64
	PoolPC     uint64
	DeductIn   uint64
}

//easyjson:json
type RaySwapLog struct {
	LogType    LogType
	AmountIn   uint64
	MinimumOut uint64
	Direction  uint64
	UserSource uint64
	PoolCoin   uint64
	PoolPC     uint64
	OutAmount  uint64
	MaxIn      uint64
	AmountOut  uint64
	DeductIn   uint64
}

type TransactionSource string

//easyjson:json
type SolBalanceDiff struct {
	Mint     string `json:"token" bson:"token"`
	Decimals int    `json:"decimals" bson:"decimals"`
	Amount   string `json:"amount" bson:"amount"`
	Owner    string `json:"owner" bson:"owner"`
}

//easyjson:json
type SolTransfer struct {
	InnerIndex       int
	IxIndex          int
	ToUserAccount    string
	ToTokenAccount   string
	FromUserAccount  string
	FromTokenAccount string

	Mint     string
	Decimals int
	Amount   string

	Type string
}

//easyjson:json
type SolSwapData struct {
	NativeInput  NativeInOutData  `json:"nativeInput"`
	NativeOutput NativeInOutData  `json:"nativeOutput"`
	TokenInputs  []TokenInOutData `json:"tokenInputs"`
	TokenOutputs []TokenInOutData `json:"tokenOutputs"`
	InnerSwaps   []InnerSwap      `json:"innerSwaps"`
}

//easyjson:json
type NativeInOutData struct {
	Account string `json:"account"`
	Amount  string `json:"amount"`
}

//easyjson:json
type TokenInOutData struct {
	UserAccount  string `json:"userAccount"`
	TokenAccount string `json:"tokenAccount"`
	Mint         string `json:"mint"`
	TokenAmount  struct {
		Amount   string `json:"amount"`
		Decimals int    `json:"decimals"`
	}
}

//easyjson:json
type InnerSwap struct {
	TokenInputs  []TokenInOutData `json:"tokenInputs"`
	TokenOutputs []TokenInOutData `json:"tokenOutputs"`
	ProgramInfo  struct {
		Source      string `json:"source"`
		Account     string `json:"account"`
		ProgramName string `json:"programName"`
	}
}

//easyjson:json
type SolSwap struct {
	Pair      string
	Exchange  string
	TokenOut  string
	TokenIn   string
	AmountOut string
	AmountIn  string
	Wallet    string
}

//easyjson:json
type WSBlockMessage struct {
	Result *int `json:"result,omitempty"`
	Params *struct {
		Result *struct {
			Value struct {
				Slot  *int `json:"slot"`
				Block struct {
					BlockTime    *int64     `json:"blockTime,omitempty"`
					ParentSlot   *int       `json:"parentSlot,omitempty"`
					Transactions []SolanaTx `json:"transactions"`
				} `json:"block"`
			} `json:"value"`
		} `json:"result,omitempty"`
	} `json:"params,omitempty"`
}

//easyjson:json
type HTTPBlockMessage struct {
	Result *BlockResult `json:"result,omitempty"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

//easyjson:json
type HTTPTxMessage struct {
	Result SolanaTx `json:"result,omitempty"`
}

//easyjson:json
type BlockResult struct {
	BlockTime    int64      `json:"blockTime"`
	ParentSlot   int        `json:"parentSlot"`
	Transactions []SolanaTx `json:"transactions"`
}
