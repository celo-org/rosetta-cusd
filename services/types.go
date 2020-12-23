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
	"log"

	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// Blocks before this do not have StableCoin Contract
	StableCoinRegisteredTestnet = 544
	StableCoinRegisteredMainnet = 2962
	TestnetId                   = "44787"
	MainnetId                   = "42220"

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

// Using imported call types from Rosetta requires involved manipulations due to
// different formats expected by the requests vs. the internals of the RPC server.
type CallParams struct {
	Method      string    `json:"method,omitempty"`
	Args        [1]string `json:"args,omitempty"`
	BlockNumber *string   `json:"block_number,omitempty"`
}

type CallLogsParams struct {
	Event     string          `json:"event"`
	Topics    [][]interface{} `json:"topics,omitempty"`
	FromBlock string          `json:"from_block"`
	ToBlock   string          `json:"to_block"`
}

func logError(errMsg string) {
	log.Printf("ERROR: %s\n", errMsg)
}
