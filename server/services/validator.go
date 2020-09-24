package services

import (
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"strconv"
)

func validateCapacity(inputTotalAmount uint64, outputTotalAmount uint64) *types.Error {
	if inputTotalAmount <= outputTotalAmount {
		return CapacityNotEnoughError
	}
	return nil
}

func validateOutputOperations(operations []*types.Operation) (uint64, *types.Error) {
	var outputTotalAmount uint64
	outputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.OutputOpType
	})
	if len(outputOperations) == 0 {
		return 0, MissingOutputOperationsError
	}

	operationSize := len(outputOperations)
	for i, operation := range outputOperations {
		if operation.Amount.Value[0:1] == "-" {
			return 0, InvalidOutputOperationAmountValueError
		}
		amount, err := strconv.ParseUint(operation.Amount.Value, 10, 64)
		if err != nil {
			return 0, InvalidOutputOperationAmountValueError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return 0, AddressParseError
		}
		if isBlake160SighashAllLock(addr) {
			if i == operationSize-1 {
				continue
			}
			if amount < MinCapacity {
				return 0, LessThanMinCapacityError
			}
		}

		outputTotalAmount += amount
	}
	return outputTotalAmount, nil
}

func validateInputOperations(operations []*types.Operation) (uint64, *types.Error) {
	var inputTotalAmount uint64
	inputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.InputOpType
	})

	if len(inputOperations) == 0 {
		return 0, MissingInputOperationsError
	}

	for _, operation := range inputOperations {
		amount, err := strconv.ParseUint(operation.Amount.Value[1:], 10, 64)
		if err != nil || operation.Amount.Value[0:1] != "-" {
			return 0, InvalidInputOperationAmountValueError
		}
		err = asserter.CoinChange(operation.CoinChange)
		if err != nil {
			return 0, InvalidCoinChangeError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return 0, AddressParseError
		}
		// do not support send to multisig all lock
		if isBlake160MultisigAllLock(addr) {
			return 0, NotSupportMultisigAllLockError
		}

		inputTotalAmount += amount
	}
	return inputTotalAmount, nil
}

func validateTxType(metadata map[string]interface{}) (string, *types.Error) {
	var constructionMetadata ckb.ConstructionMetadata
	err := types.UnmarshalMap(metadata, &constructionMetadata)
	if err != nil {
		return "", InvalidConstructionMetadataError
	}

	return constructionMetadata.TxType, nil
}
