package genesis

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	"github.com/persistenceOne/persistenceSDK/schema/mappables"
	"github.com/persistenceOne/persistenceSDK/schema/traits"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
)

//TODO define genesis state
type genesisState struct{
	IdentityList []mappables.InterNFT
}

var _ helpers.GenesisState = (*genesisState)(nil)

func (genesisState genesisState) Default() helpers.GenesisState {
	return genesisState
}

func (genesisState genesisState) Validate() error {

	return nil
}

func (genesisState genesisState) Initialize(ctx sdkTypes.Context, mapper helpers.Mapper) {

	for _, identity := range genesisState.IdentityList {
		mapper.Create(ctx, identity)
	}
}

func (genesisState genesisState) Export(context sdkTypes.Context, mapper helpers.Mapper) helpers.GenesisState {
	assetsID := base.NewID("")

	appendableAssetList := func(mappable traits.Mappable) bool {
		genesisState.IdentityList = append(genesisState.IdentityList, mappable.(mappables.InterNFT))
		return false
	}
	mapper.Iterate(context, assetsID, appendableAssetList)
	return genesisState
}


func (genesisState genesisState) Marshall() []byte {
	return PackageCodec.MustMarshalJSON(genesisState)
}
func (genesisState genesisState) Unmarshall(byte []byte) helpers.GenesisState {
	if Error := PackageCodec.UnmarshalJSON(byte, &genesisState); Error != nil {
		return nil
	}
	return genesisState
}

func newGenesisState(identityList []mappables.InterNFT) helpers.GenesisState {
	return genesisState{
		IdentityList: identityList,
	}
}

var GenesisState = newGenesisState([]mappables.InterNFT{})