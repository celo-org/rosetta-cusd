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

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Implements the server.MempoolAPIServicer interface.
type MempoolAPIService struct {
	client *client.APIClient
}

func NewMempoolAPIService(
	client *client.APIClient,
) *MempoolAPIService {
	return &MempoolAPIService{
		client: client,
	}
}

// endpoint: /mempool
func (s *MempoolAPIService) Mempool(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {
	resp, clientErr, _ := s.client.MempoolAPI.Mempool(ctx, request)

	return resp, clientErr
}

// endpoint: /mempool/transaction
func (s *MempoolAPIService) MempoolTransaction(
	ctx context.Context,
	request *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {
	resp, clientErr, _ := s.client.MempoolAPI.MempoolTransaction(ctx, request)
	return resp, clientErr
}
