package paychan

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan/types"
)

// Keeper of the paychan store
// Handles validation internally. Does not rely on calling code to do validation.
// Aim to keep public methods safe, private ones not necessarily.
// Keepers contain main business logic of the module.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec // needed to serialize objects before putting them in the store
	bankKeeper bank.Keeper

	//codespace sdk.CodespaceType TODO custom errors
}

// NewKeeper returns a new payment channel keeper. This is called when creating new app.
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, bk bank.Keeper) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		bankKeeper: bk,
		//codespace:  codespace,
	}
	return keeper
}

// CreateChannel creates a new payment channel in the blockchain and locks up sender funds.
func (k Keeper) CreateChannel(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, coins sdk.Coins) (sdk.Tags, sdk.Error) {

	// Check addresses valid (Technically don't need to check sender address is valid as SubtractCoins checks)
	if len(sender) == 0 {
		return nil, sdk.ErrInvalidAddress(sender.String())
	}
	if len(receiver) == 0 {
		return nil, sdk.ErrInvalidAddress(receiver.String())
	}
	// check coins are sorted and positive (disallow channels with zero balance)
	if !coins.IsValid() {
		return nil, sdk.ErrInvalidCoins(coins.String())
	}
	if !coins.IsAllPositive() {
		return nil, sdk.ErrInvalidCoins(coins.String())
	}

	// subtract coins from sender
	_, err := k.bankKeeper.SubtractCoins(ctx, sender, coins)
	if err != nil {
		return nil, err
	}
	// Calculate next id
	id := k.getNewChannelID(ctx)
	// create new Paychan struct
	channel := types.Channel{
		ID:           id,
		Participants: [2]sdk.AccAddress{sender, receiver},
		Coins:        coins,
	}
	// save to db
	k.setChannel(ctx, channel)

	// TODO add to tags

	return sdk.EmptyTags(), err
}

// InitCloseChannelBySender initiates the close of a payment channel, subject to a dispute period.
func (k Keeper) InitCloseChannelBySender(ctx sdk.Context, update types.Update) (sdk.Tags, sdk.Error) {
	// This is roughly the default path for non unidirectional channels

	// get the channel
	channel, found := k.getChannel(ctx, update.ChannelID)
	if !found {
		return nil, sdk.ErrInternal("Channel doesn't exist")
	}
	err := VerifyUpdate(channel, update)
	if err != nil {
		return nil, err
	}

	q := k.getSubmittedUpdatesQueue(ctx)
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel

		// In bidirectional channels the new update is compared against existing and replaces it if it has a higher sequence number.

		// existingSUpdate, found := k.getSubmittedUpdate(ctx, update.ChannelID)
		// if !found {
		// 	panic("can't find element in queue that should exist")
		// }
		// k.addToSubmittedUpdatesQueue(ctx, k.applyNewUpdate(existingSUpdate, update))

		// However in unidirectional case, only the sender can close a channel this way. No clear need for them to be able to submit an update replacing a previous one they sent, so don't allow it.
		// TODO tags
		// TODO custom errors
		sdk.ErrInternal("Sender can't submit an update for channel if one has already been submitted.")
	} else {
		// No one has tried to update channel
		submittedUpdate := types.SubmittedUpdate{
			Update:        update,
			ExecutionTime: ctx.BlockHeight() + types.ChannelDisputeTime,
		}
		k.addToSubmittedUpdatesQueue(ctx, submittedUpdate)
	}

	// TODO tags

	return sdk.EmptyTags(), nil
}

// CloseChannelByReceiver immediately closes a payment channel.
func (k Keeper) CloseChannelByReceiver(ctx sdk.Context, update types.Update) (sdk.Tags, sdk.Error) {

	// get the channel
	channel, found := k.getChannel(ctx, update.ChannelID)
	if !found {
		return nil, sdk.ErrInternal("Channel doesn't exist")
	}
	err := VerifyUpdate(channel, update)
	if err != nil {
		return nil, err
	}

	// Check if there is an update in the queue already
	q := k.getSubmittedUpdatesQueue(ctx)
	if q.Contains(update.ChannelID) {
		// Someone has previously tried to update channel but receiver has final say
		k.removeFromSubmittedUpdatesQueue(ctx, update.ChannelID)
	}

	tags, err := k.closeChannel(ctx, update)

	return tags, err
}

