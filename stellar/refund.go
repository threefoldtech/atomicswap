package stellar

import (
	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/txnbuild"
)

func Refund(network string, refundTx txnbuild.Transaction, client horizonclient.ClientInterface) (string, error) {
	result, err := SubmitTransaction(&refundTx, client)
	if err != nil {
		return "", errors.Wrap(err, "failed to submit refund transaction")
	}
	return result.ID, nil
}
