package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SigningPayloadBuilderFactory struct{}

func (f SigningPayloadBuilderFactory) CreateSigningPayloadBuilder(txType string) builder.SigningPayloadBuilderInterface {
	switch txType {
	case ckb.Secp256k1Tx:
		sf := SignMessagesBuilderFactory{}
		signMessagesBuilder := sf.CreateSignMessagesBuilder(txType)
		return builder.NewSigningPayloadBuilderSecp256k1Blake160(txType, signMessagesBuilder)
	default:
		return nil
	}
}
