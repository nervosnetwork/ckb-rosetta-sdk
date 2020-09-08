package builder

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

// UnsignedTxBuilder is an interface for different script unsignedTxBuilder
type UnsignedTxBuilder interface {
	BuildVersion() hexutil.Uint
	BuildCellDeps() []ckbTypes.CellDep
	BuildHeaderDeps() []ckbTypes.Hash
	BuildInputs() []ckbTypes.CellInput
	BuildOutputs() []ckbTypes.CellOutput
	BuildOutputsData() [][]byte
	BuildWitnesses() [][]byte
}
