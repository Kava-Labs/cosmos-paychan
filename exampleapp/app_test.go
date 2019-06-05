package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAppExport(t *testing.T) {
	db := db.NewMemDB()
	papp := NewPaychanApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	setGenesis(papp)

	// Making a new app object with the db, so that initchain hasn't been called
	newPapp := NewPaychanApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, 0)
	_, _, err := newPapp.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

func setGenesis(papp *PaychanApp) error {

	genesisState := NewDefaultGenesisState()
	stateBytes, err := codec.MarshalJSONIndent(papp.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	papp.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	papp.Commit()
	return nil
}
