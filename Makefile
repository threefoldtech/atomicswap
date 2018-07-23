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
	test -z "$(shell gometalinter --vendor --disable-all \
		--enable=gofmt \
		--enable=vet \
		--enable=gosimple \
		--enable=unconvert \
		--enable=ineffassign \
		--deadline=10m ./... 2>&1 | tee /dev/stderr)"

test-go:
	go test -v -race $(testpkgs)

.PHONY: all test install test-linter test-go ethatomicswap btcatomicswap
