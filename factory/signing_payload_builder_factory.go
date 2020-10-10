package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SigningPayloadBuilderFactory struct{}

func (f SigningPayloadBuilderFactory) CreateSigningPayloadBuilder(constructionType string) builder.SigningPayloadBuilder {
	switch constructionType {
	case ckb.TransferCKB:
		sf := SignMessagesBuilderFactory{}
		signMessagesBuilder := sf.CreateSignMessagesBuilder(constructionType)
		return builder.NewSigningPayloadBuilderSecp256k1Blake160(constructionType, signMessagesBuilder)
	default:
		return nil
	}
}
