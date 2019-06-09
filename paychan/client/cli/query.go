package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the paychan module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}

	queryCmd.AddCommand(client.GetCommands(
		GetCmd_GetChannel(storeKey, cdc),
		GetCmd_GetSubmittedUpdate(storeKey, cdc),
	)...)

	return queryCmd
}

func GetCmd_GetChannel(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "paychan [paychan-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Get details of a channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse and validate input
			channelID, err := types.NewChannelIDFromString(args[0])
			if err != nil {
				return err
			}

			// Query the node
			res, err := cliCtx.QueryStore(types.GetChannelKey(channelID), storeKey)
			if err != nil {
				return err
			}
			if len(res) == 0 {
				return fmt.Errorf("No channel found with id %v", channelID)
			}
			var channel types.Channel
			if err := cdc.UnmarshalBinaryLengthPrefixed(res, &channel); err != nil {
				return err
			}

			// Print result
			return cliCtx.PrintOutput(channel)
		},
	}
}

func GetCmd_GetSubmittedUpdate(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "update [paychan-id]",
		Args:  cobra.ExactArgs(1),
		Short: "get the latest update submitted to a channel",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Parse and validate input
			channelID, err := types.NewChannelIDFromString(args[0])
			if err != nil {
				return err
			}

			// Query the node
			res, err := cliCtx.QueryStore(types.GetSubmittedUpdateKey(channelID), storeKey)
			if err != nil {
				return err
			}
			if len(res) == 0 {
				return fmt.Errorf("No submitted update found for channel with id %v", channelID)
			}
			var sUpdate types.SubmittedUpdate
			if err := cdc.UnmarshalBinaryLengthPrefixed(res, &sUpdate); err != nil {
				return err
			}

			// Print result
			return cliCtx.PrintOutput(sUpdate)
		},
	}
}
