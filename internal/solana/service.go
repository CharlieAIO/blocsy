package solana

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/blocto/solana-go-sdk/client"
	solClient "github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/metaplex/tokenmeta"
	"github.com/blocto/solana-go-sdk/program/token"
	"github.com/blocto/solana-go-sdk/rpc"
)

func NewSolanaService(ctx context.Context) *SolanaService {
	endpoint := os.Getenv("SOL_HTTPS")

	solCli := solClient.NewClient(endpoint)

	if _, err := solCli.GetVersion(ctx); err != nil {
		log.Printf("failed to get version info, err: %v", err)
	}

	return &SolanaService{
		client: solCli,
	}
}

func (s *SolanaService) GetMintData(ctx context.Context, address string) token.MintAccount {
	accountInfo := client.AccountInfo{}
	for i := 0; i < 5; i++ {
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

func (s *SolanaService) GetMetadata(ctx context.Context, token string) (*tokenmeta.Metadata, error) {
	mint := common.PublicKeyFromString(token)
	metadataAccount, err := tokenmeta.GetTokenMetaPubkey(mint)
	if err != nil {
		log.Printf("failed to get metadata account, err: %v", err)
		return nil, fmt.Errorf("failed to get metadata account: %w", err)
	}

	accountInfo, err := s.GetAccountInfo(ctx, metadataAccount.ToBase58())
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	if accountInfo.Data == nil {
		return nil, errors.New("account data is nil")
	}

	metadata, err := tokenmeta.MetadataDeserialize(accountInfo.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	return &metadata, nil
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
