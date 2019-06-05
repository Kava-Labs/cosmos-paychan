package paychan

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const ModuleName = "paychan"
const routerKey = ModuleName
const StoreKey = ModuleName

// ---------- AppModuleBasic ----------

// AppModuleBasic
type AppModuleBasic struct{}

// check it implements the interface at compile time
var _ sdk.AppModuleBasic = AppModuleBasic{}

// module name
func (AppModuleBasic) Name() string { return ModuleName }

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) { RegisterCodec(cdc) }

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage { return nil }

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error { return nil }

// ---------- AppModule ----------

// AppModule
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// module name
func (AppModule) Name() string { return ModuleName }

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRouter) {}

// module message route name
func (AppModule) Route() string {
	return routerKey
}

// module handler
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// TODO add querier
// module querier route name
func (AppModule) QuerierRoute() string { return "" }

// module querier
func (AppModule) NewQuerierHandler() sdk.Querier { return nil }

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, genesis json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage { return nil }

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) sdk.Tags { return sdk.EmptyTags() }

// module end-block
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) ([]abci.ValidatorUpdate, sdk.Tags) {
	tags := EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}, tags
}
