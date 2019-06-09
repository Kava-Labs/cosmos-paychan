package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgCreate is for creating a payment channel.
type MsgCreate struct {
	Participants [2]sdk.AccAddress
	Coins        sdk.Coins
}

func (msg MsgCreate) Route() string { return "paychan" }
func (msg MsgCreate) Type() string  { return "paychan" }

func (msg MsgCreate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreate) ValidateBasic() sdk.Error {
	// Validate msg as an optimization to avoid all validation going to keeper. It's run before the sigs are checked by the auth module.
	// Validate without external information (such as account balance)

	//TODO implement

	/* old logic
	// check if all fields present / not 0 valued
	if len(msg.Sender) == 0 {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if len(msg.Receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.Receiver.String())
	}
	if len(msg.Amount) == 0 {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	// Check if coins are sorted, non zero, non negative
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	// TODO check if Address valid?
	*/
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

func (msg MsgSubmitUpdate) Route() string { return "paychan" }
func (msg MsgSubmitUpdate) Type() string  { return "paychan" }

func (msg MsgSubmitUpdate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgSubmitUpdate) ValidateBasic() sdk.Error {

	// TODO implement
	/* old logic
	// check if all fields present / not 0 valued
	if len(msg.Sender) == 0 {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}
	if len(msg.Receiver) == 0 {
		return sdk.ErrInvalidAddress(msg.Receiver.String())
	}
	if len(msg.ReceiverAmount) == 0 {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	// check id ≥ 0
	if msg.Id < 0 {
		return sdk.ErrInvalidAddress(strconv.Itoa(int(msg.Id))) // TODO implement custom errors
	}
	// Check if coins are sorted, non zero, non negative
	if !msg.ReceiverAmount.IsValid() {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	if !msg.ReceiverAmount.IsPositive() {
		return sdk.ErrInvalidCoins(msg.ReceiverAmount.String())
	}
	// TODO check if Address valid?
	*/
	return nil
}

func (msg MsgSubmitUpdate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Submitter}
}
