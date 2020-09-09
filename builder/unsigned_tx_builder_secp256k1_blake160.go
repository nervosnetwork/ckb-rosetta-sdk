package builder

import (
	"encoding/base64"
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/services"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
)

var _ UnsignedTxBuilder = UnsignedTxBuilderSecp256k1{}

type UnsignedTxBuilderSecp256k1 struct {
	Operations []*types.Operation
	Inputs     []ckbTypes.CellInfo
	Metadata   map[string]interface{}
}

func (b UnsignedTxBuilderSecp256k1) BuildVersion() (hexutil.Uint, *types.Error) {
	var defaultVersion uint
	strVersion, ok := b.Metadata["version"].(string)
	if !ok {
		return hexutil.Uint(defaultVersion), nil
	}
	uVersion, err := strconv.ParseUint(strVersion, 10, 64)
	if err != nil {
		return hexutil.Uint(defaultVersion), nil
	}
	return hexutil.Uint(uVersion), nil
}

func (b UnsignedTxBuilderSecp256k1) BuildCellDeps() ([]ckbTypes.CellDep, *types.Error) {
	var cellDepArr []ckbTypes.CellDep
	cellDeps, err := services.ValidateCellDeps(b.Operations)
	if err != nil {
		return nil, err
	}
	for _, cellDep := range cellDeps {
		cellDepArr = append(cellDepArr, cellDep)
	}
	return cellDepArr, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildHeaderDeps() ([]ckbTypes.Hash, *types.Error) {
	strHeaderDeps, ok := b.Metadata["header_deps"].(string)
	if !ok {
		return []ckbTypes.Hash{}, nil
	}
	decodedHeaderDeps, err := base64.StdEncoding.DecodeString(strHeaderDeps)
	if err != nil {
		return []ckbTypes.Hash{}, nil
	}
	var headerDeps []ckbTypes.Hash
	err = json.Unmarshal(decodedHeaderDeps, &headerDeps)
	if err != nil {
		return []ckbTypes.Hash{}, nil
	}
	return headerDeps, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputs() ([]ckbTypes.CellOutput, *types.Error) {
	outputOperations := services.OperationFilter(b.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Output"
	})
	var outputs []ckbTypes.CellOutput
	for _, operation := range outputOperations {
		capacity, err := strconv.ParseUint(operation.Amount.Value, 10, 64)
		if err != nil {
			return nil, services.InvalidOutputOperationAmountValueError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, services.AddressParseError
		}
		var typeScript *ckbTypes.Script
		strTypeScript, ok := operation.Metadata["type_script"].(string)
		if ok {
			decodedTypeScript, err := base64.StdEncoding.DecodeString(strTypeScript)
			if err != nil {
				return nil, services.InvalidTypeScriptError
			}
			err = json.Unmarshal(decodedTypeScript, typeScript)
			if err != nil {
				return nil, services.InvalidTypeScriptError
			}
		}
		outputs = append(outputs, ckbTypes.CellOutput{
			Capacity: capacity,
			Lock:     addr.Script,
			Type:     typeScript,
		})
	}
	return outputs, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputsData() ([][]byte, *types.Error) {
	outputOperations := services.OperationFilter(b.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Output"
	})
	var outputsData [][]byte
	for _, operation := range outputOperations {
		strOutputData, ok := operation.Metadata["output_data"].(string)
		if ok {
			decodedOutputData, err := base64.StdEncoding.DecodeString(strOutputData)
			if err != nil {
				return nil, services.InvalidOutputDataError
			}
			var outputData []byte
			err = json.Unmarshal(decodedOutputData, &outputData)
			if err != nil {
				return nil, services.InvalidOutputDataError
			}
			outputsData = append(outputsData, outputData)
		} else {
			outputsData = append(outputsData, []byte{})
		}
	}

	return outputsData, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildWitnesses() ([][]byte, *types.Error) {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) BuildInputs() ([]ckbTypes.CellInput, *types.Error) {
	panic("implement me")
}

func (b UnsignedTxBuilderSecp256k1) GetResult() (ckbTypes.Transaction, *types.Error) {
	panic("implement me")
}
