package paychan

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/cosmos-paychan/paychan/types"
)

// EndBlocker closes channels that have past their execution time.
// It runs at the end of every block, comparing submitted updates against the current block height.
func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {
	var err sdk.Error
	var channelTags sdk.Tags
	tags := sdk.EmptyTags()

	// Iterate through submittedUpdatesQueue
	// TODO optimise by using store iterator
	q := k.getSubmittedUpdatesQueue(ctx)
	var sUpdate types.SubmittedUpdate
	var found bool

	for _, id := range q {
		// close the channel if the update has reached its execution time.
		// Using >= in case some are somehow missed.
		sUpdate, found = k.getSubmittedUpdate(ctx, id)
		if !found {
			panic("can't find element in queue that should exist")
		}
		if ctx.BlockHeight() >= sUpdate.ExecutionTime {
			k.removeFromSubmittedUpdatesQueue(ctx, sUpdate.ChannelID)
			channelTags, err = k.closeChannel(ctx, sUpdate.Update)
			if err != nil {
				panic(err)
			}
			tags.AppendTags(channelTags)
		}
	}
	return tags
}
