package builder

import (
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

// UnsignedTx is an interface for different script unsignedTxBuilder
type UnsignedTxBuilder interface {
	BuildVersion() (uint, error)
	BuildCellDeps() ([]*ckbTypes.CellDep, error)
	BuildHeaderDeps() ([]ckbTypes.Hash, error)
	BuildInputs() ([]*ckbTypes.CellInput, map[string]interface{}, error)
	BuildOutputs(options map[string]interface{}) ([]*ckbTypes.CellOutput, map[string]interface{}, error)
	BuildOutputsData(options map[string]interface{}) ([][]byte, error)
	BuildWitnesses() ([][]byte, error)
	Build() (*ckbTypes.Transaction, error)
}

type UnsignedTx struct {
	BuildVersion     func() (uint, error)
	BuildCellDeps    func() ([]*ckbTypes.CellDep, error)
	BuildHeaderDeps  func() ([]ckbTypes.Hash, error)
	BuildInputs      func() ([]*ckbTypes.CellInput, map[string]interface{}, error)
	BuildOutputs     func(options map[string]interface{}) ([]*ckbTypes.CellOutput, map[string]interface{}, error)
	BuildOutputsData func(options map[string]interface{}) ([][]byte, error)
	BuildWitnesses   func() ([][]byte, error)
}

func (utb UnsignedTx) Build() (*ckbTypes.Transaction, error) {
	version, err := utb.BuildVersion()
	cellDeps, err := utb.BuildCellDeps()
	if err != nil {
		return nil, err
	}
	headerDeps, err := utb.BuildHeaderDeps()
	if err != nil {
		return nil, err
	}
	inputs, options, err := utb.BuildInputs()
	if err != nil {
		return nil, err
	}
	outputs, outputOptions, err := utb.BuildOutputs(options)
	if err != nil {
		return nil, err
	}
	outputsData, err := utb.BuildOutputsData(outputOptions)
	if err != nil {
		return nil, err
	}
	witnesses, err := utb.BuildWitnesses()
	if err != nil {
		return nil, err
	}
	tx := &ckbTypes.Transaction{
		Version:     version,
		CellDeps:    cellDeps,
		HeaderDeps:  headerDeps,
		Inputs:      inputs,
		Outputs:     outputs,
		OutputsData: outputsData,
		Witnesses:   witnesses,
	}

	return tx, nil
}
