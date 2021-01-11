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

// Implements the server.NetworkAPIService interface.
type NetworkAPIService struct {
	client *client.APIClient
}

func NewNetworkAPIService(
	client *client.APIClient,
) *NetworkAPIService {
	return &NetworkAPIService{
		client: client,
	}
}

// endpoint: /network/list
func (s *NetworkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	resp, clientErr, _ := s.client.NetworkAPI.NetworkList(ctx, request)

	return resp, clientErr
}

// endpoint: /network/status
func (s *NetworkAPIService) NetworkStatus(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkStatusResponse, *types.Error) {
	resp, clientErr, _ := s.client.NetworkAPI.NetworkStatus(ctx, request)

	return resp, clientErr
}

// endpoint: /network/options
func (s *NetworkAPIService) NetworkOptions(
	ctx context.Context,
	request *types.NetworkRequest,
) (*types.NetworkOptionsResponse, *types.Error) {
	resp, clientErr, err := s.client.NetworkAPI.NetworkOptions(ctx, request)
	if err != nil {
		return nil, clientErr
	}

	// TODO check that resp.Version.MiddlewareVersion matches expected RosettaCoreVersion

	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    resp.Version.RosettaVersion,
			NodeVersion:       resp.Version.NodeVersion,
			MiddlewareVersion: &MiddlewareVersion,
		},
		Allow: &types.Allow{
			OperationStatuses: AllOperationStatuses,
			OperationTypes:    AllOperationTypes,
			Errors:            AllErrors,
		},
	}, nil
}
