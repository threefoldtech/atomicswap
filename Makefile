testpkgs = ./cmd/ethatomicswap
BIN = $(GOPATH)/bin

all: test install

install: ethatomicswap btcatomicswap

ethatomicswap:
	go build -o $(BIN)/ethatomicswap ./cmd/ethatomicswap

btcatomicswap:
	go build -o $(BIN)/btcatomicswap ./cmd/btcatomicswap

test: test-linter test-go

test-linter:
	test -z "$(shell golangci-lint run  --no-config --disable-all \
		--enable=gofmt \
		--enable=vet \
		--enable=gosimple \
		--enable=goimports \
		--enable=unconvert \
		--enable=ineffassign \
		--deadline=10m 2>&1 | tee /dev/stderr)"

test-go:
	go test -v -race $(testpkgs)

test-web3:
	cd cmd/ethatomicswap/contract/src && truffle test

.PHONY: all test install test-linter test-go ethatomicswap btcatomicswap
