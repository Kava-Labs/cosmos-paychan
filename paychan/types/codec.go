package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc) // TODO maynot be needed
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers the paychan message types with the given codec.
// A codec needs to have implementations of interface types registered before it can serialize them.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "paychan/MsgCreate", nil)
	cdc.RegisterConcrete(MsgSubmitUpdate{}, "paychan/MsgSubmitUpdate", nil)
}
