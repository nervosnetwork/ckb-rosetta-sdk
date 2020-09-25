package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
)

var _ UnsignedTxBuilderInterface = UnsignedTxBuilderSecp256k1{}

const txVersion uint = 0

type UnsignedTxBuilderSecp256k1 struct {
	UnsignedTxBuilder
	Cfg              *config.Config
	InputOperations  []*types.Operation
	OutputOperations []*types.Operation
}

func NewUnsignedTxBuilderSecp256k1(cfg *config.Config, inputOperations []*types.Operation, outputOperations []*types.Operation) *UnsignedTxBuilderSecp256k1 {
	b := UnsignedTxBuilderSecp256k1{
		Cfg:              cfg,
		InputOperations:  inputOperations,
		OutputOperations: outputOperations,
	}
	b.UnsignedTxBuilder.BuildVersion = b.BuildVersion
	b.UnsignedTxBuilder.BuildCellDeps = b.BuildCellDeps
	b.UnsignedTxBuilder.BuildHeaderDeps = b.BuildHeaderDeps
	b.UnsignedTxBuilder.BuildInputs = b.BuildInputs
	b.UnsignedTxBuilder.BuildOutputs = b.BuildOutputs
	b.UnsignedTxBuilder.BuildOutputsData = b.BuildOutputsData
	b.UnsignedTxBuilder.BuildWitnesses = b.BuildWitnesses
	return &b
}

func (b UnsignedTxBuilderSecp256k1) BuildVersion() (uint, error) {
	return txVersion, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildCellDeps() ([]*ckbTypes.CellDep, error) {
	var cellDeps []*ckbTypes.CellDep
	cellDeps = append(cellDeps, &ckbTypes.CellDep{
		OutPoint: &ckbTypes.OutPoint{
			TxHash: ckbTypes.HexToHash(b.Cfg.Secp256k1Blake160.Deps[0].TxHash),
			Index:  b.Cfg.Secp256k1Blake160.Deps[0].Index,
		},
		DepType: ckbTypes.DepType(b.Cfg.Secp256k1Blake160.Deps[0].DepType),
	})

	return cellDeps, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildHeaderDeps() ([]ckbTypes.Hash, error) {
	return []ckbTypes.Hash{}, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildInputs() ([]*ckbTypes.CellInput, map[string]interface{}, error) {
	var cellInputs []*ckbTypes.CellInput
	for _, operation := range b.InputOperations {
		outPoint, err := GenerateOutPointFromCoinIdentifier(operation.CoinChange.CoinIdentifier.Identifier)
		if err != nil {
			return nil, nil, err
		}
		cellInputs = append(cellInputs, &ckbTypes.CellInput{
			Since:          0,
			PreviousOutput: outPoint,
		})
	}
	return cellInputs, nil, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputs(options map[string]interface{}) ([]*ckbTypes.CellOutput, map[string]interface{}, error) {
	var cellOutputs []*ckbTypes.CellOutput
	for _, operation := range b.OutputOperations {
		parsedAddress, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, nil, err
		}
		capacity, err := strconv.ParseUint(operation.Amount.Value, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		cellOutputs = append(cellOutputs, &ckbTypes.CellOutput{
			Capacity: capacity,
			Lock:     parsedAddress.Script,
		})
	}
	return cellOutputs, nil, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildOutputsData(options map[string]interface{}) ([][]byte, error) {
	var outputsData [][]byte
	outputsSize := len(b.OutputOperations)
	for i := 0; i < outputsSize-1; i++ {
		outputsData = append(outputsData, []byte{})
	}
	return outputsData, nil
}

func (b UnsignedTxBuilderSecp256k1) BuildWitnesses() ([][]byte, error) {
	cellInputsSize := len(b.InputOperations)
	witnesses := make([][]byte, cellInputsSize)
	lockScriptHashes := make(map[ckbTypes.Hash][]int)
	for i, operation := range b.InputOperations {
		parsedAddress, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, err
		}
		lockHash, err := parsedAddress.Script.Hash()
		if err != nil {
			return nil, err
		}
		if _, ok := lockScriptHashes[lockHash]; !ok {
			lockScriptHashes[lockHash] = append(lockScriptHashes[lockHash], i)
		}
	}

	indexGroups := make([][]int, 0, len(lockScriptHashes))
	for _, index := range lockScriptHashes {
		indexGroups = append(indexGroups, index)
	}
	emptyWitness, _ := transaction.EmptyWitnessArg.Serialize()
	for _, indexes := range indexGroups {
		firstIndexOfGroup := indexes[0]
		witnesses[firstIndexOfGroup] = emptyWitness
	}

	return witnesses, nil
}
