package types

import "fmt"

const (
	// ModuleName is the name of the module
	ModuleName = "paychan"
	// StoreKey is the store key string
	StoreKey = ModuleName

	// RouterKey is the message route
	RouterKey = ModuleName

	// QuerierRoute is the querier route
	QuerierRoute = ModuleName
)

// GetChannelKey returns the store key for the channel with the given ID.
func GetChannelKey(channelID ChannelID) []byte {
	return []byte(fmt.Sprintf("channel:%d", channelID))
}

// GetSubmittedUpdateKey returns the store key for the SubmittedUpdate corresponding to the channel with the given ID.
func GetSubmittedUpdateKey(channelID ChannelID) []byte {
	return []byte(fmt.Sprintf("submittedUpdate:%d", channelID))
}
