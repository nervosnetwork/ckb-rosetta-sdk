package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
)

var _ UnsignedTxBuilder = UnsignedTxBuilderSecp256k1{}

type UnsignedTxBuilderSecp256k1 struct {
	Operations []*types.Operation
	Inputs     []ckbTypes.CellInfo
	Metadata   map[string]interface{}
}

func (b UnsignedTxBuilderSecp256k1) BuildVersion() hexutil.Uint {
	var defaultVersion uint
	version, ok := b.Metadata["version"]
	if !ok || version == nil {
		return hexutil.Uint(defaultVersion)
	}
	strVersion := version.(string)
	uVersion, err := strconv.ParseUint(strVersion, 10, 64)
	if err != nil {
		return hexutil.Uint(defaultVersion)
	}
	return hexutil.Uint(uVersion)
}

func (b UnsignedTxBuilderSecp256k1) BuildCellDeps() []ckbTypes.CellDep {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildHeaderDeps() []ckbTypes.Hash {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputs() []ckbTypes.CellOutput {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputsData() [][]byte {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildWitnesses() [][]byte {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildInputs() []ckbTypes.CellInput {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) GetResult() ckbTypes.Transaction {
	panic("implement me")
}
