package stellar

import (
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateKeyPair(t *testing.T) {
	pair, err := GenerateKeyPair()
	if assert.NoError(t, err) {
		assert.NotNil(t, pair.Address())
		assert.NotNil(t, pair.Seed())
	}
}
func TestGetAccount(t *testing.T) {
	address := "GAA6DAO4EQAEUK7MWQAIVGAMO3IBCY5WU5YZM6KSDKZJ7ONLRGIRSL7M"
	client := horizonclient.MockClient{}
	client.Mock.On("AccountDetail", mock.Anything).Return(horizon.Account{
		AccountID: address,
	}, nil)
	account, err := GetAccount(address, &client)
	if assert.NoError(t, err) {
		assert.Equal(t, address, account.GetAccountID())
	}
}
