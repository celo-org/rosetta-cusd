module github.com/celo-org/rosetta-cusd

go 1.14

require (
	github.com/celo-org/kliento v0.2.0
	github.com/celo-org/rosetta v0.7.7-0.20210119175349-040db59734bb
	github.com/coinbase/rosetta-sdk-go v0.5.9
	github.com/ethereum/go-ethereum v1.9.23
	github.com/segmentio/golines v0.0.0-20200306054842-869934f8da7b // indirect
)

replace github.com/ethereum/go-ethereum => github.com/celo-org/celo-blockchain v1.1.2
