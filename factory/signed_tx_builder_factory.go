package factory

import (
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
)

type SignedTxBuilder struct{}

func (u SignedTxBuilder) CreateSignedTxBuilder(txType string) builder.SignedTxBuilder {
	switch txType {
	case ckb.Secp256k1Tx:
		return builder.NewSignedTxCombinerSecp256k1Blake160()
	default:
		return nil
	}
}
