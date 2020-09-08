package builder

import (
	"encoding/base64"
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/services"
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
	var cellDepArr []ckbTypes.CellDep
	cellDeps, _ := services.ValidateCellDeps(b.Operations)
	for _, cellDep := range cellDeps {
		cellDepArr = append(cellDepArr, cellDep)
	}
	return cellDepArr
}

func (b UnsignedTxBuilderSecp256k1) BuildHeaderDeps() []ckbTypes.Hash {
	mHeaderDeps, ok := b.Metadata["header_deps"]
	if !ok || mHeaderDeps == nil {
		return []ckbTypes.Hash{}
	}
	decodedHeaderDeps, err := base64.StdEncoding.DecodeString(mHeaderDeps.(string))
	if err != nil {
		return []ckbTypes.Hash{}
	}
	var headerDeps []ckbTypes.Hash
	err = json.Unmarshal(decodedHeaderDeps, &headerDeps)
	if err != nil {
		return []ckbTypes.Hash{}
	}
	return headerDeps
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
