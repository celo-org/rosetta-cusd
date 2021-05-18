# Celo Rosetta cUSD

A module that runs on top of the core [Celo Rosetta RPC server](https://github.com/celo-org/rosetta) to implement the Rosetta specifications for cUSD, an ERC-20 stable token on the Celo blockchain.

This module is currently a work in progress, and should not be considered production ready yet.

## Endpoints

Rosetta cUSD exposes Data and Construction API endpoints according to the [Rosetta API spec](https://www.rosetta-api.org/docs/ConstructionApi.html)

The following Data API endpoints are implemented:

- `POST /network/list`: Get List of Available Networks
- `POST /network/status`: Get Network Status
- `POST /network/options`: Get Network Options
- `POST /block`: Get a Block
- `POST /block/transaction`: Get a Block Transaction
- `POST /mempool`: Get All Mempool Transactions
- `POST /mempool/transaction`: Get a Mempool Transaction
- `POST /account/balance`: Get an Account Balance

All the Construction API (`POST /construction/*` are implemented) which allow the user to construct and sign cUSD transactions. Note that currently, this only allows transaction gas fees to be paid in CELO, although the CELO platform also allows users to pay gas fees in cUSD. This is a point of future work.

## Running Rosetta cUSD

Prerequisites: the [core Rosetta RPC server](https://github.com/celo-org/rosetta) must be running in the background, on the version/branch specified in `services/versions.go` under `RosettaCoreVersion` (currently: `beta/construction` commit `198ea51`), as this module queries it in order to service the above endpoints. See the [README.md](https://github.com/celo-org/rosetta/blob/master/README.md) for instructions on how to run the core server.

### Running from source

Navigate to the root repository. Run:

```sh
make all
go run main.go [optional flags, see below]
```

The following optional flags can be provided:

```txt
Flags:
      --core.url string     Listening URL for core Rosetta RPC server (default: "http://localhost")
      --core.port uint      Listening port for core Rosetta RPC server (default: 8080)
      --cUSD.url string     Listening address for cUSD http server (default: "")
      --cUSD.port uint      Listening port for cUSD http server (default: 8081)
```

### Building and running from Docker image

#### Recommended: Running using public image registry

To run `rosetta-cusd`, first ensure that the rosetta core service is running. If it is running on `localhost`, you can use `"http://host.docker.internal"` as the `$CORE_URL` below.

Tagged releases will have corresponding docker images in the public registry at [us.gcr.io/celo-testnet/rosetta-cusd](us.gcr.io/celo-testnet/rosetta-cusd). To run from a specific tag, run:

```sh
export RELEASE_VERSION=DESIRED_VERSION  # replace with desired version, ex) v0.0.1
# Create and delete (--rm flag)
docker run --rm  -p 8081:8081 us.gcr.io/celo-testnet/rosetta-cusd:$RELEASE_VERSION --core.url $CORE_URL --core.port $CORE_PORT
# OR name the container, can be restarted after stopping
docker run --name rosetta-cusd -p 8081:8081 us.gcr.io/celo-testnet/rosetta-cusd:$RELEASE_VERSION --core.url $CORE_URL --core.port $CORE_PORT
```

Note that this command will listen for rosetta-cusd requests on port `8081`.

If you run into issues with this, you may need to pull the image first and then retry the above command. To pull the image, you can run the following:

```sh
docker pull us.gcr.io/celo-testnet/rosetta-cusd:$RELEASE_VERSION
```

#### For devs: Building and running a docker image locally

This is useful for building and testing changes made to the Dockerfile, or for ensuring that changes to the module still work as expected in the docker image that will be created from it.

To build an docker image from the local repo, run the following:

```sh
cd rosetta-cusd
docker build -t gcr.io/celo-testnet/rosetta-cusd:$USER .
```

As with running a public image, make sure that the core service is running; see instructions above.

```sh
# Create and delete:
docker run --rm -p 8081:8081 gcr.io/celo-testnet/rosetta-cusd:$USER --core.url $CORE_URL --core.port $CORE_PORT
# OR name the container, can be restarted after stopping:
docker run --name rosetta-cusd -p 8081:8081 gcr.io/celo-testnet/rosetta-cusd:$USER --core.url $CORE_URL --core.port $CORE_PORT
```

## Running `rosetta-cli` checks

Run the `rosetta-cli check:data` by running both the core and module servers and then using the appropriate CLI configuration file located in `test/rosetta-cli-conf/[NETWORK]`.
