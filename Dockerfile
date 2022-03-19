# Instructions for Building & Deploying to the image repository
#    export COMMIT_SHA=$(git rev-parse HEAD)
#    docker build --build-arg COMMIT_SHA=$COMMIT_SHA -t gcr.io/celo-testnet/rosetta-cusd:$COMMIT_SHA .
#    docker push gcr.io/celo-testnet/rosetta-cusd:$COMMIT_SHA

FROM golang:1.16.12-alpine as builder
WORKDIR /app
RUN apk add --no-cache make gcc musl-dev linux-headers git

# Download dependencies & cache them in docker layer
COPY go.mod go.sum ./
RUN go mod download

# Build rosetta-cusd
#  (this saves to redownload everything when go.mod/sum didn't change)
COPY . .
RUN go build --tags musl -o main .

# Default argument set by --cusd.port
EXPOSE 8081/tcp

ENTRYPOINT ["./main"]
