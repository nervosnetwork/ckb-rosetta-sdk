package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SigningPayloadBuilder interface {
	BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error)
}

func NewSigningPayloadBuilderSecp256k1Blake160(constructionType string, signMessagesBuilder SignMessagesBuilder) *SigningPayloadBuilderSecp256k1Blake160 {
	return &SigningPayloadBuilderSecp256k1Blake160{constructionType, signMessagesBuilder}
}

type SigningPayloadBuilderSecp256k1Blake160 struct {
	ConstructionType    string
	signMessagesBuilder SignMessagesBuilder
}

func (b SigningPayloadBuilderSecp256k1Blake160) BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error) {
	payloads := make([]*types.SigningPayload, 0)
	indexGroups, err := BuildIndexGroups(inputOperations)
	messages, err := b.signMessagesBuilder.BuildSignMessages(unsignedTx, inputOperations)
	metadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
		LockType: ckb.Secp256k1Blake160Lock.String(),
	})
	if err != nil {
		return nil, err
	}

	for i, message := range messages {
		index := indexGroups[i][0]
		operation := inputOperations[index]
		payloads = append(payloads, &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{
				Address:  operation.Account.Address,
				Metadata: metadata,
			},
			Bytes:         message,
			SignatureType: types.EcdsaRecovery,
		})
	}

	return payloads, nil
}
