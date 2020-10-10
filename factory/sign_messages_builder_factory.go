package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SignMessagesBuilderFactory struct{}

func (f SignMessagesBuilderFactory) CreateSignMessagesBuilder(constructionType string) builder.SignMessagesBuilder {
	switch constructionType {
	case ckb.TransferCKB:
		return builder.NewSignMessagesBuilderSecp256k1Blake160()
	default:
		return nil
	}
}
