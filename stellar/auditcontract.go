package stellar

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/txnbuild"
)

type (
	AuditContractOutput struct {
		ContractAddress  string `json:"contractAddress"`
		ContractValue    string `json:"contractValue"`
		RecipientAddress string `json:"recipientAddress"`
		RefundAddress    string `json:"refundAddress"`
		SecretHash       string `json:"secretHash"`
		Locktime         int64  `json:"Locktime"`
	}
)

func AuditContract(network string, refundTx txnbuild.Transaction, holdingAccountAdress string, asset txnbuild.Asset, client horizonclient.ClientInterface) (AuditContractOutput, error) {
	holdingAccount, err := client.AccountDetail(horizonclient.AccountRequest{AccountID: holdingAccountAdress})
	if err != nil {
		return AuditContractOutput{}, fmt.Errorf("Error getting the holding account details: %v", err)
	}
	//Check if the signing tresholds are correct
	if holdingAccount.Thresholds.HighThreshold != 2 || holdingAccount.Thresholds.MedThreshold != 2 || holdingAccount.Thresholds.LowThreshold != 2 {
		return AuditContractOutput{}, fmt.Errorf("Holding account signing tresholds are wrong.\nTresholds: High: %d, Medium: %d, Low: %d", holdingAccount.Thresholds.HighThreshold, holdingAccount.Thresholds.MedThreshold, holdingAccount.Thresholds.LowThreshold)
	}
	//Get the signing conditions
	var refundTxHashFromSigningConditions []byte
	recipientAddress := ""
	var secretHash []byte
	for _, signer := range holdingAccount.Signers {
		if signer.Weight == 0 { //The original keypair's signing weight is set to 0
			continue
		}
		switch signer.Type {
		case horizon.KeyTypeNames[strkey.VersionByteAccountID]:
			if recipientAddress != "" {
				return AuditContractOutput{}, fmt.Errorf("Multiple recipients as signer: %s and %s", recipientAddress, signer.Key)
			}
			recipientAddress = signer.Key
			if signer.Weight != 1 {
				return AuditContractOutput{}, fmt.Errorf("Signing weight of the recipient is wrong. Recipient: %s Weight: %d", signer.Key, signer.Weight)
			}
		case horizon.KeyTypeNames[strkey.VersionByteHashTx]:
			if refundTxHashFromSigningConditions != nil {
				return AuditContractOutput{}, errors.New("Multiple refund transaction hashes as signer")
			}

			refundTxHashFromSigningConditions, err = strkey.Decode(strkey.VersionByteHashTx, signer.Key)
			if err != nil {
				return AuditContractOutput{}, fmt.Errorf("Faulty encoded refund transaction hash: %s", err)
			}
			if signer.Weight != 2 {
				return AuditContractOutput{}, fmt.Errorf("Signing weight of the refund transaction is wrong. Weight: %d", signer.Weight)
			}

		case horizon.KeyTypeNames[strkey.VersionByteHashX]:
			if secretHash != nil {
				return AuditContractOutput{}, fmt.Errorf("Multiple secret hashes  transaction hashes as signer: %s and %s", secretHash, signer.Key)
			}
			secretHash, err = strkey.Decode(strkey.VersionByteHashX, signer.Key)
			if err != nil {
				return AuditContractOutput{}, fmt.Errorf("Faulty encoded secret hash: %s", err)
			}
			if signer.Weight != 1 {
				return AuditContractOutput{}, fmt.Errorf("Signing weight of the secret hash is wrong. Weight: %d", signer.Weight)
			}
		default:
			return AuditContractOutput{}, fmt.Errorf("Unexpected signer type: %s", signer.Type)
		}
	}
	//Make sure all signing conditions are present
	if refundTxHashFromSigningConditions == nil {
		return AuditContractOutput{}, errors.New("Missing refund transaction hash as signer")
	}
	if secretHash == nil {
		return AuditContractOutput{}, errors.New("Missing secret as signer")
	}
	if recipientAddress == "" {
		return AuditContractOutput{}, errors.New("Missing recipient as signer")
	}
	//Compare the refund transaction hash in the signing condition to the one of the passed refund transaction
	//cmd.refundTx.Network = targetNetwork TODO: Still required with new library or implied?
	refundTxHash, err := refundTx.Hash(network)
	if err != nil {
		return AuditContractOutput{}, fmt.Errorf("Unable to hash the passed refund transaction: %v", err)
	}
	if !bytes.Equal(refundTxHashFromSigningConditions, refundTxHash[:]) {
		return AuditContractOutput{}, errors.New("Refund transaction hash in the signing condition is not equal to the one of the passed refund transaction")
	}
	//and finally get the locktime and refund address
	lockTime := refundTx.Timebounds().MinTime
	if len(refundTx.Operations()) != 1 {
		return AuditContractOutput{}, fmt.Errorf("Refund transaction is expected to have 1 operation instead of %d", len(refundTx.Operations()))
	}
	refundoperation := refundTx.Operations()[0]
	accountMergeOperation, ok := refundTx.Operations()[0].(*txnbuild.AccountMerge)
	if !ok {
		return AuditContractOutput{}, fmt.Errorf("Expecting an accountmerge operation in the refund transaction but got a %v", reflect.TypeOf(refundoperation))
	}
	if accountMergeOperation.SourceAccount != holdingAccountAdress {
		return AuditContractOutput{}, fmt.Errorf("The refund transaction does not refund from the holding account but from %v", accountMergeOperation.SourceAccount)
	}
	refundAddress := accountMergeOperation.Destination

	balance := ""
	if asset.IsNative() {
		balance, err = holdingAccount.GetNativeBalance()
		if err != nil {
			return AuditContractOutput{}, errors.Wrap(err, "could not get native balance for holding account")
		}
	} else {
		balance = holdingAccount.GetCreditBalance(asset.GetCode(), asset.GetIssuer())
	}
	output := AuditContractOutput{
		ContractAddress:  fmt.Sprintf("%v", holdingAccountAdress),
		ContractValue:    balance,
		RecipientAddress: recipientAddress,
		RefundAddress:    refundAddress,
		SecretHash:       fmt.Sprintf("%x", secretHash),
		Locktime:         lockTime,
	}
	return output, nil
}
