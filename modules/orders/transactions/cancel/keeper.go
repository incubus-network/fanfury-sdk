package cancel

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/constants"
	"github.com/persistenceOne/persistenceSDK/modules/exchanges/auxiliaries/reverse"
	"github.com/persistenceOne/persistenceSDK/modules/identities/auxiliaries/verify"
	"github.com/persistenceOne/persistenceSDK/modules/orders/mapper"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
)

type transactionKeeper struct {
	mapper                    helpers.Mapper
	identitiesVerifyAuxiliary helpers.Auxiliary
	exchangesReverseAuxiliary helpers.Auxiliary
}

var _ helpers.TransactionKeeper = (*transactionKeeper)(nil)

func (transactionKeeper transactionKeeper) Transact(context sdkTypes.Context, msg sdkTypes.Msg) error {
	message := messageFromInterface(msg)
	orders := mapper.NewOrders(transactionKeeper.mapper, context).Fetch(message.OrderID)
	order := orders.Get(message.OrderID)
	if order == nil {
		return constants.EntityNotFound
	}

	makerID := base.NewID(order.GetImmutables().Get().Get(base.NewID(constants.MakerIDProperty)).GetFact().String())
	makerSplitID := base.NewID(order.GetImmutables().Get().Get(base.NewID(constants.MakerSplitIDProperty)).GetFact().String())
	makerSplit, Error := sdkTypes.NewDecFromStr(order.GetMutables().Get().Get(base.NewID(constants.MakerSplitProperty)).GetFact().String())
	if Error != nil {
		return Error
	}
	if Error := transactionKeeper.identitiesVerifyAuxiliary.GetKeeper().Help(context,
		verify.NewAuxiliaryRequest(message.From, makerID)); Error != nil {
		return Error
	}

	if Error := transactionKeeper.exchangesReverseAuxiliary.GetKeeper().Help(context,
		reverse.NewAuxiliaryRequest(makerID, makerSplit, makerSplitID)); Error != nil {
		return Error
	}
	orders.Remove(order)
	return nil
}

func initializeTransactionKeeper(mapper helpers.Mapper, externalKeepers []interface{}) helpers.TransactionKeeper {
	transactionKeeper := transactionKeeper{mapper: mapper}
	for _, externalKeeper := range externalKeepers {
		switch value := externalKeeper.(type) {
		case helpers.Auxiliary:
			switch value.GetName() {
			case verify.Auxiliary.GetName():
				transactionKeeper.identitiesVerifyAuxiliary = value
			case reverse.Auxiliary.GetName():
				transactionKeeper.exchangesReverseAuxiliary = value
			}
		}
	}
	return transactionKeeper
}