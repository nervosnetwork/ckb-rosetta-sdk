package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/factory"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SigningPayloadBuilderFactory struct{}

func (f SigningPayloadBuilderFactory) CreateSigningPayloadBuilder(txType string) SigningPayloadBuilderInterface {
	switch txType {
	case ckb.Secp256k1Tx:
		return NewSecp256k1Blake160SigningPayloadBuilder(txType)
	default:
		return nil
	}
}

type SigningPayloadBuilderInterface interface {
	BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error)
}

func NewSecp256k1Blake160SigningPayloadBuilder(txType string) *Secp256k1Blake160SigningPayloadBuilder {
	return &Secp256k1Blake160SigningPayloadBuilder{txType}
}

type Secp256k1Blake160SigningPayloadBuilder struct {
	TxType string
}

func (b Secp256k1Blake160SigningPayloadBuilder) BuildSigningPayload(inputOperations []*types.Operation, unsignedTx *ckbTypes.Transaction) ([]*types.SigningPayload, error) {
	payloads := make([]*types.SigningPayload, 0)
	indexGroups, err := builder.BuildIndexGroups(inputOperations)
	sf := factory.SignMessagesBuilderFactory{}
	signMessagesBuilder := sf.CreateSignMessagesBuilder(b.TxType)
	messages, err := signMessagesBuilder.BuildSignMessages(unsignedTx, inputOperations)
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
