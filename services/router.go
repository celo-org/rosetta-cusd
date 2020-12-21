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
	"net/http"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// Creates a Mux http.Handler from a collection of server controllers.
func CreateRouter(
	config *configuration.Configuration,
	client *client.APIClient,
) (http.Handler, error) {

	// Make sure network options match underlying core service options
	resp, _, err := client.NetworkAPI.NetworkList(context.Background(), &types.MetadataRequest{})
	if err != nil {
		return nil, err
	}

	asserter, err := asserter.NewServer(
		AllOperationTypes,
		true,
		resp.NetworkIdentifiers,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Proxy calls to /network from core rosetta
	networkAPIService := NewNetworkAPIService(client)
	networkAPIController := server.NewNetworkAPIController(networkAPIService, asserter)

	// Proxy calls to /account from core rosetta + implement own options
	blockAPIService := NewBlockAPIService(client)
	blockAPIController := server.NewBlockAPIController(blockAPIService, asserter)

	// Proxy calls to /mempool from core rosetta
	mempoolAPIService := NewMempoolAPIService(client)
	mempoolAPIController := server.NewMempoolAPIController(mempoolAPIService, asserter)

	// Proxy calls to /account from core rosetta + implement own options
	accountAPIService := NewAccountAPIService(client)
	accountAPIController := server.NewAccountAPIController(accountAPIService, asserter)

	return server.NewRouter(
		networkAPIController,
		blockAPIController,
		mempoolAPIController,
		accountAPIController,
	), nil
}
