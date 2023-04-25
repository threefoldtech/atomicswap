package stellar

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/xdr"
)

func ExtractSecret(network string, holdingAccountAdress string, secretHash string, client horizonclient.ClientInterface) ([]byte, error) {
	transactions, err := GetAccountDebitediTransactions(holdingAccountAdress, client)
	if err != nil {
		return nil, fmt.Errorf("Error getting the transaction that debited the holdingAccount: %v", err)
	}
	if len(transactions) == 0 {
		return nil, errors.New("The holdingaccount has not been redeemed yet")
	}
	var extractedSecret []byte
transactionsLoop:
	for _, transaction := range transactions {

		for _, rawSignature := range transaction.Signatures {

			decodedSignature, err := base64.StdEncoding.DecodeString(rawSignature)
			if err != nil {
				return nil, fmt.Errorf("Error base64 decoding signature :%v", err)
			}
			if len(decodedSignature) > xdr.Signature(decodedSignature).XDRMaxSize() {
				continue // this is certainly not the secret we are looking for
			}
			signatureHash := sha256.Sum256(decodedSignature)
			hexSignatureHash := fmt.Sprintf("%x", signatureHash)
			if hexSignatureHash == secretHash {
				extractedSecret = decodedSignature
				break transactionsLoop
			}
		}
	}

	if extractedSecret == nil {
		return nil, errors.New("Unable to find the matching secret")
	}

	return extractedSecret, nil
}
