package services

import (
	"net/http"

	"github.com/celo-org/rosetta-cusd/configuration"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/server"
)

// NewBlockchainRouter creates a Mux http.Handler from a collection
// of server controllers.
func NewBlockchainRouter(
	config *configuration.Configuration,
	client *client.APIClient,
	asserter *asserter.Asserter,
) http.Handler {
	// Proxy calls to /network from core rosetta + implement own options
	networkAPIService := NewNetworkAPIService(config, client)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	// Proxy calls to /account from core rosetta + implement own options
	blockAPIService := NewBlockAPIService(config, client)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	// Proxy calls to /mempool from core rosetta
	mempoolAPIService := NewMempoolAPIService(config, client)
	mempoolAPIController := server.NewMempoolAPIController(
		mempoolAPIService,
		asserter,
	)

	// Proxy calls to /account from core rosetta + implement own options
	accountAPIService := NewAccountAPIService(config, client)
	accountAPIController := server.NewAccountAPIController(
		accountAPIService,
		asserter,
	)

	return server.NewRouter(
		networkAPIController,
		blockAPIController,
		mempoolAPIController,
		accountAPIController,
	)
}
