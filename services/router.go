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
	"net/http"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/server"
)

// Creates a Mux http.Handler from a collection of server controllers.
func CreateRouter(
	client *client.APIClient,
	asserter *asserter.Asserter,
) (http.Handler, error) {

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

	// Proxy calls to /construction/* from core rosetta + implement own options
	constructionAPIService := NewConstructionAPIService(client)
	constructionAPIController := server.NewConstructionAPIController(constructionAPIService, asserter)

	return server.NewRouter(
		networkAPIController,
		blockAPIController,
		mempoolAPIController,
		accountAPIController,
		constructionAPIController,
	), nil
}
