package services

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ququzone/ckb-sdk-go/types"
)

type outPoint struct {
	TxHash types.Hash   `json:"tx_hash"`
	Index  hexutil.Uint `json:"index"`
}

type cellDep struct {
	OutPoint outPoint      `json:"out_point"`
	DepType  types.DepType `json:"dep_type"`
}

type cellInput struct {
	Since          hexutil.Uint64 `json:"since"`
	PreviousOutput outPoint       `json:"previous_output"`
}

type script struct {
	CodeHash types.Hash           `json:"code_hash"`
	HashType types.ScriptHashType `json:"hash_type"`
	Args     hexutil.Bytes        `json:"args"`
}

type cellOutput struct {
	Capacity hexutil.Uint64 `json:"capacity"`
	Lock     *script        `json:"lock"`
	Type     *script        `json:"type"`
}

type transaction struct {
	Version     hexutil.Uint    `json:"version"`
	Hash        types.Hash      `json:"hash"`
	CellDeps    []cellDep       `json:"cell_deps"`
	HeaderDeps  []types.Hash    `json:"header_deps"`
	Inputs      []cellInput     `json:"inputs"`
	Outputs     []cellOutput    `json:"outputs"`
	OutputsData []hexutil.Bytes `json:"outputs_data"`
	Witnesses   []hexutil.Bytes `json:"witnesses"`
}

type inTransaction struct {
	Version     hexutil.Uint    `json:"version"`
	CellDeps    []cellDep       `json:"cell_deps"`
	HeaderDeps  []types.Hash    `json:"header_deps"`
	Inputs      []cellInput     `json:"inputs"`
	Outputs     []cellOutput    `json:"outputs"`
	OutputsData []hexutil.Bytes `json:"outputs_data"`
	Witnesses   []hexutil.Bytes `json:"witnesses"`
}

func ToTransaction(data string) (*types.Transaction, error) {
	var tx transaction
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		return nil, err
	}
	return &types.Transaction{
		Version:     uint(tx.Version),
		Hash:        tx.Hash,
		CellDeps:    toCellDeps(tx.CellDeps),
		HeaderDeps:  tx.HeaderDeps,
		Inputs:      toInputs(tx.Inputs),
		Outputs:     toOutputs(tx.Outputs),
		OutputsData: toBytesArray(tx.OutputsData),
		Witnesses:   toBytesArray(tx.Witnesses),
	}, nil
}

func toBytesArray(bytes []hexutil.Bytes) [][]byte {
	result := make([][]byte, len(bytes))
	for i, data := range bytes {
		result[i] = data
	}
	return result
}

func toOutputs(outputs []cellOutput) []*types.CellOutput {
	result := make([]*types.CellOutput, len(outputs))
	for i := 0; i < len(outputs); i++ {
		output := outputs[i]
		result[i] = &types.CellOutput{
			Capacity: uint64(output.Capacity),
			Lock: &types.Script{
				CodeHash: output.Lock.CodeHash,
				HashType: output.Lock.HashType,
				Args:     output.Lock.Args,
			},
		}
		if output.Type != nil {
			result[i].Type = &types.Script{
				CodeHash: output.Type.CodeHash,
				HashType: output.Type.HashType,
				Args:     output.Type.Args,
			}
		}
	}
	return result
}

func toInputs(inputs []cellInput) []*types.CellInput {
	result := make([]*types.CellInput, len(inputs))
	for i := 0; i < len(inputs); i++ {
		input := inputs[i]
		result[i] = &types.CellInput{
			Since: uint64(input.Since),
			PreviousOutput: &types.OutPoint{
				TxHash: input.PreviousOutput.TxHash,
				Index:  uint(input.PreviousOutput.Index),
			},
		}
	}
	return result
}

func toCellDeps(deps []cellDep) []*types.CellDep {
	result := make([]*types.CellDep, len(deps))
	for i := 0; i < len(deps); i++ {
		dep := deps[i]
		result[i] = &types.CellDep{
			OutPoint: &types.OutPoint{
				TxHash: dep.OutPoint.TxHash,
				Index:  uint(dep.OutPoint.Index),
			},
			DepType: dep.DepType,
		}
	}
	return result
}

func FromTransaction(tx *types.Transaction) (string, error) {
	result := inTransaction{
		Version:     hexutil.Uint(tx.Version),
		HeaderDeps:  tx.HeaderDeps,
		CellDeps:    fromCellDeps(tx.CellDeps),
		Inputs:      fromInputs(tx.Inputs),
		Outputs:     fromOutputs(tx.Outputs),
		OutputsData: fromBytesArray(tx.OutputsData),
		Witnesses:   fromBytesArray(tx.Witnesses),
	}
	data, err := json.Marshal(&result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fromCellDeps(deps []*types.CellDep) []cellDep {
	result := make([]cellDep, len(deps))
	for i := 0; i < len(deps); i++ {
		dep := deps[i]
		result[i] = cellDep{
			OutPoint: outPoint{
				TxHash: dep.OutPoint.TxHash,
				Index:  hexutil.Uint(dep.OutPoint.Index),
			},
			DepType: dep.DepType,
		}
	}
	return result
}

func fromInputs(inputs []*types.CellInput) []cellInput {
	result := make([]cellInput, len(inputs))
	for i := 0; i < len(inputs); i++ {
		input := inputs[i]
		result[i] = cellInput{
			Since: hexutil.Uint64(input.Since),
			PreviousOutput: outPoint{
				TxHash: input.PreviousOutput.TxHash,
				Index:  hexutil.Uint(input.PreviousOutput.Index),
			},
		}
	}
	return result
}

func fromOutputs(outputs []*types.CellOutput) []cellOutput {
	result := make([]cellOutput, len(outputs))
	for i := 0; i < len(outputs); i++ {
		output := outputs[i]
		result[i] = cellOutput{
			Capacity: hexutil.Uint64(output.Capacity),
			Lock: &script{
				CodeHash: output.Lock.CodeHash,
				HashType: output.Lock.HashType,
				Args:     output.Lock.Args,
			},
		}
		if output.Type != nil {
			result[i].Type = &script{
				CodeHash: output.Type.CodeHash,
				HashType: output.Type.HashType,
				Args:     output.Type.Args,
			}
		}
	}
	return result
}

func fromBytesArray(bytes [][]byte) []hexutil.Bytes {
	result := make([]hexutil.Bytes, len(bytes))
	for i, data := range bytes {
		result[i] = data
	}
	return result
}
