package services

import (
	"context"
	"log"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

type AccountAPIService struct {
	config *configuration.Configuration
	client *client.APIClient
}

func NewAccountAPIService(
	config *configuration.Configuration,
	client *client.APIClient,
) *AccountAPIService {
	return &AccountAPIService{
		config: config,
		client: client,
	}
}

// endpoint: /account/balance
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
	) (*types.AccountBalanceResponse, *types.Error) {

	errMsg := "Account Balance Unimplemented"
	log.Printf("ERROR %s", errMsg)

	return nil, &types.Error{
		Code: ErrUnimplemented,
		Message: errMsg,
	}
}
