package solana

import (
	"blocsy/internal/types"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/near/borsh-go"

	"github.com/blocto/solana-go-sdk/client"
	solClient "github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/metaplex/tokenmeta"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/rpc"
)

var (
	METAPLEX_PROGRAM_ID   = common.PublicKeyFromString("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	TOKEN_2022_PROGRAM_ID = common.PublicKeyFromString("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
	tokenListURL          = "https://token.jup.ag/strict"
)

const (
	accountTypeMint            = 1
	extensionTypeTokenMetadata = 19
	baseAccountLength          = 165
)

type TokenMetadataExtension struct {
	UpdateAuthority    common.PublicKey
	Mint               common.PublicKey
	Name               string
	Symbol             string
	URI                string
	AdditionalMetadata map[string]string
}

func NewSolanaService(ctx context.Context) *SolanaService {
	endpoint := os.Getenv("SOL_HTTPS")

	solCli := solClient.NewClient(endpoint)

	if _, err := solCli.GetVersion(ctx); err != nil {
		log.Printf("RPC CLIENT ERR | failed to version info, err: %v", err)
	}

	return &SolanaService{
		client:     solCli,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SolanaService) GetMintData(ctx context.Context, address string) token.MintAccount {
	accountInfo := client.AccountInfo{}
	for i := 0; i < 2; i++ {
		var err error
		accountInfo, err = s.GetAccountInfo(ctx, address)
		if err != nil {
			log.Printf("failed to get account info, err: %v", err)
			continue
		}

		if accountInfo.Data != nil {
			mintData, err := token.MintAccountFromData(accountInfo.Data)
			if err != nil {
				log.Printf("failed to get mint account from data, err: %v", err)
				return token.MintAccount{}
			}

			return mintData
		}
	}

	return token.MintAccount{}
}

func (s *SolanaService) GetMetadata(ctx context.Context, mintStr string) (*types.Metadata, error) {
	mint := common.PublicKeyFromString(mintStr)
	var md *types.Metadata
	var err error
	if md, err = s.metaplex(ctx, mint); md != nil {
		return md, err
	}
	if md, err = s.token2022(ctx, mint); md != nil {
		return md, err
	}

	return nil, nil
}

func (s *SolanaService) GetAccountInfo(ctx context.Context, base58Addr string) (solClient.AccountInfo, error) {
	ai, err := s.client.GetAccountInfo(ctx, base58Addr)
	if err != nil {
		return solClient.AccountInfo{}, fmt.Errorf("failed to get account info: %w", err)
	}

	return ai, nil
}

func (s *SolanaService) GetSlot(ctx context.Context) (uint64, error) {
	slot, err := s.client.GetSlot(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get slot: %w", err)
	}

	return slot, nil
}

func (s *SolanaService) GetTokenSupplyAndContext(ctx context.Context, address string) (rpc.ValueWithContext[solClient.TokenAmount], error) {
	ts, err := s.client.GetTokenSupplyAndContext(ctx, address)
	if err != nil {
		return rpc.ValueWithContext[solClient.TokenAmount]{}, fmt.Errorf("failed to get token supply and context: %w", err)
	}

	return ts, nil
}

func (s *SolanaService) metaplex(ctx context.Context, mint common.PublicKey) (*types.Metadata, error) {

	metaPDA, _, err := common.FindProgramAddress(
		[][]byte{[]byte("metadata"), METAPLEX_PROGRAM_ID.Bytes(), mint.Bytes()},
		METAPLEX_PROGRAM_ID,
	)
	if err != nil {
		return nil, err
	}
	acc, err := s.GetAccountInfo(ctx, metaPDA.ToBase58())
	if err != nil || acc.Data == nil {
		return nil, err
	}
	raw, err := decompress(acc.Data)
	if err != nil {
		return nil, err
	}
	mpx, err := tokenmeta.MetadataDeserialize(raw)
	if err != nil {
		return nil, err
	}
	return &types.Metadata{
		Name:   mpx.Data.Name,
		Symbol: mpx.Data.Symbol,
		URI:    mpx.Data.Uri,
	}, nil
}

func (s *SolanaService) token2022(ctx context.Context, mint common.PublicKey) (*types.Metadata, error) {
	acc, err := s.GetAccountInfo(ctx, mint.ToBase58())
	if err != nil || acc.Data == nil || acc.Owner != TOKEN_2022_PROGRAM_ID {
		return nil, err
	}
	metadataExt, err := ParseMetadataExtension(acc.Data)
	if err != nil {
		return nil, err
	}

	return &types.Metadata{
		Name:   metadataExt.Name,
		Symbol: metadataExt.Symbol,
		URI:    metadataExt.URI,
	}, nil

}

func ParseMetadataExtension(data []byte) (*TokenMetadataExtension, error) {
	if len(data) <= baseAccountLength {
		return nil, errors.New("no extensions on this mint")
	}
	if data[baseAccountLength] != accountTypeMint {
		return nil, fmt.Errorf("unexpected account type %d", data[baseAccountLength])
	}

	tlv := data[baseAccountLength+1:]
	i := 0
	for i+4 <= len(tlv) {
		typ := binary.LittleEndian.Uint16(tlv[i : i+2])
		length := binary.LittleEndian.Uint16(tlv[i+2 : i+4])
		if int(length) < 0 || i+4+int(length) > len(tlv) {
			return nil, errors.New("malformed TLV entry")
		}
		if typ == extensionTypeTokenMetadata {
			start := i + 4
			end := start + int(length)
			raw := tlv[start:end]

			var tmp struct {
				UpdateAuthority    [32]byte
				Mint               [32]byte
				Name               string
				Symbol             string
				URI                string
				AdditionalMetadata []struct {
					Key   string
					Value string
				}
			}
			if err := borsh.Deserialize(&tmp, raw); err != nil {
				return nil, fmt.Errorf("borsh decode error: %w", err)
			}

			ext := &TokenMetadataExtension{
				UpdateAuthority:    common.PublicKeyFromBytes(tmp.UpdateAuthority[:]),
				Mint:               common.PublicKeyFromBytes(tmp.Mint[:]),
				Name:               tmp.Name,
				Symbol:             tmp.Symbol,
				URI:                tmp.URI,
				AdditionalMetadata: make(map[string]string, len(tmp.AdditionalMetadata)),
			}
			for _, kv := range tmp.AdditionalMetadata {
				ext.AdditionalMetadata[kv.Key] = kv.Value
			}
			return ext, nil
		}
		i += 4 + int(length)
	}
	return nil, errors.New("token metadata extension not found")
}

func decompress(b []byte) ([]byte, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	log.Printf("Checking compression header: %x", b[:min(4, len(b))])

	if bytes.HasPrefix(b, []byte{0x28, 0xb5, 0x2f, 0xfd}) {
		log.Printf("Detected zstd compression")
		r, err := zstd.NewReader(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer r.Close()
		decoded, err := r.DecodeAll(b, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress zstd data: %w", err)
		}
		log.Printf("Successfully decompressed %d bytes to %d bytes", len(b), len(decoded))
		return decoded, nil
	}

	log.Printf("No compression detected, returning raw data")
	return b, nil
}
