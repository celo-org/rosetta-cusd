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

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/celo-org/rosetta-cusd/configuration"
	"github.com/celo-org/rosetta-cusd/services"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/server"
)

func main() {
	// TODO create CLI for config options
	config := configuration.Default()
	clientCfg := client.NewConfiguration(
		config.RosettaCeloURL,
		fetcher.DefaultUserAgent,
		&http.Client{
			Timeout: fetcher.DefaultHTTPTimeout,
		},
	)
	client := client.NewAPIClient(clientCfg)

	router, err := services.CreateRouter(config, client)
	if err != nil {
		log.Fatal(err)
	}
	loggedRouter := server.LoggerMiddleware(router)
	log.Printf("Listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), loggedRouter))
}
