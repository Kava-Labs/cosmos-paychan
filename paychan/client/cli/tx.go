package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kava-labs/cosmos-paychan/paychan/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Paychan transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       utils.ValidateCmd,
	}

	txCmd.AddCommand(client.PostCommands(
		GetCmd_CreateChannel(cdc),
		GetCmd_SubmitPayment(cdc),
		GetCmd_GeneratePayment(cdc),
	)...)

	return txCmd
}

func GetCmd_CreateChannel(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create [receiver-address] [amount]",
		Short: "Create a new payment channel",
		Long:  "Create a new unidirectional payment channel from a local address to a remote address, funded with some amount of coins. These coins are removed from the sender account and put into the channel.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create cli helpers
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Parse inputs
			receiverAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}

			// Create msg
			senderAddr := cliCtx.GetFromAddress()
			msg := types.MsgCreate{
				Participants: [2]sdk.AccAddress{senderAddr, receiverAddr},
				Coins:        amount,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate tx and maybe sign and broadcast to blockchain
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmd_SubmitPayment(cdc *codec.Codec) *cobra.Command {
	flagPaymentFile := "payment"

	cmd := &cobra.Command{
		Use:   "close",
		Short: "Submit a payment to the blockchain to close the channel.",
		Long:  fmt.Sprintf("Submit a payment to the blockchain to either close a channel immediately (if you are the receiver) or after a dispute period of %d blocks (if you are the sender).", types.ChannelDisputeTime),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create cli helpers
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Get the payment to be submitted to the blockchain
			bz, err := ioutil.ReadFile(viper.GetString(flagPaymentFile))
			if err != nil {
				return err
			}
			var update types.Update
			err = json.Unmarshal(bz, &update)
			if err != nil {
				return err
			}

			// Create msg
			msg := types.MsgSubmitUpdate{
				Update:    update,
				Submitter: cliCtx.GetFromAddress(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate tx and maybe sign and broadcast to blockchain
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagPaymentFile, "payment.json", "File to read the payment from.")
	return cmd
}

func GetCmd_GeneratePayment(cdc *codec.Codec) *cobra.Command {
	flagPaymentFile := "filename"

	cmd := &cobra.Command{
		Use:   "pay [channel-id] [sender-amount] [receiver-amount]",
		Short: "generate a new payment",
		Long: `Generate a payment file (json) to send to the receiver as a payment.
Specify the channel id, and the total coins to be received by the channel's sender and receiver when the channel is eventually closed.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			// Create cli helpers
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Parse inputs
			channelID, err := types.NewChannelIDFromString(args[0])
			if err != nil {
				return err
			}
			senderAmount, err := sdk.ParseCoins(args[1])
			if err != nil {
				return err
			}
			receiverAmount, err := sdk.ParseCoins(args[2])
			if err != nil {
				return err
			}

			// Create an update
			update := types.Update{
				ChannelID: channelID,
				Payout:    types.Payout{senderAmount, receiverAmount},
				// empty signature
			}

			// Sign the update
			name := cliCtx.GetFromName()
			passphrase, err := keys.GetPassphrase(name)
			if err != nil {
				return err
			}
			bz := update.GetSignBytes()
			sig, pubKey, err := txBldr.Keybase().Sign(name, passphrase, bz)
			if err != nil {
				return err
			}
			update.Sigs = [1]types.UpdateSignature{{
				PubKey:          pubKey,
				CryptoSignature: sig,
			}}

			// Write out the update
			// TODO can this use the cli helpers? Can it be printed to stdOut instead?
			jsonUpdate, err := codec.MarshalJSONIndent(cdc, update)
			if err != nil {
				return err
			}
			paymentFile := viper.GetString(flagPaymentFile)
			err = ioutil.WriteFile(paymentFile, jsonUpdate, 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Written payment out to %v.\n", paymentFile)

			return nil
		},
	}
	cmd.Flags().String(flagPaymentFile, "payment.json", "File name to write the payment into.")
	return cmd
}
