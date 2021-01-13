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

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
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
	// TODO test: should work as is
	return resp, clientErr

}

// endpoint: /construction/preprocess
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	resp, _, _ := s.client.ConstructionAPI.ConstructionPreprocess(ctx, request)
	// TODO implement cUSD logic
	fmt.Printf("%v\n", resp)
	return nil, ErrUnimplemented
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
