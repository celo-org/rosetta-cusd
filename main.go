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
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/celo-org/rosetta-cusd/services"

	"github.com/coinbase/rosetta-sdk-go/client"
	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/server"
)

func main() {
	rosettaCoreURL := flag.String("core.url", "http://localhost", "Listening URL for core Rosetta RPC server")
	rosettaCorePort := flag.Uint("core.port", 8080, "Listening port for core Rosetta RPC server")
	rosettaCusdAddr := flag.String("cusd.addr", "", "Listening address for cUSD http server")
	rosettaCusdPort := flag.Uint("cusd.port", 8081, "Listening port for cUSD http server")
	flag.Parse()

	listenAddress := func(addr string, port uint) string {
		return fmt.Sprintf("%s:%d", addr, port)
	}

	clientCfg := client.NewConfiguration(
		listenAddress(*rosettaCoreURL, *rosettaCorePort),
		fetcher.DefaultUserAgent,
		&http.Client{
			Timeout: fetcher.DefaultHTTPTimeout,
		},
	)
	client := client.NewAPIClient(clientCfg)

	router, err := services.CreateRouter(client)
	if err != nil {
		log.Fatal(err)
	}
	loggedRouter := server.LoggerMiddleware(router)
	log.Printf("Listening on port %d\n", *rosettaCusdPort)
	log.Fatal(http.ListenAndServe(listenAddress(*rosettaCusdAddr, *rosettaCusdPort), loggedRouter))
}
