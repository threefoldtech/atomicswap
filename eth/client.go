package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient struct {
	*ethclient.Client
	rpcClient *rpc.Client
}

// DialClient dials a new rpc client at the given url
func DialClient(url string) (*EthClient, error) {
	c, err := rpc.DialContext(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return &EthClient{
		Client:    ethclient.NewClient(c),
		rpcClient: c,
	}, nil
}
