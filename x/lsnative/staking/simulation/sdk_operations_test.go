package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	furyappparams "github.com/incubus-network/fanfury-sdk/v2/furyapp/params"
	distrtypes "github.com/incubus-network/fanfury-sdk/v2/x/lsnative/distribution/types"
	"github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/simulation"
	"github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/types"
)

// TestSdkWeightedOperations tests the weights of the operations for sdkstaking types.
func TestSdkWeightedOperations(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accs := createTestApp(t, false, r, 3)

	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.SdkWeightedOperations(appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.StakingKeeper,
	)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{furyappparams.DefaultWeightMsgCreateValidator, sdkstaking.ModuleName, sdkstaking.TypeMsgCreateValidator},
		{furyappparams.DefaultWeightMsgEditValidator, sdkstaking.ModuleName, sdkstaking.TypeMsgEditValidator},
		{furyappparams.DefaultWeightMsgDelegate, sdkstaking.ModuleName, sdkstaking.TypeMsgDelegate},
		{furyappparams.DefaultWeightMsgUndelegate, sdkstaking.ModuleName, sdkstaking.TypeMsgUndelegate},
		{furyappparams.DefaultWeightMsgBeginRedelegate, sdkstaking.ModuleName, sdkstaking.TypeMsgBeginRedelegate},
		{furyappparams.DefaultWeightMsgCancelUnbondingDelegation, sdkstaking.ModuleName, sdkstaking.TypeMsgCancelUnbondingDelegation},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateSdkMsgCreateValidator tests the normal scenario of a valid message of type TypeMsgCreateValidator.
// Abonormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateSdkMsgCreateValidator(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateSdkMsgCreateValidator(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgCreateValidator
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "0.080000000000000000", msg.Commission.MaxChangeRate.String())
	require.Equal(t, "0.080000000000000000", msg.Commission.MaxRate.String())
	require.Equal(t, "0.019527679037870745", msg.Commission.Rate.String())
	require.Equal(t, sdkstaking.TypeMsgCreateValidator, msg.Type())
	require.Equal(t, []byte{0xa, 0x20, 0x51, 0xde, 0xbd, 0xe8, 0xfa, 0xdf, 0x4e, 0xfc, 0x33, 0xa5, 0x16, 0x94, 0xf6, 0xee, 0xd3, 0x69, 0x7a, 0x7a, 0x1c, 0x2d, 0x50, 0xb6, 0x2, 0xf7, 0x16, 0x4e, 0x66, 0x9f, 0xff, 0x38, 0x91, 0x9b}, msg.Pubkey.Value)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	require.Equal(t, "cosmosvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnsvnaes", msg.ValidatorAddress)
	require.Len(t, futureOperations, 0)
}

// TestSimulateSdkMsgCancelUnbondingDelegation tests the normal scenario of a valid message of type TypeMsgCancelUnbondingDelegation.
// Abonormal scenarios, where the message is
func TestSimulateSdkMsgCancelUnbondingDelegation(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// setup delegation
	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := types.NewDelegation(delegator.Address, validator0.GetOperator(), issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.GetOperator())

	// unbonding delegation
	udb := types.NewUnbondingDelegation(delegator.Address, validator0.GetOperator(), app.LastBlockHeight(), blockTime.Add(2*time.Minute), delTokens)
	app.StakingKeeper.SetUnbondingDelegation(ctx, udb)
	setupValidatorRewards(app, ctx, validator0.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateSdkMsgCancelUnbondingDelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	accounts = []simtypes.Account{accounts[1]}
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgCancelUnbondingDelegation
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgCancelUnbondingDelegation, msg.Type())
	require.Equal(t, delegator.Address.String(), msg.DelegatorAddress)
	require.Equal(t, validator0.GetOperator().String(), msg.ValidatorAddress)
	require.Len(t, futureOperations, 0)
}

// TestSimulateSdkMsgEditValidator tests the normal scenario of a valid message of type TypeMsgEditValidator.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateSdkMsgEditValidator(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator
	_ = getTestingValidator0(t, app, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateSdkMsgEditValidator(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgEditValidator
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "0.280623462081924936", msg.CommissionRate.String())
	require.Equal(t, "xKGLwQvuyN", msg.Description.Moniker)
	require.Equal(t, "SlcxgdXhhu", msg.Description.Identity)
	require.Equal(t, "WeLrQKjLxz", msg.Description.Website)
	require.Equal(t, "rBqDOTtGTO", msg.Description.SecurityContact)
	require.Equal(t, types.TypeMsgEditValidator, msg.Type())
	require.Equal(t, "cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(t, futureOperations, 0)
}

// TestSimulateSdkMsgDelegate tests the normal scenario of a valid message of type TypeMsgDelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateSdkMsgDelegate(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// execute operation
	op := simulation.SimulateSdkMsgDelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgDelegate
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	require.Equal(t, "98100858108421259236", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgDelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress)
	require.Len(t, futureOperations, 0)
}

// TestSimulateSdkMsgUndelegate tests the normal scenario of a valid message of type TypeMsgUndelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateSdkMsgUndelegate(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// setup delegation
	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := types.NewDelegation(delegator.Address, validator0.GetOperator(), issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateSdkMsgUndelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgUndelegate
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	require.Equal(t, "280623462081924937", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgUndelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(t, futureOperations, 0)
}

// TestSimulateSdkMsgBeginRedelegate tests the normal scenario of a valid message of type TypeMsgBeginRedelegate.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateSdkMsgBeginRedelegate(t *testing.T) {
	s := rand.NewSource(12)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 4)

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator0 and accounts[1] as validator1
	validator0 := getTestingValidator0(t, app, ctx, accounts)
	validator1 := getTestingValidator1(t, app, ctx, accounts)

	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)

	// setup accounts[2] as delegator
	delegator := accounts[2]
	delegation := types.NewDelegation(delegator.Address, validator1.GetOperator(), issuedShares, false)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator1.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.GetOperator())
	setupValidatorRewards(app, ctx, validator1.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateSdkMsgBeginRedelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg sdkstaking.MsgBeginRedelegate
	sdkstaking.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1092v0qgulpejj8y8hs6dmlw82x9gv8f7jfc7jl", msg.DelegatorAddress)
	require.Equal(t, "1883752832348281252", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgBeginRedelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1gnkw3uqzflagcqn6ekjwpjanlne928qhruemah", msg.ValidatorDstAddress)
	require.Equal(t, "cosmosvaloper1kk653svg7ksj9fmu85x9ygj4jzwlyrgs89nnn2", msg.ValidatorSrcAddress)
	require.Len(t, futureOperations, 0)
}
