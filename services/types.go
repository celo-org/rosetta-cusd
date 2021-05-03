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
	"errors"
	"log"
	"math/big"

	"github.com/celo-org/kliento/contracts"
	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/celo-org/celo-blockchain/accounts/abi"
	"github.com/celo-org/celo-blockchain/common"
)

const (
	// Operations
	OpTransfer = "transfer"
	OpFee      = "fee"
	OpMint     = "mint"
	OpBurn     = "burn"
)

var (
	// TODO potentially remove from Rosetta core, as it shouldn't really be used there (perhaps for Construction)
	CeloDollar = rpc.CeloDollar

	// StableToken contract param
	ZeroAddress common.Address = common.HexToAddress("0x0")

	// Error codes and messages
	ErrValidation    = rpc.ErrValidation
	ErrCeloClient    = rpc.ErrCeloClient
	ErrUnimplemented = rpc.ErrUnimplemented
	ErrInternal      = rpc.ErrInternal

	AllErrors = []*types.Error{
		ErrValidation,
		ErrCeloClient,
		ErrUnimplemented,
		ErrInternal,
	}

	// Operations and statuses
	OpSuccess = types.OperationStatus{
		Status:     "success",
		Successful: true,
	}
	OpFailed = types.OperationStatus{
		Status:     "failed",
		Successful: false,
	}
	AllOperationStatuses = []*types.OperationStatus{
		&OpSuccess,
		&OpFailed,
	}
	AllOperationTypes = []string{
		OpTransfer,
		OpFee,
		OpMint,
		OpBurn,
	}
)

// Types and wrappers for types that are not specific to one service
type StableToken struct {
	BlockThreshold int64
	Address        common.Address
	ABI            *abi.ABI
}

func NewStableToken(networkId string) (*StableToken, error) {
	var params StableToken
	var err error
	params.ABI, err = contracts.ParseStableTokenABI()
	if err != nil {
		logError("could not parse StableToken ABI")
		return nil, err
	}

	switch networkId {
	// Mainnet
	case "42220":
		params.BlockThreshold = 2962
		params.Address = common.HexToAddress("0x765de816845861e75a25fca122bb6898b8b1282a")
	// Testnet
	case "44787":
		params.BlockThreshold = 544
		params.Address = common.HexToAddress("0x874069Fa1Eb16D44d622F2e0Ca25eeA172369bC1")
	default:
		return nil, errors.New("unable to initialize StableToken")
	}
	return &params, nil
}

func newAtomicOp(
	account common.Address,
	opIndex int64,
	value *big.Int,
	opStatus *types.OperationStatus,
	opType string,
	relatedOps []*types.OperationIdentifier,
) *types.Operation {

	accountId := rpc.NewAccountIdentifier(account, nil)
	opId := rpc.NewOperationIdentifier(opIndex)
	op := &types.Operation{
		OperationIdentifier: opId,
		RelatedOperations:   relatedOps,
		Type:                opType,
		Account:             &accountId,
		Amount:              rpc.NewAmount(value, CeloDollar),
	}
	if opStatus != nil {
		op.Status = opStatus.Status
	}
	return op
}

func logError(errMsg string) {
	log.Printf("ERROR: %s\n", errMsg)
}
