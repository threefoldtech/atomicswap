testpkgs = ./cmd/ethatomicswap ./cmd/stellaratomicswap/stellar
BIN = $(GOPATH)/bin

all: test install

install: ethatomicswap btcatomicswap

ethatomicswap:
	go build -o $(BIN)/ethatomicswap ./cmd/ethatomicswap

btcatomicswap:
	go build -o $(BIN)/btcatomicswap ./cmd/btcatomicswap

stellaratomicswap:
	go build -o $(BIN)/stellaratomicswap ./cmd/stellaratomicswap

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

.PHONY: all test install test-linter test-go ethatomicswap btcatomicswap stellaratomicswap
