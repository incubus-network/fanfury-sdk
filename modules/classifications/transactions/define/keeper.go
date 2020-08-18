/*
 Copyright [2019] - [2020], PERSISTENCE TECHNOLOGIES PTE. LTD. and the persistenceSDK contributors
 SPDX-License-Identifier: Apache-2.0
*/

package define

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/constants"
	"github.com/persistenceOne/persistenceSDK/modules/classifications/mapper"
	"github.com/persistenceOne/persistenceSDK/modules/identities/auxiliaries/verify"
	"github.com/persistenceOne/persistenceSDK/modules/metas/auxiliaries/scrub"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
)

type transactionKeeper struct {
	mapper          helpers.Mapper
	verifyAuxiliary helpers.Auxiliary
	scrubAuxiliary  helpers.Auxiliary
}

var _ helpers.TransactionKeeper = (*transactionKeeper)(nil)

func (transactionKeeper transactionKeeper) Transact(context sdkTypes.Context, msg sdkTypes.Msg) helpers.TransactionResponse {
	message := messageFromInterface(msg)
	if auxiliaryResponse := transactionKeeper.verifyAuxiliary.GetKeeper().Help(context, verify.NewAuxiliaryRequest(message.From, message.FromID)); !auxiliaryResponse.IsSuccessful() {
		return newTransactionResponse(auxiliaryResponse.GetError())
	}

	scrubImmutableMetaTraitsAuxiliaryResponse, Error := scrub.ValidateResponse(transactionKeeper.scrubAuxiliary.GetKeeper().Help(context, scrub.NewAuxiliaryRequest(message.ImmutableMetaTraits)))
	if Error != nil {
		return newTransactionResponse(Error)
	}
	immutableTraits := base.NewImmutables(base.NewProperties(append(scrubImmutableMetaTraitsAuxiliaryResponse.Properties.GetList(), message.ImmutableTraits.GetList()...)))

	scrubMutableMetaTraitsAuxiliaryResponse, Error := scrub.ValidateResponse(transactionKeeper.scrubAuxiliary.GetKeeper().Help(context, scrub.NewAuxiliaryRequest(message.MutableMetaTraits)))
	if Error != nil {
		return newTransactionResponse(Error)
	}
	mutableTraits := base.NewMutables(base.NewProperties(append(scrubMutableMetaTraitsAuxiliaryResponse.Properties.GetList(), message.MutableMetaTraits.GetList()...)))

	classificationID := mapper.NewClassificationID(base.NewID(context.ChainID()), immutableTraits, mutableTraits)
	classifications := mapper.NewClassifications(transactionKeeper.mapper, context).Fetch(classificationID)
	if classifications.Get(classificationID) != nil {
		return newTransactionResponse(constants.EntityAlreadyExists)
	}

	classifications = classifications.Add(mapper.NewClassification(classificationID, immutableTraits, mutableTraits))
	return newTransactionResponse(nil)
}

func initializeTransactionKeeper(mapper helpers.Mapper, auxiliaries []interface{}) helpers.TransactionKeeper {
	transactionKeeper := transactionKeeper{mapper: mapper}
	for _, auxiliary := range auxiliaries {
		switch value := auxiliary.(type) {
		case helpers.Auxiliary:
			switch value.GetName() {
			case scrub.Auxiliary.GetName():
				transactionKeeper.scrubAuxiliary = value
			case verify.Auxiliary.GetName():
				transactionKeeper.verifyAuxiliary = value
			}
		}
	}
	return transactionKeeper
}
