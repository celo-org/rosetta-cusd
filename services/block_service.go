package services

import (
	"context"
	"log"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Implements the server.BlockAPIServicer interface.
type BlockAPIService struct {
	config *configuration.Configuration
	client *client.APIClient
}

func NewBlockAPIService(
	config *configuration.Configuration,
	client *client.APIClient,
) *BlockAPIService {
	return &BlockAPIService{
		config: config,
		client: client,
	}
}

// endpoint: /block
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	resp, clientErr, _ := s.client.BlockAPI.Block(ctx, request)
	if clientErr != nil {
		return nil, clientErr
	}

	// Ensure all transactions are requested
	populatedTransactions := resp.Block.Transactions
	resp.Block.Transactions = nil

	otherTransactions := []*types.TransactionIdentifier{}
	for _, populatedTransaction := range populatedTransactions {
		// Skip block transactions
		if populatedTransaction.TransactionIdentifier.Hash == resp.Block.BlockIdentifier.Hash {
			continue
		}

		otherTransactions = append(otherTransactions, populatedTransaction.TransactionIdentifier)
	}

	if len(resp.OtherTransactions) == 0 {
		resp.OtherTransactions = otherTransactions
	} else {
		resp.OtherTransactions = append(resp.OtherTransactions, otherTransactions...)
	}

	return resp, nil
}

// endpoint: /block/transaction
func (s *BlockAPIService) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	errMsg := "Block Transaction Unimplemented"
	log.Printf("ERROR %s", errMsg)

	return nil, &types.Error{
		Code: ErrUnimplemented,
		Message: errMsg,
	}
}
