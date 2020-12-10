package configuration

import (
	"github.com/coinbase/rosetta-sdk-go/types"
)

type Configuration struct {
	Network           *types.NetworkIdentifier
	OperationTypes    []string
	OperationStatuses []*types.OperationStatus

	Port               int
	RosettaCeloURL string
}

// Default returns the default configuration
func Default() *Configuration {
	return &Configuration{
		// TODO: parameterize this + Port, URL, etc.
		Network: &types.NetworkIdentifier{
			Blockchain: "celo",
			Network:    "44787",
		},
		OperationTypes: []string{ // TODO: move to const in contracts
			"transfer",
		},
		OperationStatuses: []*types.OperationStatus{
			{
				Status:     "success",
				Successful: true,
			},
		},
		Port:               8081,
		RosettaCeloURL: "http://localhost:8080",
	}
}
