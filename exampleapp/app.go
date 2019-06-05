package app

import (
	"io"
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/bank"

	//"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/cosmos-sdk-paychan/paychan"
)

const appName = "PaychanApp"

var (
	// default home directories for paychancli
	DefaultCLIHome = os.ExpandEnv("$HOME/.paychancli")

	// default home directories for paychand
	DefaultNodeHome = os.ExpandEnv("$HOME/.paychand")

	// The ModuleBasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics sdk.ModuleBasicManager
)

func init() {
	ModuleBasics = sdk.NewModuleBasicManager(
		genaccounts.AppModuleBasic{}, // TODO is this needed here?
		//genutil.AppModuleBasic{}, // TODO is this needed here?
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		//staking.AppModuleBasic{},
		//mint.AppModuleBasic{},
		//distr.AppModuleBasic{},
		//gov.AppModuleBasic{},
		params.AppModuleBasic{},
		//crisis.AppModuleBasic{},
		//slashing.AppModuleBasic{},
		paychan.AppModuleBasic{},
	)
}

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// Extended ABCI application
type PaychanApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint // remove?

	// keys to access the substores
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	//keyStaking       *sdk.KVStoreKey
	//tkeyStaking      *sdk.TransientStoreKey
	//keySlashing      *sdk.KVStoreKey
	//keyMint          *sdk.KVStoreKey
	//keyDistr         *sdk.KVStoreKey
	//tkeyDistr        *sdk.TransientStoreKey
	//keyGov           *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey
	keyPaychan       *sdk.KVStoreKey

	// keepers
	accountKeeper       auth.AccountKeeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	bankKeeper          bank.Keeper
	//stakingKeeper       staking.Keeper
	//slashingKeeper      slashing.Keeper
	//mintKeeper          mint.Keeper
	//distrKeeper         distr.Keeper
	//govKeeper           gov.Keeper
	//crisisKeeper        crisis.Keeper
	paramsKeeper  params.Keeper
	paychanKeeper paychan.Keeper

	// the module manager
	mm *sdk.ModuleManager
}

// NewPaychanApp returns a reference to an initialized PaychanApp.
func NewPaychanApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp)) *PaychanApp {

	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	var app = &PaychanApp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keyMain:        sdk.NewKVStoreKey(bam.MainStoreKey),
		keyAccount:     sdk.NewKVStoreKey(auth.StoreKey),
		//keyStaking:       sdk.NewKVStoreKey(staking.StoreKey),
		//tkeyStaking:      sdk.NewTransientStoreKey(staking.TStoreKey),
		//keyMint:          sdk.NewKVStoreKey(mint.StoreKey),
		//keyDistr:         sdk.NewKVStoreKey(distr.StoreKey),
		//tkeyDistr:        sdk.NewTransientStoreKey(distr.TStoreKey),
		//keySlashing:      sdk.NewKVStoreKey(slashing.StoreKey),
		//keyGov:           sdk.NewKVStoreKey(gov.StoreKey),
		keyFeeCollection: sdk.NewKVStoreKey(auth.FeeStoreKey),
		keyParams:        sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:       sdk.NewTransientStoreKey(params.TStoreKey),
		keyPaychan:       sdk.NewKVStoreKey(paychan.StoreKey),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams, params.DefaultCodespace)
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)
	//stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
	//mintSubspace := app.paramsKeeper.Subspace(mint.DefaultParamspace)
	//distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	//slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	//govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	//crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.keyAccount, authSubspace, auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(app.accountKeeper, bankSubspace, bank.DefaultCodespace)
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)
	//stakingKeeper := staking.NewKeeper(app.cdc, app.keyStaking, app.tkeyStaking, app.bankKeeper,
	//	stakingSubspace, staking.DefaultCodespace)
	//app.mintKeeper = mint.NewKeeper(app.cdc, app.keyMint, mintSubspace, &stakingKeeper, app.feeCollectionKeeper)
	//app.distrKeeper = distr.NewKeeper(app.cdc, app.keyDistr, distrSubspace, app.bankKeeper, &stakingKeeper,
	//	app.feeCollectionKeeper, distr.DefaultCodespace)
	//app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, &stakingKeeper,
	//	slashingSubspace, slashing.DefaultCodespace)
	//app.crisisKeeper = crisis.NewKeeper(crisisSubspace, invCheckPeriod, app.distrKeeper,
	//	app.bankKeeper, app.feeCollectionKeeper)
	app.paychanKeeper = paychan.NewKeeper(app.cdc, app.keyPaychan, app.bankKeeper)

	// register the proposal types
	//govRouter := gov.NewRouter()
	//govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
	//	AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
	//	AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper))
	//app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.paramsKeeper, govSubspace,
	//	app.bankKeeper, &stakingKeeper, gov.DefaultCodespace, govRouter)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	//app.stakingKeeper = *stakingKeeper.SetHooks(
	//	staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()))

	app.mm = sdk.NewModuleManager(
		genaccounts.NewAppModule(app.accountKeeper),
		//genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper, app.feeCollectionKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		//crisis.NewAppModule(app.crisisKeeper, app.Logger()),
		//distr.NewAppModule(app.distrKeeper),
		//gov.NewAppModule(app.govKeeper),
		//mint.NewAppModule(app.mintKeeper),
		//slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		//staking.NewAppModule(app.stakingKeeper, app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
		paychan.NewAppModule(app.paychanKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	//app.mm.SetOrderBeginBlockers(mint.ModuleName, distr.ModuleName, slashing.ModuleName)

	//app.mm.SetOrderEndBlockers(gov.ModuleName, staking.ModuleName)
	app.mm.SetOrderEndBlockers(paychan.ModuleName) // TODO can this be removed?

	// genutils must occur after staking so that pools are properly
	// initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(genaccounts.ModuleName, /*distr.ModuleName,
		staking.ModuleName,*/auth.ModuleName, bank.ModuleName, /*slashing.ModuleName,
		gov.ModuleName, mint.ModuleName, crisis.ModuleName, genutil.ModuleName*/)

	//app.mm.RegisterInvariants(&app.crisisKeeper)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize stores
	app.MountStores(app.keyMain, app.keyAccount, /*app.keyStaking, app.keyMint,
		app.keyDistr, app.keySlashing, app.keyGov, */app.keyFeeCollection,
		app.keyParams, app.tkeyParams /*app.tkeyStaking, app.tkeyDistr*/, app.keyPaychan)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper, auth.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keyMain)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}
	return app
}

// application updates every begin block
func (app *PaychanApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// TODO remove Blocker Functions
// application updates every end block
func (app *PaychanApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// application update at chain initialization
func (app *PaychanApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	return app.mm.InitGenesis(ctx, genesisState)
}

// load a particular height
func (app *PaychanApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keyMain)
}
