package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeKey string) {
	//r.HandleFunc("/channels", getChannelsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/channels/{id}", getChannelHandlerFn(cliCtx, storeKey)).Methods("GET")
	r.HandleFunc("/channels/{id}/submitted-update", getUpdateHandlerFn(cliCtx, storeKey)).Methods("GET")
	r.HandleFunc("/channels", createChannelHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/channels/{id}/submitted-update", submitUpdateHandlerFn(cliCtx)).Methods("POST") // use simulate flag on post body to verify an update is valid
}

func getChannelHandlerFn(cliCtx context.CLIContext, storeKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse inputs
		vars := mux.Vars(r)
		channelID, err := types.NewChannelIDFromString(vars["id"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get channel from store
		res, err := cliCtx.QueryStore(types.GetChannelKey(channelID), storeKey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(res) == 0 {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("No channel found with id %s", channelID))
		}

		// Print response
		var channel types.Channel
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res, &channel); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
func getUpdateHandlerFn(cliCtx context.CLIContext, storeKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse inputs
		vars := mux.Vars(r)
		channelID, err := types.NewChannelIDFromString(vars["id"])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Get Update from store
		res, err := cliCtx.QueryStore(types.GetSubmittedUpdateKey(channelID), storeKey)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(res) == 0 {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("No submitted update found for channel with id %s", channelID))
		}

		// Print response
		var sUpdate types.SubmittedUpdate
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res, &sUpdate); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

type CreateChannelRequest struct {
	BaseReq  rest.BaseReq   `json:"base_req"`
	Receiver sdk.AccAddress `json:"receiver"` // in bech32
	Coins    sdk.Coins      `json:"coins"`
}

func createChannelHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get args from post body
		var req CreateChannelRequest
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create the msg
		msg := types.MsgCreate{
			Participants: [2]sdk.AccAddress{fromAddr, req.Receiver},
			Coins:        req.Coins,
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Generate tx and write response
		clientrest.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

type SubmitUpdateRequest struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Update  types.Update `json:"update"`
}

func submitUpdateHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get args from post body
		var req SubmitUpdateRequest
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}
		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}
		sender, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Create the msg
		msg := types.MsgSubmitUpdate{
			Update:    req.Update,
			Submitter: sender,
		}
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Generate tx and write response
		clientrest.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
