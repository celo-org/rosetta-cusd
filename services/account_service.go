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
	"math/big"
	"strconv"

	"github.com/celo-org/rosetta/airgap"
	"github.com/celo-org/rosetta/service/rpc"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

type AccountAPIService struct {
	client *client.APIClient
}

func NewAccountAPIService(
	client *client.APIClient,
) *AccountAPIService {
	return &AccountAPIService{
		client: client,
	}
}

// endpoint: /account/balance
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {

	rawParams := &CallParams{
		Method: "StableToken.balanceOf",
		Args:   [1]string{request.AccountIdentifier.Address},
	}

	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Index != nil {
			blockNumber := strconv.FormatInt(*request.BlockIdentifier.Index, 10)
			rawParams.BlockNumber = &blockNumber
		} else {
			logError("Block number is required when passing in a block identifier.")
			return nil, ErrValidation
		}
	}

	paramsMap, err := airgap.MarshallToMap(rawParams)
	if err != nil {
		return nil, ErrValidation
	}
	callReq := &types.CallRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		Method:            "celo_call",
		Parameters:        paramsMap,
	}
	resp, _, err := s.client.CallAPI.Call(ctx, callReq)
	if err != nil {
		return nil, ErrCeloClient
	}

	var result rpc.CallResult
	err = airgap.UnmarshallFromMap(resp.Result, &result)
	if err != nil {
		return nil, ErrValidation
	}
	// Sanity check
	if request.BlockIdentifier != nil {
		if request.BlockIdentifier.Hash != nil && *request.BlockIdentifier.Hash != result.BlockIdentifier.Hash {
			logError("Mismatch between requested and returned block hash.")
			return nil, ErrInternal
		}
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: result.BlockIdentifier,
		Balances: []*types.Amount{
			rpc.NewAmount(new(big.Int).SetBytes(result.Raw), CeloDollar),
		},
	}, nil
}
