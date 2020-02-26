package send

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/modules/hub/share/constants"
)

const (
	DefaultCodespace sdkTypes.CodespaceType = constants.ModuleName

	IncorrectMessageCode sdkTypes.CodeType = 101
)

func IncorrectMessageError(errorMessage string) sdkTypes.Error {
	return sdkTypes.NewError(DefaultCodespace, IncorrectMessageCode, errorMessage)
}