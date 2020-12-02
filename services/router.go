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
	// Proxy calls to /network from ETH + implement own options
	networkAPIService := NewNetworkAPIService(config, client)
	networkAPIController := server.NewNetworkAPIController(
		networkAPIService,
		asserter,
	)

	// // Populate blocks using ABI -> use rosetta-ethereum /call
	blockAPIService := NewBlockAPIService(config, client)
	blockAPIController := server.NewBlockAPIController(
		blockAPIService,
		asserter,
	)

	return server.NewRouter(networkAPIController, blockAPIController)
}
