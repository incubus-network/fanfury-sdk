package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosTypes "github.com/persistenceOne/persistenceSDK/x/cosmos/types"
)

// InitGenesis new cosmos genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, data *cosmosTypes.GenesisState) {
	keeper.SetParams(ctx, data.Params)
	keeper.SetProposalID(ctx, 1)
	keeper.setID(ctx, 0, []byte(cosmosTypes.KeyLastTXPoolID))
	keeper.setTotalDelegatedAmountTillDate(ctx, sdk.Coin{})
	//keeper.SetVotingParams(ctx, data.Params.CosmosProposalParams)
	//TODO add remaining : Setup initial multisig account and custodial address (either with a proposal or through genesis initialisation
}
