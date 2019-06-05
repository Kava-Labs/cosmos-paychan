package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan"
 "github.com/kava-labs/cosmos-sdk-paychan/paychan/client/cli"
)

type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   paychan.ModuleName,
		Short: "Querying commands for the paychan module",
	}

	queryCmd.AddCommand(client.GetCommands(
			cli.GetCmd_QueryChannel(mc.cdc),
		)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   paychan.ModuleName,
		Short: "Paychan transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		cli.GetCmd_CreateChannel(mc.cdc),
		cli.GetCmd_SubmitPayment(mc.cdc),
	)...)

	// TODO where do these go?
	// GeneratePaymentCmd
	// VerifyPaymentCmd

	return txCmd
}
