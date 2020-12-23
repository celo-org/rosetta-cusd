// Copyright 2020 Celo Org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
)

// Implements the server.BlockAPIServicer interface.
type BlockAPIService struct {
	client *client.APIClient
}

func NewBlockAPIService(
	client *client.APIClient,
) *BlockAPIService {
	return &BlockAPIService{
		client: client,
	}
}

// Extract operations from transferLog
// and update opIndex, operations, prevRelatedOps in place accordingly.
func opsFromLog(
	transferLog gethTypes.Log,
	opIndex *int64,
	operations *[]*types.Operation,
	relatedOps *[]*types.OperationIdentifier,
) {
	// TODO ?: Is there a more direct/better way of converting Topic Hash -> Address?
	from := common.HexToAddress(transferLog.Topics[1].Hex())
	to := common.HexToAddress(transferLog.Topics[2].Hex())

	value := new(big.Int).SetBytes(transferLog.Data)
	// TODO ?: Can failed transfers appear in the logs with status !Removed?
	var status types.OperationStatus
	if transferLog.Removed {
		status = OpFailed
	} else {
		status = OpSuccess
	}

	var opType string
	var inGroup bool = false
	// Distinguish between balance-changing transactions which emit "Transfer" event
	switch {
	case to == ZeroAddress:
		if from == ZeroAddress {
			return
		}
		opType = OpBurn
		// Reset related ops, treat burns as standalone
		relatedOps = &[]*types.OperationIdentifier{}
	case from == ZeroAddress:
		opType = OpMint
		// Reset related ops, treat mints as standalone
		relatedOps = &[]*types.OperationIdentifier{}
	default:
		// TODO: for now, cannot differentiate between transfers and gas fees
		opType = OpTransfer
		inGroup = true
	}
	processOp := func(address common.Address, opValue *big.Int, inGroup bool) {
		op := newAtomicOp(address, *opIndex, opValue, status, opType, *relatedOps)
		*operations = append(*operations, op)
		*opIndex++
		// Do not include standalone ops in a related group
		if inGroup {
			*relatedOps = append(op.RelatedOperations, op.OperationIdentifier)
		}
		return
	}
	// Split operations into atomic ops with effects per-account
	if from != ZeroAddress {
		processOp(from, new(big.Int).Neg(value), inGroup)
	}
	if to != ZeroAddress {
		processOp(to, value, inGroup)
	}
	return
}

func callParamsFromBlock(
	block int64,
	networkId *types.NetworkIdentifier,
) (*types.CallRequest, error) {
	blockIdStr := strconv.FormatInt(block, 10)
	rawParams := &CallLogsParams{
		Event:     "StableToken.Transfer",
		FromBlock: blockIdStr, // fetch single block
		ToBlock:   blockIdStr,
	}
	paramsMap, err := airgap.MarshallToMap(rawParams)
	if err != nil {
		return nil, err
	}
	return &types.CallRequest{
		NetworkIdentifier: networkId,
		Method:            "celo_getLogs",
		Parameters:        paramsMap,
	}, nil
}

// endpoint: /block
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {

	blockResp, clientErr, _ := s.client.BlockAPI.Block(ctx, request)
	if clientErr != nil {
		return nil, clientErr
	}

	// Prior to threshold, StableCoin contract not registered on chain and cannot be accessed via /call
	var threshold int64
	switch networkId := request.NetworkIdentifier.Network; networkId {
	case MainnetId:
		threshold = StableCoinRegisteredMainnet
	case TestnetId:
		threshold = StableCoinRegisteredTestnet
	default:
		logError(fmt.Sprintf("Unknown StableCoin registration for Network %s", request.NetworkIdentifier.Network))
		return nil, ErrValidation
	}

	if blockResp.Block.BlockIdentifier.Index < threshold {
		// TODO think about other_transactions
		blockResp.Block.Transactions = nil
		return blockResp, nil
	}

	// Get the filtered logs for transfer events in the requested block
	callReq, err := callParamsFromBlock(blockResp.Block.BlockIdentifier.Index, request.NetworkIdentifier)
	if err != nil {
		return nil, ErrValidation
	}
	resp, _, err := s.client.CallAPI.Call(ctx, callReq)
	if err != nil {
		return nil, ErrCeloClient
	}
	var result rpc.CallLogsResult
	err = airgap.UnmarshallFromMap(resp.Result, &result)
	if err != nil {
		return nil, ErrValidation
	}

	transactions := []*types.Transaction{}
	var opIndex *int64 = new(int64)
	var currTxHash common.Hash
	var operations *[]*types.Operation
	var relatedOps *[]*types.OperationIdentifier

	for i, transferLog := range result.Logs {
		if transferLog.TxHash != currTxHash {
			// Append fully parsed currTxHash + operations to list of transactions
			if i != 0 {
				transaction := &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{Hash: currTxHash.String()},
					Operations:            *operations,
				}
				transactions = append(transactions, transaction)
			}
			// Set currTxHash (to current hash) and initialize start of transaction
			currTxHash = transferLog.TxHash
			operations = &[]*types.Operation{}
			relatedOps = &[]*types.OperationIdentifier{}
			*opIndex = 0
		}
		// Update the index, operations, relatedOps in place
		opsFromLog(transferLog, opIndex, operations, relatedOps)

		// Ensure last seen log is included in transaction list
		if i == len(result.Logs)-1 {
			transaction := &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{Hash: currTxHash.String()},
				Operations:            *operations,
			}
			transactions = append(transactions, transaction)
		}
	}

	blockResp.Block.Transactions = transactions
	blockResp.OtherTransactions = nil

	return blockResp, nil
}

func newAtomicOp(
	account common.Address,
	opIndex int64,
	value *big.Int,
	opStatus types.OperationStatus,
	opType string,
	relatedOps []*types.OperationIdentifier,
) *types.Operation {

	accountId := rpc.NewAccountIdentifier(account, nil)
	opId := rpc.NewOperationIdentifier(opIndex)
	return &types.Operation{
		OperationIdentifier: opId,
		RelatedOperations:   relatedOps,
		Type:                opType,
		Status:              opStatus.Status,
		Account:             &accountId,
		Amount:              rpc.NewAmount(value, CeloDollar),
	}
}

// endpoint: /block/transaction
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	// ?: Is this endpoint even necessary if block contains required logic?
	blockResp, clientErr := s.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &request.BlockIdentifier.Index,
			Hash:  &request.BlockIdentifier.Hash,
		},
	},
	)
	if clientErr != nil {
		return nil, clientErr
	}
	// Find and return the specified transaction
	for _, t := range blockResp.Block.Transactions {
		if t.TransactionIdentifier.Hash == request.TransactionIdentifier.Hash {
			return &types.BlockTransactionResponse{Transaction: t}, nil
		}
	}
	// Transaction not found or does not contain cUSD relevant operations, return empty
	return &types.BlockTransactionResponse{
		Transaction: &types.Transaction{
			TransactionIdentifier: request.TransactionIdentifier,
		},
	}, nil
}
