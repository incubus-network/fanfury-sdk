package v046_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/incubus-network/fanfury-sdk/v2/app"

	v046staking "github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/migrations/v046"
	"github.com/incubus-network/fanfury-sdk/v2/x/lsnative/staking/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := furyapp.MakeTestEncodingConfig()
	stakingKey := sdk.NewKVStoreKey("staking")
	tStakingKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(stakingKey, tStakingKey)
	paramstore := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, stakingKey, tStakingKey, "staking")

	// Check no params
	require.False(t, paramstore.Has(ctx, types.KeyMinCommissionRate))

	// Run migrations.
	err := v046staking.MigrateStore(ctx, stakingKey, encCfg.Marshaler, paramstore)
	require.NoError(t, err)

	// Make sure the new params are set.
	require.True(t, paramstore.Has(ctx, types.KeyMinCommissionRate))
}
