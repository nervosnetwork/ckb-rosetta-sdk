package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

// UnsignedTxBuilder is an interface for different script unsignedTxBuilder
type UnsignedTxBuilder interface {
	BuildVersion() (hexutil.Uint, *types.Error)
	BuildCellDeps() ([]ckbTypes.CellDep, *types.Error)
	BuildHeaderDeps() ([]ckbTypes.Hash, *types.Error)
	BuildInputs() ([]ckbTypes.CellInput, *types.Error)
	BuildOutputs() ([]ckbTypes.CellOutput, *types.Error)
	BuildOutputsData() ([][]byte, *types.Error)
	BuildWitnesses() ([][]byte, *types.Error)
}
