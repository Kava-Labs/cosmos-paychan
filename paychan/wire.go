package paychan

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterWire registers the paychan message types with the given codec.
// A codec needs to have implementations of interface types registered before it can serialize them.
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "paychan/MsgCreate", nil)
	cdc.RegisterConcrete(MsgSubmitUpdate{}, "paychan/MsgSubmitUpdate", nil)
}

// TODO move this to near the msg definitions?
var msgCdc = wire.NewCodec()

func init() {
	wire.RegisterCrypto(msgCdc)
	RegisterWire(msgCdc)
}
