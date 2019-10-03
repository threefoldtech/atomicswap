package stellar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeyPair(t *testing.T) {
	pair, err := GenerateKeyPair()
	if assert.NoError(t, err) {
		assert.NotNil(t, pair.Address())
		assert.NotNil(t, pair.Seed())
	}
}
