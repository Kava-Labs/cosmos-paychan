package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/client"

	//"github.com/kava-labs/cosmos-sdk-paychan/paychan/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "paychan", // or types.ModuleName
		Short: "Querying commands for the paychan module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}

	queryCmd.AddCommand(client.GetCommands( // TODO should this be a separate subcommand?
			GetCmd_QueryChannel(cdc),
		)...)

	return queryCmd
}

func GetCmd_QueryChannel(cdc *codec.Codec /*paychanStoreName string*/) *cobra.Command {
	flagId := "chan-id"
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get info on a channel.",
		Long:  "Get the details of a non closed channel plus any submitted update waiting to be executed.",
		Args:  cobra.NoArgs,
		// RunE: func(cmd *cobra.Command, args []string) error {

		// 	// Create a cli "context": struct populated with info from common flags.
		// 	cliCtx := context.NewCLIContext().
		// 		WithCodec(cdc).
		// 		WithLogger(os.Stdout).
		// 		WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

		// 	// Get channel ID
		// 	id := paychan.ChannelID(viper.GetInt64(flagId))

		// 	// Get the channel from the node
		// 	res, err := cliCtx.QueryStore(paychan.GetChannelKey(id), paychanStoreName)
		// 	if len(res) == 0 || err != nil {
		// 		return errors.Errorf("channel with ID '%d' does not exist", id)
		// 	}
		// 	var channel paychan.Channel
		// 	cdc.MustUnmarshalBinary(res, &channel)

		// 	// Convert the channel to a json object for pretty printing
		// 	jsonChannel, err := codec.MarshalJSONIndent(cdc, channel)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	// print out json channel
		// 	fmt.Println(string(jsonChannel))

		// 	// Get any submitted updates from the node
		// 	res, err = cliCtx.QueryStore(paychan.GetSubmittedUpdateKey(id), paychanStoreName)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	// Print out the submitted update if it exists
		// 	if len(res) != 0 {
		// 		var submittedUpdate paychan.SubmittedUpdate
		// 		cdc.MustUnmarshalBinary(res, &submittedUpdate)

		// 		// Convert the submitted update to a json object for pretty printing
		// 		jsonSU, err := codec.MarshalJSONIndent(cdc, submittedUpdate)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		// print out json submitted update
		// 		fmt.Println(string(jsonSU))
		// 	}
		// 	return nil
		// },
	}
	cmd.Flags().Int(flagId, 0, "ID of the payment channel.")
	return cmd
}
