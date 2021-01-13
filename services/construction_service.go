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
	"math/big"
	"context"
	"fmt"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/analyzer"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
)

// Implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	client *client.APIClient
}

func NewConstructionAPIService(
	client *client.APIClient,
) *ConstructionAPIService {
	return &ConstructionAPIService{
		client: client,
	}
}

// endpoint: /construction/derive
func (s *ConstructionAPIService) ConstructionDerive(
	ctx context.Context,
	request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionDerive(ctx, request)

	return resp, clientErr

}

// TODO move into core, as part of parser interface refactor
// for testing now
func checksumAddress(address string) (*common.Address, bool) {
	var addr common.Address
	mcAddr, err := common.NewMixedcaseAddressFromString(address)
	if err != nil {
		return &addr, false
	}
	addr = mcAddr.Address()
	return &addr, true
}

// transferParser creates a closure which when called will
// match and parse a transfer in a specified currency
func transferParser(opTransfer string, currency *types.Currency) (func (ops []*types.Operation) (*airgap.TxArgs, *types.Error)) {
	return func (ops []*types.Operation) (*airgap.TxArgs, *types.Error) {
		descriptions := &parser.Descriptions{
			OperationDescriptions: []*parser.OperationDescription{
				{
					Type:    opTransfer,
					Account: &parser.AccountDescription{Exists: true},
					Amount: &parser.AmountDescription{
						Exists:   true,
						Sign:     parser.NegativeAmountSign,
						Currency: currency,
					},
				},
				{
					Type:    opTransfer,
					Account: &parser.AccountDescription{Exists: true},
					Amount: &parser.AmountDescription{
						Exists:   true,
						Sign:     parser.PositiveAmountSign,
						Currency: currency,
					},
				},
			},
			OppositeAmounts: [][]int{{0, 1}},
			ErrUnmatched:    true,
		}
		matches, err := parser.MatchOperations(descriptions, ops)
		if err != nil {
			return nil, ErrUnclearIntent
		}
		fromOp, _ := matches[0].First()
		fromAddr, ok := checksumAddress(fromOp.Account.Address)
		if !ok {
			return nil, ErrValidation
		}
		toOp, _ := matches[1].First()
		toAddr, ok := checksumAddress(toOp.Account.Address)
		if !ok {
			return nil, ErrValidation
		}
	
		var txArgs airgap.TxArgs
		txArgs.From = *fromAddr
		txArgs.To = toAddr
		txArgs.Value, ok = new(big.Int).SetString(toOp.Amount.Value, 10)
		if !ok {
			return nil, ErrInternal
		}

		return &txArgs, nil
	}
}

// endpoint: /construction/preprocess
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {

	// TODO --> have parseTransferOps use a more general, currency independent function (transferParser(currency) -> func transferParserOps)
	parseCUSDTransfer := transferParser(OpTransfer, CeloDollar)
	txArgs, rosettaErr := parseCUSDTransfer(request.Operations)

	if rosettaErr != nil {
		return nil, rosettaErr
	}
	// Prepare request for "send" transaction in core construction preprocess
	sendReq := &types.ConstructionPreprocessRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		Operations: []*types.Operation{
			{
				OperationIdentifier: &types.OperationIdentifier{
					Index: 0,
				},
				Type: analyzer.OpSend.String(),
				// Ops were parsed successfully, so use from (ops[0]) identifier
				Account: request.Operations[0].Account,
				Metadata: map[string]interface{}{
					"method": "StableToken.transfer",
					"args": [2]string{txArgs.To.Hex(), txArgs.Value.String()},
				},
			},
		},
	}
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionPreprocess(ctx, sendReq)

	return resp, clientErr
}

// endpoint: /construction/metadata
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionMetadata(ctx, request)
	// TODO likely passthrough will work here, may need to prepare formatting of request
	return resp, clientErr
}

// endpoint: /construction/payloads
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	resp, _, _ := s.client.ConstructionAPI.ConstructionPayloads(ctx, request)
	// TODO implement cUSD logic
	fmt.Printf("%v\n", resp)
	return nil, ErrUnimplemented
}

// endpoint: /construction/parse
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	resp, _, _ := s.client.ConstructionAPI.ConstructionParse(ctx, request)
	// TODO implement cUSD logic
	fmt.Printf("%v\n", resp)
	return nil, ErrUnimplemented
}

// endpoint: /construction/combine
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	resp, _, _ := s.client.ConstructionAPI.ConstructionCombine(ctx, request)
	// TODO likely passthrough will work here, may need to prepare formatting of request
	fmt.Printf("%v\n", resp)
	return nil, ErrUnimplemented
}

// endpoint: /construction/hash
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionHash(ctx, request)
	// TODO test: should work as is
	return resp, clientErr
}

// endpoint: /construction/submit
func (s *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionSubmit(ctx, request)
	// TODO test: should work as is
	return resp, clientErr
}
