package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
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
		p := Payout{cs(c("usd", 4), c("gbp", 0)), cs(c("usd", 129879234), c("gbp", 1))}
		assert.True(t, p.IsNotNegative())

		p = Payout{cs(c("usd", 0), c("gbp", 0)), cs(c("usd", 129879234), c("gbp", 1))}
		assert.True(t, p.IsNotNegative())
	})

	t.Run("Sum", func(t *testing.T) {
		p := Payout{
			cs(c("eur", 1), c("usd", 0)),
			cs(c("eur", 1), c("usd", 100), c("gbp", 1)),
		}
		expected := cs(c("eur", 2), c("gbp", 1), c("usd", 100))
		assert.Equal(t, expected, p.Sum())
	})

	// TODO test IsAnyNegative
	// TODO test IsValid
}

func TestMsgCreate(t *testing.T) {
	tests := []struct {
		name       string
		sender     sdk.AccAddress
		receiver   sdk.AccAddress
		coins      sdk.Coins
		expectPass bool
	}{
		{"happyPath", testAddrs[0], testAddrs[1], cs(c("gbp", 1000)), true},
		{"emptyAddresses", sdk.AccAddress{}, sdk.AccAddress{}, cs(c("gbp", 1000)), false},
		{"emptyCoins", testAddrs[0], testAddrs[1], cs(), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgCreate{
				Participants: [2]sdk.AccAddress{tc.sender, tc.receiver},
				Coins:        tc.coins,
			}
			if tc.expectPass {
				assert.NoError(t, msg.ValidateBasic())
			} else {
				assert.Error(t, msg.ValidateBasic())
			}
		})
	}
}
func TestMsgSubmitUpdate(t *testing.T) {
	tests := []struct {
		name       string
		submitter  sdk.AccAddress
		update     Update
		expectPass bool
	}{
		{"happyPath", testAddrs[0], Update{0, Payout{cs(c("usd", 1000)), cs(c("gbp", 1000))}, [1]UpdateSignature{UpdateSignature{}}}, true},
		{"negativeID", testAddrs[0], Update{-9999999, Payout{cs(c("usd", 1000)), cs(c("gbp", 1000))}, [1]UpdateSignature{UpdateSignature{}}}, false},
		{"emptyAddr", sdk.AccAddress{}, Update{0, Payout{cs(c("usd", 1000)), cs(c("gbp", 1000))}, [1]UpdateSignature{UpdateSignature{}}}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := MsgSubmitUpdate{
				Submitter: tc.submitter,
				Update:    tc.update,
			}
			if tc.expectPass {
				assert.NoError(t, msg.ValidateBasic())
			} else {
				assert.Error(t, msg.ValidateBasic())
			}
		})
	}
}

var _, testAddrs = mock.GeneratePrivKeyAddressPairs(10)

// TODO change these to not use any input validation
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
