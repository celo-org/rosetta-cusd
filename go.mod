module github.com/celo-org/rosetta-cusd

go 1.14

require (
	github.com/celo-org/rosetta v0.7.7-0.20210105143400-c749ba869cc8
	github.com/coinbase/rosetta-sdk-go v0.5.9
	github.com/ethereum/go-ethereum v1.9.23
)

replace github.com/ethereum/go-ethereum => github.com/celo-org/celo-blockchain v1.1.2
