package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/celo-org/rosetta-cusd/configuration"
	"github.com/celo-org/rosetta-cusd/services"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
)

func main() {
	config := configuration.Default()

	asserter, err := asserter.NewServer(
		config.OperationTypes,
		true,
		[]*types.NetworkIdentifier{config.Network},
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	clientCfg := client.NewConfiguration(
		config.RosettaCeloURL,
		fetcher.DefaultUserAgent,
		&http.Client{
			Timeout: fetcher.DefaultHTTPTimeout,
		},
	)
	client := client.NewAPIClient(clientCfg)

	router := services.NewBlockchainRouter(config, client, asserter)
	loggedRouter := server.LoggerMiddleware(router)
	log.Printf("Listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), loggedRouter))
}