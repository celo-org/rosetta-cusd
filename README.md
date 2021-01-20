# Celo Rosetta cUSD

A module that runs on top of the core [Celo Rosetta RPC server](https://github.com/celo-org/rosetta) to implement the Rosetta specifications for cUSD, an ERC-20 stable token on the Celo blockchain.

This module is currently a work in progress, and should not be considered production ready yet.

## Endpoints

Rosetta cUSD exposes the following endpoints:

* `POST /network/list`: Get List of Available Networks
* `POST /network/status`: Get Network Status
* `POST /network/options`: Get Network Options
* `POST /block`: Get a Block
* `POST /block/transaction`: Get a Block Transaction
* `POST /mempool`: Get All Mempool Transactions
* `POST /mempool/transaction`: Get a Mempool Transaction
* `POST /account/balance`: Get an Account Balance

Currently, the Construction API endpoints are not yet implemented.

## Running Rosetta cUSD

Prerequisites: the core Rosetta RPC server must be running in the background, on the version/branch specified in `services/versions.go` under `RosettaCoreVersion` (currently: `master` commit `c749ba869cc8fd70f4719fa726e0efead033c60b`), as this module queries it in order to service the above endpoints.

The main command is `go run main.go` with the following flags:

```txt
Flags:
      --core.url string     Listening URL for core Rosetta RPC server (default: "http://localhost")
      --core.port uint      Listening port for core Rosetta RPC server (default: 8080)
      --cUSD.url string     Listening address for cUSD http server (default: "")
      --cUSD.port uint      Listening port for cUSD http server (default: 8081)
```

### Building and running from Docker image

*Coming soon -- a rosetta-cusd image in our public repository, once we have tagged our first release.*

To build an up-to-date docker image locally, run the following:

```sh
cd rosetta-cusd
docker build -t gcr.io/celo-testnet/rosetta-cusd:$USER .
```

To run `rosetta-cusd`, first ensure that the rosetta core service is running. If it is running on `localhost`, you can use `"http://host.docker.internal"` as the `CORE_URL` below.

```sh
# Create and delete:
docker run --rm -p 8081:8081 gcr.io/celo-testnet/rosetta-cusd:$USER --core.url CORE_URL --core.port CORE_PORT
# OR name the container, can be restarted after stopping:
docker run --name rosetta-cusd -p 8081:8081 gcr.io/celo-testnet/rosetta-cusd:$USER --core.url CORE_URL --core.port CORE_PORT
```

Note that this command will listen for rosetta-cusd requests on port `8081`.

## Running `rosetta-cli` checks

Run the `rosetta-cli check:data` by running both the core and module servers and then using the appropriate CLI configuration file located in `test/rosetta-cli-conf/[NETWORK]`.
