package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestSubmittedUpdatesQueue(t *testing.T) {
	t.Run("RemoveMatchingElements", func(t *testing.T) {
		// SETUP
		q := SubmittedUpdatesQueue{4, 8, 23, 0, 5645657}
		// ACTION
		q.RemoveMatchingElements(23)
		// CHECK RESULTS
		expectedQ := SubmittedUpdatesQueue{4, 8, 0, 5645657}
		assert.Equal(t, expectedQ, q)

		// SETUP
		q = SubmittedUpdatesQueue{0}
		// ACTION
		q.RemoveMatchingElements(0)
		// CHECK RESULTS
		expectedQ = SubmittedUpdatesQueue{}
		assert.Equal(t, expectedQ, q)
	})
}

func TestPayout(t *testing.T) {
	t.Run("IsNotNegative", func(t *testing.T) {
		p := Payout{sdk.Coins{sdk.NewInt64Coin("usd", 4), sdk.NewInt64Coin("gbp", 0)}, sdk.Coins{sdk.NewInt64Coin("usd", 129879234), sdk.NewInt64Coin("gbp", 1)}}
		assert.True(t, p.IsNotNegative())

		p = Payout{sdk.Coins{sdk.NewInt64Coin("usd", 0), sdk.NewInt64Coin("gbp", 0)}, sdk.Coins{sdk.NewInt64Coin("usd", 129879234), sdk.NewInt64Coin("gbp", 1)}}
		assert.True(t, p.IsNotNegative())
	})
	t.Run("Sum", func(t *testing.T) {
		p := Payout{
			sdk.Coins{sdk.NewInt64Coin("eur", 1), sdk.NewInt64Coin("usd", 0)},
			sdk.Coins{sdk.NewInt64Coin("eur", 1), sdk.NewInt64Coin("usd", 100), sdk.NewInt64Coin("gbp", 1)},
		}
		expected := sdk.Coins{sdk.NewInt64Coin("eur", 2), sdk.NewInt64Coin("gbp", 1), sdk.NewInt64Coin("usd", 100)}
		assert.Equal(t, expected, p.Sum())
	})
}
