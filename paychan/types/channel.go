package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

// Channel represents a payment channel.
// Participants is limited to two as currently these are unidirectional channels.
// Last participant is designated as receiver.
type Channel struct {
	ID           ChannelID
	Participants [2]sdk.AccAddress // [senderAddr, receiverAddr]
	Coins        sdk.Coins
}

// Implement fmt.Stringer interface for compatibility while sdk moves over to using yaml // TODO
func (Channel) String() string { return "CHANNEL FORMATTING ERROR" }

const ChannelDisputeTime = int64(50000) // about 3 days measured in blocks // TODO add as param in channels

type ChannelID int64 // TODO swap for uint64

func NewChannelIDFromString(s string) (ChannelID, error) {
	n, err := strconv.ParseInt(s, 10, 64) // parse using base 10, into an int64
	if err != nil {
		return 0, err
	}
	// TODO check ≥ 0
	return ChannelID(n), nil
}

// The data that is passed between participants as payments, and submitted to the blockchain to close a channel.
type Update struct {
	ChannelID ChannelID
	Payout    Payout
	//Sequence  int64 Not needed for unidirectional channels
	Sigs [1]UpdateSignature // only sender needs to sign in unidirectional
}

func (u Update) GetSignBytes() []byte {
	bz, err := ModuleCdc.MarshalJSON(struct {
		ChannelID ChannelID
		Payout    Payout
	}{
		ChannelID: u.ChannelID,
		Payout:    u.Payout})

	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

type UpdateSignature struct { // TODO rename to Signature
	PubKey          crypto.PubKey
	CryptoSignature []byte
}

// Payout is a list of coins to be paid to each of Channel.Participants
type Payout [2]sdk.Coins

func (p Payout) IsAnyNegative() bool {
	for _, coins := range p {
		if coins.IsAnyNegative() {
			return true
		}
	}
	return false
}
func (p Payout) Sum() sdk.Coins {
	var total sdk.Coins
	for _, coins := range p {
		total = total.Add(coins.Sort())
		total = total.Sort()
	}
	return total
}
func (p Payout) IsValid() bool {
	result := true
	for _, coins := range p {
		result = result && coins.IsValid()
	}
	return result
}

// An update that has been submitted to the blockchain, but not yet acted on.
type SubmittedUpdate struct {
	Update
	ExecutionTime int64 // BlockHeight
}

// Implement fmt.Stringer interface for compatibility while sdk moves over to using yaml // TODO
func (SubmittedUpdate) String() string { return "SUBMITTED UPDATE FORMATTING ERROR" }

type SubmittedUpdatesQueue []ChannelID // not technically a queue

// Check if value is in queue
func (suq SubmittedUpdatesQueue) Contains(channelID ChannelID) bool {
	found := false
	for _, id := range suq {
		if id == channelID {
			found = true
			break
		}
	}
	return found
}

// Remove all values from queue that match argument
func (suq *SubmittedUpdatesQueue) RemoveMatchingElements(channelID ChannelID) {
	newSUQ := SubmittedUpdatesQueue{}

	for _, id := range *suq {
		if id != channelID {
			newSUQ = append(newSUQ, id)
		}
	}
	*suq = newSUQ
}
