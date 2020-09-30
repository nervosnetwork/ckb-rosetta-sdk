package services

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type outPoint struct {
	TxHash ckbTypes.Hash `json:"tx_hash"`
	Index  hexutil.Uint  `json:"index"`
}

type cellDep struct {
	OutPoint outPoint         `json:"out_point"`
	DepType  ckbTypes.DepType `json:"dep_type"`
}

type cellInput struct {
	Since          hexutil.Uint64 `json:"since"`
	PreviousOutput outPoint       `json:"previous_output"`
}

type script struct {
	CodeHash ckbTypes.Hash           `json:"code_hash"`
	HashType ckbTypes.ScriptHashType `json:"hash_type"`
	Args     hexutil.Bytes           `json:"args"`
}

type cellOutput struct {
	Capacity hexutil.Uint64 `json:"capacity"`
	Lock     *script        `json:"lock"`
	Type     *script        `json:"type"`
}

type transaction struct {
	Version     hexutil.Uint    `json:"version"`
	Hash        ckbTypes.Hash   `json:"hash"`
	CellDeps    []cellDep       `json:"cell_deps"`
	HeaderDeps  []ckbTypes.Hash `json:"header_deps"`
	Inputs      []cellInput     `json:"inputs"`
	Outputs     []cellOutput    `json:"outputs"`
	OutputsData []hexutil.Bytes `json:"outputs_data"`
	Witnesses   []hexutil.Bytes `json:"witnesses"`
}

type inTransaction struct {
	Version     hexutil.Uint    `json:"version"`
	CellDeps    []cellDep       `json:"cell_deps"`
	HeaderDeps  []ckbTypes.Hash `json:"header_deps"`
	Inputs      []cellInput     `json:"inputs"`
	Outputs     []cellOutput    `json:"outputs"`
	OutputsData []hexutil.Bytes `json:"outputs_data"`
	Witnesses   []hexutil.Bytes `json:"witnesses"`
}

type inRosettaTransaction struct {
	Version                  hexutil.Uint               `json:"version"`
	Hash                     ckbTypes.Hash              `json:"hash"`
	CellDeps                 []cellDep                  `json:"cell_deps"`
	HeaderDeps               []ckbTypes.Hash            `json:"header_deps"`
	Inputs                   []cellInput                `json:"inputs"`
	Outputs                  []cellOutput               `json:"outputs"`
	OutputsData              []hexutil.Bytes            `json:"outputs_data"`
	Witnesses                []hexutil.Bytes            `json:"witnesses"`
	InputAmounts             []*types.Amount            `json:"input_amounts"`
	InputAccounts            []*types.AccountIdentifier `json:"input_accounts"`
	OutputAmounts            []*types.Amount            `json:"output_amounts"`
	OutputAccounts           []*types.AccountIdentifier `json:"output_accounts"`
	AccountIdentifierSigners []*types.AccountIdentifier `json:"account_identifier_signers,omitempty"`
}

type rosettaTransaction struct {
	Version                  uint                       `json:"version"`
	Hash                     ckbTypes.Hash              `json:"hash"`
	CellDeps                 []*ckbTypes.CellDep        `json:"cell_deps"`
	HeaderDeps               []ckbTypes.Hash            `json:"header_deps"`
	Inputs                   []*ckbTypes.CellInput      `json:"inputs"`
	Outputs                  []*ckbTypes.CellOutput     `json:"outputs"`
	OutputsData              [][]byte                   `json:"outputs_data"`
	Witnesses                [][]byte                   `json:"witnesses"`
	InputAmounts             []*types.Amount            `json:"input_amounts"`
	InputAccounts            []*types.AccountIdentifier `json:"input_accounts"`
	OutputAmounts            []*types.Amount            `json:"output_amounts"`
	OutputAccounts           []*types.AccountIdentifier `json:"output_accounts"`
	AccountIdentifierSigners []*types.AccountIdentifier `json:"account_identifier_signers,omitempty"`
}

func ToTransaction(data string) (*ckbTypes.Transaction, error) {
	var tx transaction
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		return nil, err
	}
	return &ckbTypes.Transaction{
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

func toOutputs(outputs []cellOutput) []*ckbTypes.CellOutput {
	result := make([]*ckbTypes.CellOutput, len(outputs))
	for i := 0; i < len(outputs); i++ {
		output := outputs[i]
		result[i] = &ckbTypes.CellOutput{
			Capacity: uint64(output.Capacity),
			Lock: &ckbTypes.Script{
				CodeHash: output.Lock.CodeHash,
				HashType: output.Lock.HashType,
				Args:     output.Lock.Args,
			},
		}
		if output.Type != nil {
			result[i].Type = &ckbTypes.Script{
				CodeHash: output.Type.CodeHash,
				HashType: output.Type.HashType,
				Args:     output.Type.Args,
			}
		}
	}
	return result
}

func toInputs(inputs []cellInput) []*ckbTypes.CellInput {
	result := make([]*ckbTypes.CellInput, len(inputs))
	for i := 0; i < len(inputs); i++ {
		input := inputs[i]
		result[i] = &ckbTypes.CellInput{
			Since: uint64(input.Since),
			PreviousOutput: &ckbTypes.OutPoint{
				TxHash: input.PreviousOutput.TxHash,
				Index:  uint(input.PreviousOutput.Index),
			},
		}
	}
	return result
}

func toCellDeps(deps []cellDep) []*ckbTypes.CellDep {
	result := make([]*ckbTypes.CellDep, len(deps))
	for i := 0; i < len(deps); i++ {
		dep := deps[i]
		result[i] = &ckbTypes.CellDep{
			OutPoint: &ckbTypes.OutPoint{
				TxHash: dep.OutPoint.TxHash,
				Index:  uint(dep.OutPoint.Index),
			},
			DepType: dep.DepType,
		}
	}
	return result
}

func FromTransaction(tx *ckbTypes.Transaction) (string, error) {
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

func fromCellDeps(deps []*ckbTypes.CellDep) []cellDep {
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

func fromInputs(inputs []*ckbTypes.CellInput) []cellInput {
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

func fromOutputs(outputs []*ckbTypes.CellOutput) []cellOutput {
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

func getCoinIdentifier(outPoint *ckbTypes.OutPoint) *types.CoinIdentifier {
	return &types.CoinIdentifier{
		Identifier: fmt.Sprintf("%s:%d", outPoint.TxHash.String(), outPoint.Index),
	}
}

func toScript(s ckb.Script) (*ckbTypes.Script, error) {
	args, err := hex.DecodeString(s.Args[2:])
	if err != nil {
		return nil, err
	}

	return &ckbTypes.Script{
		CodeHash: ckbTypes.HexToHash(s.CodeHash),
		HashType: ckbTypes.ScriptHashType(s.HashType),
		Args:     args,
	}, nil
}
