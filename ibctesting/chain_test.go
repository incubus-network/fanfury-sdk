package ibctesting_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	ibctesting "github.com/incubus-network/fanfury-sdk/v2/ibctesting"
)

func TestChangeValSet(t *testing.T) {
	coord := ibctesting.NewCoordinator(t, 2)
	chainA := coord.GetChain(ibctesting.GetChainID(1))
	chainB := coord.GetChain(ibctesting.GetChainID(2))

	path := ibctesting.NewPath(chainA, chainB)
	coord.Setup(path)

	amount, ok := sdk.NewIntFromString("10000000000000000000")
	require.True(t, ok)
	amount2, ok := sdk.NewIntFromString("30000000000000000000")
	require.True(t, ok)

	val := chainA.GetFuryApp().StakingKeeper.GetValidators(chainA.GetContext(), 4)

	chainA.GetFuryApp().StakingKeeper.Delegate(chainA.GetContext(), chainA.SenderAccounts[1].SenderAccount.GetAddress(),
		amount, sdkstaking.Unbonded, val[1], true)
	chainA.GetFuryApp().StakingKeeper.Delegate(chainA.GetContext(), chainA.SenderAccounts[3].SenderAccount.GetAddress(),
		amount2, sdkstaking.Unbonded, val[3], true)

	coord.CommitBlock(chainA)

	// verify that update clients works even after validator update goes into effect
	path.EndpointB.UpdateClient()
	path.EndpointB.UpdateClient()
}
