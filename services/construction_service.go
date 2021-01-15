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
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/analyzer"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/celo-org/kliento/contracts"
	// "github.com/celo-org/kliento/registry"
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

	return resp, clientErr
}

// endpoint: /construction/payloads
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionPayloads(ctx, request)

	return resp, clientErr
}

// endpoint: /construction/parse
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionParse(ctx, request)
	if clientErr != nil {
		return nil, clientErr
	}
	// Match expected response
	descriptions := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type:    analyzer.OpSend.String(),
				Account: &parser.AccountDescription{Exists: true},
				Amount: &parser.AmountDescription{Exists: false},
				Metadata: []*parser.MetadataDescription{
					{Key: "contract_address", ValueKind: reflect.String},
					{Key: "data", ValueKind: reflect.String},
				},
			},
		},
		ErrUnmatched:    true,
	}
	fmt.Printf("%+v\n", resp.Operations[0])
	matches, err := parser.MatchOperations(descriptions, resp.Operations)
	if err != nil {
		return nil, ErrUnclearIntent
	}
	fmt.Printf("%+v\n", matches)

	// TODO move to utils if this logic stays
	// for now try to parse as long as it matches hard-coded stableToken addr.
	var stableTokenAddr string
	switch networkId := request.NetworkIdentifier.Network; networkId {
	case MainnetId:
		stableTokenAddr = StableTokenAddrMainnet
	case TestnetId:
		stableTokenAddr = StableTokenAddrTestnet
	default:
		logError(fmt.Sprintf("Unknown StableToken contract address for Network %s", request.NetworkIdentifier.Network))
		return nil, ErrValidation
	}

	sendOp, _ := matches[0].First()
	strAddr, ok := sendOp.Metadata["contract_address"].(string)

	if !ok {
		logError("unexpected 'contract_address' type; string conversion failed")
		return nil, ErrValidation
	}
	// Confirm that this is a transaction sent to the StableToken contract
	if strAddr != stableTokenAddr {
		logError("'to' address does not match StableToken contract address")
		return nil, ErrInternal
	}

	// Unclear if we can use this??
	// stableToken, err := contracts.NewStableToken(common.HexToAddress(strAddr), nil)
	// if err != nil {
	// 	logError("initializing StableToken object failed")
	// 	return nil, ErrInternal
	// }
	// fmt.Printf("Initialized stabletoken\n")
	// fmt.Printf("%+v\n", stableToken)

	// Process using the deserialize/serialize args of argbuilder/celo method?
	// TODO --> data is produced through the ABI; perhaps find a way of including the ABI here?

	// TODO could abstract all of this into a ParseStableTokenTransfer function

	parsedABI, err := contracts.ParseStableTokenABI()
	if err != nil {
		logError("could not parse StableToken ABI")
		return nil, ErrInternal
	}
	dataStr, ok := sendOp.Metadata["data"].(string)
	if !ok {
		logError("unexpected 'data' type; string conversion failed")
		return nil, ErrValidation
	}
	dataBytes, err := hex.DecodeString(dataStr)
	if err != nil {
		logError("decoding data to hex bytes")
		return nil, ErrValidation
	}
	method, err := parsedABI.MethodById(dataBytes)
	if err != nil {
		return nil, ErrInternal
	}
	// fmt.Printf("parsed method: %s\n", method)
	// fmt.Printf("parsed method: %s\n", method.Name)
	// fmt.Printf("parsed method: %s\n", method.RawName)
	// fmt.Printf("parsed method: %v\n", method.Outputs)
	// fmt.Printf("parsed method: %v\n", method.Inputs)

	beepParsed := parsedABI.Methods["transfer"]
	// fmt.Printf("parsed method: %s\n", beepParsed)
	// fmt.Printf("parsed method: %s\n", beepParsed.Name)
	// fmt.Printf("parsed method: %s\n", beepParsed.RawName)
	// fmt.Printf("parsed method: %v\n", beepParsed.Outputs)
	// fmt.Printf("parsed method: %v\n", beepParsed.Inputs)

	fmt.Printf("beep?: %s", parsedABI.Methods["transfer"])
	// TODO --> more robust check that this is the correct ID
	if (parsedABI.Methods["transfer"].RawName == method.RawName) {
		fmt.Printf("WEEEEEE we made it!!!\n")
	}
	// attempt to process "send" operation (with assumption that it is StableToken.transfer)
	// failure at any point --> return ErrUnclearIntent

	// metadata := map[string]interface{}{
	// 	"contract_address": tx.To.Hex(),
	// 	"data":             tx.Data,
	// }

	// ops := []*types.Operation{
	// 	{
	// 		Type: analyzer.OpSend.String(),
	// 		OperationIdentifier: &types.OperationIdentifier{
	// 			Index: 0,
	// 		},
	// 		Account: &types.AccountIdentifier{
	// 			Address: tx.From.Hex(),
	// 		},
	// 		Metadata: metadata,
	// 	},
	// }



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
