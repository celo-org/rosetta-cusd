package services

import (
	"context"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Implements the server.NetworkAPIService interface.
type NetworkAPIService struct {
	config *configuration.Configuration
	client *client.APIClient
}

func NewNetworkAPIService(
	config *configuration.Configuration,
	client *client.APIClient,
) *NetworkAPIService {
	return &NetworkAPIService{
		config: config,
		client: client,
	}
}

// endpoint: /network/list
func (s *NetworkAPIService) NetworkList(
	ctx context.Context,
	request *types.MetadataRequest,
) (*types.NetworkListResponse, *types.Error) {
	return &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			s.config.Network,
		},
	}, nil
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
	return &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    RosettaVersion,
			NodeVersion:       NodeVersion,
			MiddlewareVersion: &MiddlewareVersion,
		},
		Allow: &types.Allow{
			OperationStatuses: s.config.OperationStatuses,
			OperationTypes:    s.config.OperationTypes,
		},
	}, nil
}
