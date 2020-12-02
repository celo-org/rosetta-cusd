package services

import (
	"context"
	"log"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// MempoolAPIService implements the server.MempoolAPIServicer interface.
type MempoolAPIService struct {
	config *configuration.Configuration
	client *client.APIClient
}

func NewMempoolAPIService(
	config *configuration.Configuration,
	client *client.APIClient,
) *MempoolAPIService {
	return &MempoolAPIService{
		config: config,
		client: client,
	}
}

// Mempool - Get All Mempool Transactions
func (s *MempoolAPIService) Mempool(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {
	resp, clientErr, _ := s.client.MempoolAPI.Mempool(ctx, request)

	return resp, clientErr
}

// MempoolTransaction - Get a Mempool Transaction
func (s *MempoolAPIService) MempoolTransaction(
	ctx context.Context,
	request *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {
	resp, clientErr, _ := s.client.MempoolAPI.MempoolTransaction(ctx, request)
	return resp, clientErr
}