// Main function that compares updates against each other.
// Pure function, Not needed in unidirectional case.
// func (k Keeper) applyNewUpdate(existingSUpdate SubmittedUpdate, proposedUpdate Update) SubmittedUpdate {
// 	var returnUpdate SubmittedUpdate

// 	if existingSUpdate.Sequence > proposedUpdate.Sequence {
// 		// update accepted
// 		returnUpdate = SubmittedUpdate{
// 			Update:        proposedUpdate,
// 			ExecutionTime: existingSUpdate.ExecutionTime, // FIXME any new update proposal should be subject to full dispute period from submission
// 		}
// 	} else {
// 		// update rejected
// 		returnUpdate = existingSUpdate
// 	}
// 	return returnUpdate
// }

// VerifyUpdate checks that a given update is valid for a given channel.
func VerifyUpdate(channel types.Channel, update types.Update) sdk.Error {

	// Check the num of payout participants match channel participants
	if len(update.Payout) != len(channel.Participants) {
		return sdk.ErrInternal("Payout doesn't match number of channel participants")
	}
	// Check each coins are valid
	for _, coins := range update.Payout {
		if !coins.IsValid() {
			return sdk.ErrInternal("Payout coins aren't formatted correctly")
		}
	}
	// Check payout coins are each not negative (can be zero though)
	if !update.Payout.IsNotNegative() {
		return sdk.ErrInternal("Payout cannot be negative")
	}
	// Check payout sums to match channel.Coins
	if !channel.Coins.IsEqual(update.Payout.Sum()) {
		return sdk.ErrInternal("Payout amount doesn't match channel amount")
	}
	// Check sender signature is OK
	if !verifySignatures(channel, update) {
		return sdk.ErrInternal("Signature on update not valid")
	}
	return nil
}

// closeChannel closes a payment channel without any checks.
// It doesn't check if the given update matches an existing channel.
func (k Keeper) closeChannel(ctx sdk.Context, update types.Update) (sdk.Tags, sdk.Error) {
	var err sdk.Error

	channel, _ := k.getChannel(ctx, update.ChannelID)
	// TODO check channel exists and participants matches update payout length

	// Add coins to sender and receiver
	// TODO check for possible errors first to avoid coins being half paid out?
	for i, coins := range update.Payout {
		_, err = k.bankKeeper.AddCoins(ctx, channel.Participants[i], coins)
		if err != nil {
			panic(err)
		}
	}

	k.deleteChannel(ctx, update.ChannelID)

	// TODO tags
	return sdk.EmptyTags(), nil
}

// verifySignatures checks whether the signatures on a given update are correct.
func verifySignatures(channel types.Channel, update types.Update) bool {
	// In non unidirectional channels there will be more than one signature to check

	signBytes := update.GetSignBytes()

	address := channel.Participants[0] // sender
	pubKey := update.Sigs[0].PubKey
	cryptoSig := update.Sigs[0].CryptoSignature

	// Check public key submitted with update signature matches the account address
	valid := bytes.Equal(pubKey.Address(), address) &&
		// Check the signature is correct
		pubKey.VerifyBytes(signBytes, cryptoSig)
	return valid

}

// ============================================================
// SUBMITTED UPDATES QUEUE
// ============================================================

func (k Keeper) addToSubmittedUpdatesQueue(ctx sdk.Context, sUpdate types.SubmittedUpdate) {
	// always overwrite prexisting values - leave paychan logic to higher levels
	// get current queue
	q := k.getSubmittedUpdatesQueue(ctx)
	// append ID to queue
	if !q.Contains(sUpdate.ChannelID) {
		q = append(q, sUpdate.ChannelID)
	}
	// set queue
	k.setSubmittedUpdatesQueue(ctx, q)
	// store submittedUpdate
	k.setSubmittedUpdate(ctx, sUpdate)
}
func (k Keeper) removeFromSubmittedUpdatesQueue(ctx sdk.Context, channelID types.ChannelID) {
	// get current queue
	q := k.getSubmittedUpdatesQueue(ctx)
	// remove id
	q.RemoveMatchingElements(channelID)
	// set queue
	k.setSubmittedUpdatesQueue(ctx, q)
	// delete submittedUpdate
	k.deleteSubmittedUpdate(ctx, channelID)
}

