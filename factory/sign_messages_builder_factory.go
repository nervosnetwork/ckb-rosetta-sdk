package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SignMessagesBuilderFactory struct{}

func (f SignMessagesBuilderFactory) CreateSignMessagesBuilder(txType string) builder.SignMessagesBuilderInterface {
	switch txType {
	case ckb.Secp256k1Tx:
		return builder.NewSecp256k1Blake160SignMessagesBuilder()
	default:
		return nil
	}
}
