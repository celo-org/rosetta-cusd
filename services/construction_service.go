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
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/parser"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
)

// Implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	client      *client.APIClient
	stableToken *StableToken
}

func NewConstructionAPIService(
	client *client.APIClient,
	stableToken *StableToken,
) *ConstructionAPIService {
	return &ConstructionAPIService{
		client:      client,
		stableToken: stableToken,
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

func parseTransfer(ops []*types.Operation) (*transferTx, error) {
	// TODO add a third transfer (without amount) to allow for GasCurrency specification?
	descriptions := &parser.Descriptions{
		OperationDescriptions: []*parser.OperationDescription{
			{
				Type:    OpTransfer,
				Account: &parser.AccountDescription{Exists: true},
				Amount: &parser.AmountDescription{
					Exists:   true,
					Sign:     parser.NegativeAmountSign,
					Currency: CeloDollar,
				},
			},
			{
				Type:    OpTransfer,
				Account: &parser.AccountDescription{Exists: true},
				Amount: &parser.AmountDescription{
					Exists:   true,
					Sign:     parser.PositiveAmountSign,
					Currency: CeloDollar,
				},
			},
		},
		OppositeAmounts: [][]int{{0, 1}},
		ErrUnmatched:    true,
	}

	matches, err := parser.MatchOperations(descriptions, ops)
	if err != nil {
		return nil, err
	}
	// Check inputs
	fieldErr := func(field string) error {
		return errors.New(fmt.Sprintf("Invalid field: '%s'", field))
	}
	fromOp, _ := matches[0].First()
	fromAddr, ok := rpc.ChecksumAddress(fromOp.Account.Address)
	if !ok {
		return nil, fieldErr("From")
	}
	toOp, _ := matches[1].First()
	toAddr, ok := rpc.ChecksumAddress(toOp.Account.Address)
	if !ok {
		return nil, fieldErr("To")
	}
	value, ok := new(big.Int).SetString(toOp.Amount.Value, 10)
	if !ok {
		return nil, fieldErr("Value")
	}
	return &transferTx{
		To:    toAddr,
		From:  fromAddr,
		Value: value,
	}, nil
}

// endpoint: /construction/preprocess
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {

	transferTx, err := parseTransfer(request.Operations)
	if err != nil {
		logError(fmt.Sprintf("%s", err))
		return nil, ErrValidation
	}

	options := make(map[string]interface{})
	options["From"] = transferTx.From.String()
	// This is currently necessary to properly estimate gas
	options["Method"] = "StableToken.transfer"
	options["Args"] = []string{
		transferTx.To.String(),
		transferTx.Value.String(),
	}

	return &types.ConstructionPreprocessResponse{
		Options: options,
	}, nil
}

// endpoint: /construction/metadata
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionMetadata(ctx, request)

	return resp, clientErr
}

type transferArgs struct {
	To    common.Address
	Value *big.Int
}
type transferTx struct {
	To    *common.Address
	From  *common.Address
	Value *big.Int
	// TODO gasCurrency
}

// endpoint: /construction/payloads
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {

	// Construct unsigned cUSD transaction blob
	var metadata airgap.TxMetadata
	err := airgap.UnmarshallFromMap(request.Metadata, &metadata)
	if err != nil {
		return nil, ErrValidation
	}

	transferTx, err := parseTransfer(request.Operations)
	if err != nil {
		logError(fmt.Sprintf("%s", err))
		return nil, ErrValidation
	}
	metadata.Data, err = s.stableToken.ABI.Pack(
		s.stableToken.ABI.Methods["transfer"].Name,
		transferTx.To,
		transferTx.Value,
	)
	if err != nil {
		logError("could not pack transfer data")
		return nil, ErrValidation
	}

	tx := airgap.Transaction{
		TxMetadata: &metadata,
		Signature:  []byte{},
	}
	// TODO core: extract this into a helper function in core
	gethTx, _ := tx.AsGethTransaction()
	signer := gethTypes.NewEIP155Signer(tx.ChainId)

	// Construct SigningPayload
	payload := &types.SigningPayload{
		AccountIdentifier: &types.AccountIdentifier{
			Address: tx.From.Hex(),
		},
		Bytes:         signer.Hash(gethTx).Bytes(),
		SignatureType: types.EcdsaRecovery,
	}

	unsignedTxJSON, err := json.Marshal(tx)
	if err != nil {
		return nil, ErrInternal
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: string(unsignedTxJSON),
		Payloads:            []*types.SigningPayload{payload},
	}, nil
}

// endpoint: /construction/parse
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	var tx airgap.Transaction
	if !request.Signed {
		err := json.Unmarshal([]byte(request.Transaction), &tx)
		if err != nil {
			return nil, ErrInternal
		}
	} else {
		t := new(gethTypes.Transaction)
		err := t.UnmarshalJSON([]byte(request.Transaction))
		if err != nil {
			return nil, ErrInternal
		}

		from, err := gethTypes.Sender(gethTypes.NewEIP155Signer(t.ChainId()), t)
		if err != nil {
			return nil, ErrInternal
		}

		txMetadata := &airgap.TxMetadata{
			To:                  *t.To(),
			From:                from,
			ChainId:             t.ChainId(),
			Gas:                 t.Gas(),
			GasPrice:            t.GasPrice(),
			Nonce:               t.Nonce(),
			Data:                t.Data(),
			Value:               t.Value(),
			FeeCurrency:         t.FeeCurrency(),
			GatewayFee:          t.GatewayFee(),
			GatewayFeeRecipient: t.GatewayFeeRecipient(),
		}
		v, r, s := t.RawSignatureValues()
		signature := airgap.ValuesToSignature(t.ChainId(), v, r, s)

		tx = airgap.Transaction{
			TxMetadata: txMetadata,
			Signature:  signature,
		}
	}
	// Confirm that the transaction will be sent to the StableToken contract
	if tx.To != s.stableToken.Address {
		logError("transaction 'To' does not match StableToken address")
		return nil, ErrValidation
	}

	// Check method ID
	transferMethod := s.stableToken.ABI.Methods["transfer"]
	method, err := s.stableToken.ABI.MethodById(tx.Data[:4])
	if err != nil || method.Name != "transfer" {
		logError("could not parse method ID")
		return nil, ErrValidation
	}
	// Parse data according to transfer(to, value)
	var transferArgs transferArgs
	err = transferMethod.Inputs.Unpack(&transferArgs, tx.Data[4:])
	if err != nil {
		logError("could not unpack transaction data")
		return nil, ErrValidation
	}
	toAddr := transferArgs.To
	value := transferArgs.Value

	ops := []*types.Operation{
		newAtomicOp(tx.From, 0, new(big.Int).Neg(value), nil, OpTransfer, nil),
		newAtomicOp(toAddr, 1, value, nil, OpTransfer, []*types.OperationIdentifier{{Index: 0}}),
	}
	var resp *types.ConstructionParseResponse
	resp = &types.ConstructionParseResponse{
		Operations: ops,
	}
	if request.Signed {
		resp.AccountIdentifierSigners = []*types.AccountIdentifier{
			{
				Address: tx.From.Hex(),
			},
		}
	}
	return resp, nil
}

// endpoint: /construction/combine
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionCombine(ctx, request)

	return resp, clientErr
}

// endpoint: /construction/hash
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionHash(ctx, request)

	return resp, clientErr
}

// endpoint: /construction/submit
func (s *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	resp, clientErr, _ := s.client.ConstructionAPI.ConstructionSubmit(ctx, request)

	return resp, clientErr
}
