package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gorilla/mux"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {}

// handler functions ...
// create paychan
// close paychan
// get paychan(s)
// send paychan payment
// get balance from receiver
// get balance from local storage
// handle incoming payment
