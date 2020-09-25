package factory

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/builder"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
)

type UnsignedTxBuilderFactory struct{}

func (f UnsignedTxBuilderFactory) CreateUnsignedTxBuilder(txType string, cfg *config.Config, inputOperations []*types.Operation, outputOperations []*types.Operation) builder.UnsignedTxBuilderInterface {
	switch txType {
	case ckb.Secp256k1Tx:
		return builder.NewUnsignedTxBuilderSecp256k1(cfg, inputOperations, outputOperations)
	default:
		return nil
	}
}