func (k Keeper) getSubmittedUpdatesQueue(ctx sdk.Context) types.SubmittedUpdatesQueue {
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSubmittedUpdatesQueueKey())

	var suq types.SubmittedUpdatesQueue // if the submittedUpdatesQueue not found then return an empty one
	if bz != nil {
		// unmarshal
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &suq)
	}
	return suq

}
func (k Keeper) setSubmittedUpdatesQueue(ctx sdk.Context, suq types.SubmittedUpdatesQueue) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(suq)
	// write to db
	key := k.getSubmittedUpdatesQueueKey()
	store.Set(key, bz)
}
func (k Keeper) getSubmittedUpdatesQueueKey() []byte {
	return []byte("submittedUpdatesQueue")
}

// ============================================================
// SUBMITTED UPDATES
// These are keyed by the IDs of their associated Channels
// This section deals with only setting and getting
// ============================================================

func (k Keeper) getSubmittedUpdate(ctx sdk.Context, channelID types.ChannelID) (types.SubmittedUpdate, bool) {

	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(GetSubmittedUpdateKey(channelID))

	var sUpdate types.SubmittedUpdate
	if bz == nil {
		return sUpdate, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &sUpdate)
	// return
	return sUpdate, true
}

// Store payment channel struct in blockchain store.
func (k Keeper) setSubmittedUpdate(ctx sdk.Context, sUpdate types.SubmittedUpdate) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(sUpdate) // panics if something goes wrong
	// write to db
	key := GetSubmittedUpdateKey(sUpdate.ChannelID)
	store.Set(key, bz) // panics if something goes wrong
}

func (k Keeper) deleteSubmittedUpdate(ctx sdk.Context, channelID types.ChannelID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(GetSubmittedUpdateKey(channelID))
	// TODO does this have return values? What happens when key doesn't exist?
}

// GetSubmittedUpdateKey returns the store key for the SubmittedUpdate corresponding to the channel with the given ID.
func GetSubmittedUpdateKey(channelID types.ChannelID) []byte {
	return []byte(fmt.Sprintf("submittedUpdate:%d", channelID))
}

// ============================================================
// CHANNELS
// ============================================================

// getChannel retrieves a payment channel struct from the blockchain store.
func (k Keeper) getChannel(ctx sdk.Context, channelID types.ChannelID) (types.Channel, bool) {
	// load from DB
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetChannelKey(channelID))

	var channel types.Channel
	if bz == nil {
		return channel, false
	}
	// unmarshal
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &channel)
	// return
	return channel, true
}

// setChannel stores a payment channel struct in the blockchain store.
func (k Keeper) setChannel(ctx sdk.Context, channel types.Channel) {
	store := ctx.KVStore(k.storeKey)
	// marshal
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(channel) // panics if something goes wrong
	// write to db
	key := types.GetChannelKey(channel.ID)
	store.Set(key, bz) // panics if something goes wrong
}

// deleteChannel removes a channel struct from the blockchain store.
func (k Keeper) deleteChannel(ctx sdk.Context, channelID types.ChannelID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetChannelKey(channelID))
	// TODO does this have return values? What happens when key doesn't exist?
}

// getNewChannelID deterministically creates a new id, updating a global counter counter.
func (k Keeper) getNewChannelID(ctx sdk.Context) types.ChannelID {
	// get last channel ID
	var lastID types.ChannelID
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(getLastChannelIDKey())
	if bz == nil {
		lastID = -1 // TODO is just setting to zero if uninitialized ok?
	} else {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &lastID)
	}
	// increment to create new one
	newID := lastID + 1
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(newID)
	// set last channel id again
	store.Set(getLastChannelIDKey(), bz)
	// return
	return newID
}

// getLastChannelIDKey returns the store key used for the global id counter.
func getLastChannelIDKey() []byte {
	return []byte("lastChannelID")
}
