package factory

import (
	"encoding/hex"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type TxSizeEstimatorFactory struct{}

func (tf TxSizeEstimatorFactory) CreateTxSizeEstimator(txType string) TxSizeEstimater {
	switch txType {
	case ckb.Secp256k1Tx:
		return NewSecp256k1TxSizeEstimator()
	default:
		return nil
	}
}

type TxSizeEstimater interface {
	EstimatedTxSize(operations []*types.Operation) (uint64, error)
	HeaderDepsSize() uint64
	CellDepsSize() uint64
	WitnessSize(int) uint64
	OutputSize(*types.Operation) (uint64, error)
	OutputDataSize(string) (uint64, error)
}

type TxSizeEstimator struct {
	HeaderDepsSize func() uint64
	CellDepsSize   func() uint64
	WitnessSize    func(int) uint64
	OutputSize     func(*types.Operation) (uint64, error)
	OutputDataSize func(string) (uint64, error)
}

func (tse *TxSizeEstimator) EstimatedTxSize(operations []*types.Operation) (uint64, error) {
	var sizeArr []uint64
	var txSize uint64
	sizeArr = append(sizeArr, ckb.BaseTxSize, tse.HeaderDepsSize(), tse.CellDepsSize())
	for i, operation := range operations {
		switch operation.Type {
		case ckb.InputOpType:
			sizeArr = append(sizeArr, ckb.InputSize, tse.WitnessSize(i))
		case ckb.OutputOpType:
			var metadata ckb.OperationMetadata
			if err := types.UnmarshalMap(operation.Metadata, &metadata); err != nil {
				return 0, err
			}
			outputSize, err := tse.OutputSize(operation)
			if err != nil {
				return 0, err
			}
			outputDataSize, err := tse.OutputDataSize(metadata.Data)
			if err != nil {
				return 0, err
			}
			sizeArr = append(sizeArr, outputSize, outputDataSize)
		}
	}
	for _, size := range sizeArr {
		txSize += size
	}
	txSize += ckb.SerializedOffsetByteSize

	return txSize, nil
}

type Secp256k1TxSizeEstimator struct {
	TxSizeEstimator
}

func NewSecp256k1TxSizeEstimator() *Secp256k1TxSizeEstimator {
	tes := Secp256k1TxSizeEstimator{}
	tes.TxSizeEstimator.HeaderDepsSize = tes.HeaderDepsSize
	tes.TxSizeEstimator.CellDepsSize = tes.CellDepsSize
	tes.TxSizeEstimator.WitnessSize = tes.WitnessSize
	tes.TxSizeEstimator.OutputSize = tes.OutputSize
	tes.TxSizeEstimator.OutputDataSize = tes.OutputDataSize
	return &tes
}

func (tse Secp256k1TxSizeEstimator) HeaderDepsSize() uint64 {
	return 0
}

func (tse Secp256k1TxSizeEstimator) CellDepsSize() uint64 {
	return ckb.CellDepSize
}

func (tse Secp256k1TxSizeEstimator) WitnessSize(i int) uint64 {
	var witnessBytes []byte
	if i == 0 {
		witnessBytes, _ = transaction.EmptyWitnessArg.Serialize()
		witnessBytes = ckbTypes.SerializeBytes(witnessBytes)
	}
	size := uint64(len(witnessBytes)) + ckb.SerializedOffsetByteSize

	return size
}

func (tse Secp256k1TxSizeEstimator) OutputSize(operation *types.Operation) (uint64, error) {
	parsedAddress, err := address.Parse(operation.Account.Address)
	if err != nil {
		return 0, err
	}
	cellOutput := ckbTypes.CellOutput{
		Capacity: 0,
		Lock:     parsedAddress.Script,
		Type:     nil,
	}
	serializedCellOutput, _ := cellOutput.Serialize()
	var outputSize uint64
	outputSize = uint64(len(serializedCellOutput)) + ckb.SerializedOffsetByteSize

	return outputSize, nil
}

func (tse Secp256k1TxSizeEstimator) OutputDataSize(data string) (uint64, error) {
	var dataBytes []byte
	var err error
	if data != "" {
		dataBytes, err = hex.DecodeString(data)
		if err != nil {
			return 0, err
		}
	}
	serializedDataBytes := ckbTypes.SerializeBytes(dataBytes)
	outputDataSize := uint64(len(serializedDataBytes)) + ckb.SerializedOffsetByteSize

	return outputDataSize, nil
}
