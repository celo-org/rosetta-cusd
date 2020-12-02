package services

import (
	"context"
	"log"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	// "github.com/ethereum/go-ethereum/common"
	// EthTypes "github.com/ethereum/go-ethereum/core/types"
)

// BlockAPIService implements the server.BlockAPIServicer interface.
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

// func (s *BlockAPIService) getReceipt(ctx context.Context, transaction *types.TransactionIdentifier) (*EthTypes.Receipt, error) {
// 	resp, _, err := s.client.CallAPI.Call(ctx, &types.CallRequest{
// 		NetworkIdentifier: s.config.Network,
// 		Method:            "eth_getTransactionReceipt",
// 		Parameters: map[string]interface{}{
// 			"tx_hash": transaction.Hash,
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	hexReceipt, ok := resp.Result["receipt"]
// 	if !ok {
// 		return nil, errors.New("receipt is missing")
// 	}
// 	bytesReceipt := common.Hex2Bytes(hexReceipt.(string))

// 	receipt := new(EthTypes.Receipt)
// 	if err := receipt.UnmarshalJSON(bytesReceipt); err != nil {
// 		return nil, err
// 	}

// 	return receipt, nil
// }

func (s *BlockAPIService) BlockTransaction(ctx context.Context, request *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	errMsg := "Block Transaction Unimplemented"
	log.Printf("ERROR %s", errMsg)

	return nil, &types.Error{
		Code: ErrUnimplemented,
		Message: errMsg,
	}

	// receipt, err := s.getReceipt(ctx, request.TransactionIdentifier)
	// if err != nil {
	// 	return nil, &types.Error{Message: err.Error()}
	// }

	// operations, err := parser.ParseLogs(ctx, s.config.Contracts, receipt.Logs)
	// if err != nil {
	// 	return nil, &types.Error{Message: err.Error()}
	// }

	// return &types.BlockTransactionResponse{
	// 	Transaction: &types.Transaction{
	// 		TransactionIdentifier: request.TransactionIdentifier,
	// 		Operations:            operations,
	// 	},
	// }, nil
}
