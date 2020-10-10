package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SignedTxBuilder struct{}

func (u SignedTxBuilder) CreateSignedTxBuilder(constructionType string) builder.SignedTxBuilder {
	switch constructionType {
	case ckb.TransferCKB:
		return builder.NewSignedTxCombinerSecp256k1Blake160()
	default:
		return nil
	}
}
