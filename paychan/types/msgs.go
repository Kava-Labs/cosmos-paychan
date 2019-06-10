package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCreate is for creating a payment channel.
type MsgCreate struct {
	Participants [2]sdk.AccAddress // sender, receiver
	Coins        sdk.Coins
}

func (msg MsgCreate) Route() string { return RouterKey }
func (msg MsgCreate) Type() string  { return "create" }

func (msg MsgCreate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	// check if addresses are ok
	if msg.Participants[0].Empty() {
		return sdk.ErrInvalidAddress(msg.Participants[0].String())
	}
	if msg.Participants[1].Empty() {
		return sdk.ErrInvalidAddress(msg.Participants[1].String())
	}
	// Check if coins are sorted, have valid denoms, non zero, non negative
	if !(msg.Coins.IsValid() && msg.Coins.IsAllPositive()) {
		return sdk.ErrInvalidCoins(msg.Coins.String())
	}
	return nil
}

func (msg MsgCreate) GetSigners() []sdk.AccAddress {
	// Only sender must sign to create a paychan
	return []sdk.AccAddress{msg.Participants[0]} // select sender address
}

// MsgSubmitUpdate is for closing a payment channel.
type MsgSubmitUpdate struct {
	Update
	Submitter sdk.AccAddress
}

func (msg MsgSubmitUpdate) Route() string { return RouterKey }
func (msg MsgSubmitUpdate) Type() string  { return "submit_update" }

func (msg MsgSubmitUpdate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSubmitUpdate) ValidateBasic() sdk.Error {

	// check if submitter address ok
	if msg.Submitter.Empty() {
		return sdk.ErrInvalidAddress(msg.Submitter.String())
	}
	// check id ≥ 0
	if msg.Update.ChannelID < 0 {
		return sdk.ErrInvalidAddress(strconv.Itoa(int(msg.ChannelID)))
	}
	// Check if coins are sorted, have valid denoms, non negative
	if !msg.Update.Payout.IsValid() || msg.Update.Payout.IsAnyNegative() { // a payout can be zero
		return sdk.ErrInvalidCoins(fmt.Sprintf("coins in payout invalid: %v", msg.Update.Payout))
	}
	return nil
}

func (msg MsgSubmitUpdate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}
