package paychan

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan/types"
)

// NewHandler returns a handler for "paychan" type messages.
// Called when adding routes to a newly created app.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgCreate:
			return handleMsgCreate(ctx, k, msg)
		case types.MsgSubmitUpdate:
			return handleMsgSubmitUpdate(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized distribution message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgCreate
// Leaves validation to the keeper methods.
func handleMsgCreate(ctx sdk.Context, k Keeper, msg types.MsgCreate) sdk.Result {
	tags, err := k.CreateChannel(ctx, msg.Participants[0], msg.Participants[len(msg.Participants)-1], msg.Coins)
	if err != nil {
		return err.Result()
	}
	// TODO any other information that should be returned in Result?
	return sdk.Result{
		Tags: tags,
	}
}

// Handle MsgSubmitUpdate
// Leaves validation to the keeper methods.
func handleMsgSubmitUpdate(ctx sdk.Context, k Keeper, msg types.MsgSubmitUpdate) sdk.Result {
	var err sdk.Error
	tags := sdk.EmptyTags()

	// TODO refactor signer detection - move to keeper or find nicer setup
	channel, _ := k.getChannel(ctx, msg.Update.ChannelID)
	participants := channel.Participants

	// if only sender signed
	if msg.Submitter.Equals(participants[0]) {
		tags, err = k.InitCloseChannelBySender(ctx, msg.Update)
		// else if receiver signed
	} else if msg.Submitter.Equals(participants[len(participants)-1]) {
		tags, err = k.CloseChannelByReceiver(ctx, msg.Update)
	}

	if err != nil {
		return err.Result()
	}
	// These tags can be used by clients to subscribe to channel close attempts
	return sdk.Result{
		Tags: tags,
	}
}
