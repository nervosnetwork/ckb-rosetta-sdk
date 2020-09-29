package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SigningPayloadBuilder interface {
	BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error)
}

func NewSigningPayloadBuilderSecp256k1Blake160(txType string, signMessagesBuilder SignMessagesBuilder) *SigningPayloadBuilderSecp256k1Blake160 {
	return &SigningPayloadBuilderSecp256k1Blake160{txType, signMessagesBuilder}
}

type SigningPayloadBuilderSecp256k1Blake160 struct {
	TxType              string
	signMessagesBuilder SignMessagesBuilder
}

func (b SigningPayloadBuilderSecp256k1Blake160) BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error) {
	payloads := make([]*types.SigningPayload, 0)
	indexGroups, err := BuildIndexGroups(inputOperations)
	messages, err := b.signMessagesBuilder.BuildSignMessages(unsignedTx, inputOperations)
	if err != nil {
		return nil, err
	}

	for i, message := range messages {
		index := indexGroups[i][0]
		operation := inputOperations[index]
		payloads = append(payloads, &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{
				Address: operation.Account.Address,
			},
			Bytes:         message,
			SignatureType: types.EcdsaRecovery,
		})
	}

	return payloads, nil
}
