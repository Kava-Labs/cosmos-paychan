package paychan

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the paychan message types with the given codec.
// A codec needs to have implementations of interface types registered before it can serialize them.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "paychan/MsgCreate", nil)
	cdc.RegisterConcrete(MsgSubmitUpdate{}, "paychan/MsgSubmitUpdate", nil)
}

// TODO move this to near the msg definitions?
var msgCdc = codec.New()

func init() {
	codec.RegisterCrypto(msgCdc)
	RegisterCodec(msgCdc)
}
