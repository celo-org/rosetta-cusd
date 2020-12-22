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
	"log"
	"math/big"
	"strconv"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
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

// endpoint: /block
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {

	// TODO revisit if we need to filter out non-cUSD transactions; currently return all transactions
	blockResp, clientErr, _ := s.client.BlockAPI.Block(ctx, request)
	// return blockResp, clientErr
	if clientErr != nil {
		return nil, clientErr
	}

	// grab logs for block
	// parse transactions per log
	// ignore other_transactions? -- or take set difference of these transactions at the end

	// Prior to threshold, StableCoin contract not registered on chain and cannot be accessed via /call
	var threshold int64
	switch networkId := request.NetworkIdentifier.Network; networkId {
	case MainnetId:
		threshold = StableCoinRegisteredMainnet
	case TestnetId:
		threshold = StableCoinRegisteredTestnet
	default:
		log.Printf("Unknown StableCoin registration for Network %s\n", request.NetworkIdentifier.Network)
		return nil, ErrValidation
	}
	
	if blockResp.Block.BlockIdentifier.Index < threshold {
		// TODO think about other_transactions
		blockResp.Block.Transactions = nil
		return blockResp, nil
	}

	// get the filtered logs for transfer events in the requested block

	// TODO: func callLogsForBlock() --> get callLogs for this block
	blockIdStr := strconv.FormatInt(blockResp.Block.BlockIdentifier.Index, 10)
	rawParams := &CallLogsParams{
		Event:     "StableToken.Transfer",
		FromBlock: blockIdStr, // fetch single block
		ToBlock:   blockIdStr,
	}
	paramsMap, err := airgap.MarshallToMap(rawParams)
	if err != nil {
		return nil, ErrValidation
	}
	callReq := &types.CallRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		Method:            "celo_getLogs",
		Parameters:        paramsMap,
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
	var opIndex int64
	var currTxHash common.Hash // TODO --> tx identifier instead?
	var operations []*types.Operation
	var prevRelatedOps *[]*types.OperationIdentifier

	for i, transferLog := range result.Logs {

		if transferLog.TxHash != currTxHash {
			// Append fully parsed currTxHash + operations to list of transactions
			if i != 0 {
				transaction := &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{Hash: currTxHash.String()},
					Operations: operations,
				}
				transactions = append(transactions, transaction)	
			}
			// Set currTxHash (to current hash) and initialize start of transaction
			currTxHash = transferLog.TxHash
			operations = []*types.Operation{}
			prevRelatedOps = &[]*types.OperationIdentifier{}
			opIndex = 0
		}

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
				continue
			}
			opType = OpBurn
			// Reset related ops, treat burns as standalone
			prevRelatedOps = &[]*types.OperationIdentifier{}
		case from == ZeroAddress:
			opType = OpMint
			// Reset related ops, treat mints as standalone
			prevRelatedOps = &[]*types.OperationIdentifier{}
		default:
			// TODO: for now, cannot differentiate between transfers and gas fees
			opType = OpTransfer
			inGroup = true
		}

		processOp := func(address common.Address, opValue *big.Int, inGroup bool) {
			op := newAtomicOp(address, opIndex, opValue, status, opType, *prevRelatedOps)
			operations = append(operations, op)
			opIndex += 1
			// Do not include standalone ops in a related group
			if inGroup {
				*prevRelatedOps = append(op.RelatedOperations, op.OperationIdentifier)
			}
		}
		// Split operations into atomic ops with effects per-account
		if from != ZeroAddress {
			processOp(from, new(big.Int).Neg(value), inGroup)
		}
		if to != ZeroAddress {
			processOp(to, value, inGroup)
		}
		if i == len(result.Logs) - 1 {
			// Ensure last seen log is included in transaction list
			transaction := &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{Hash: currTxHash.String()},
				Operations: operations,
			}
			transactions = append(transactions, transaction)
		}

	}
	// TODO append the last processed transaction in the list (or at the end of for loop --> if i == len(result.logs) - 1)
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

	// Prior to threshold, StableCoin contract not registered on chain and cannot be accessed via /call
	var threshold int64
	switch networkId := request.NetworkIdentifier.Network; networkId {
	case MainnetId:
		threshold = StableCoinRegisteredMainnet
	case TestnetId:
		threshold = StableCoinRegisteredTestnet
	default:
		log.Printf("Unknown StableCoin registration for Network %s\n", request.NetworkIdentifier.Network)
		return nil, ErrValidation
	}

	if request.BlockIdentifier.Index < threshold {
		return &types.BlockTransactionResponse{
			Transaction: &types.Transaction{
				TransactionIdentifier: request.TransactionIdentifier,
				Operations:            []*types.Operation{},
			},
		}, nil
	}

	// TODO moving logic to block (to do the looping in one pass as opposed to per transaction); this is a first pass
	// ? for CB: should this all instead be happening in /block? i.e. computing all transactions + gas, etc.

	// get the filtered logs for transfer events
	blockIdStr := strconv.FormatInt(request.BlockIdentifier.Index, 10)
	rawParams := &CallLogsParams{
		Event:     "StableToken.Transfer",
		FromBlock: blockIdStr, // fetch single block
		ToBlock:   blockIdStr,
	}
	paramsMap, err := airgap.MarshallToMap(rawParams)
	if err != nil {
		return nil, ErrValidation
	}
	callReq := &types.CallRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		Method:            "celo_getLogs",
		Parameters:        paramsMap,
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

	var opIndex int64 = 0
	
	// TODO test
	// var operations []*types.Operation
	operations := []*types.Operation{}
	prevRelatedOps := &[]*types.OperationIdentifier{}

	// TODO: first pass -- more efficient to have this within block logic, but for now, match structure of core rosetta
	// loop through the logs until the transaction matches the requested transaction hash
	for _, transferLog := range result.Logs {
		if transferLog.TxHash.String() != request.TransactionIdentifier.Hash {
			continue
		}

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
				continue
			}
			opType = OpBurn
			// Reset related ops, treat burns as standalone
			prevRelatedOps = &[]*types.OperationIdentifier{}
		case from == ZeroAddress:
			opType = OpMint
			// Reset related ops, treat mints as standalone
			prevRelatedOps = &[]*types.OperationIdentifier{}
		default:
			// TODO: for now, cannot differentiate between transfers and gas fees
			opType = OpTransfer
			inGroup = true
		}

		processOp := func(address common.Address, opValue *big.Int, inGroup bool) {
			op := newAtomicOp(address, opIndex, opValue, status, opType, *prevRelatedOps)
			operations = append(operations, op)
			opIndex += 1
			// Do not include standalone ops in a related group
			if inGroup {
				*prevRelatedOps = append(op.RelatedOperations, op.OperationIdentifier)
			}
		}
		// Split operations into atomic ops with effects per-account
		if from != ZeroAddress {
			processOp(from, new(big.Int).Neg(value), inGroup)
		}
		if to != ZeroAddress {
			processOp(to, value, inGroup)
		}
	}

	return &types.BlockTransactionResponse{
		Transaction: &types.Transaction{
			TransactionIdentifier: request.TransactionIdentifier,
			Operations:            operations,
		},
	}, nil
}
