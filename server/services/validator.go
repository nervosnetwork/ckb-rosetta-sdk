package services

import (
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"strconv"
)

func validateCapacity(inputTotalAmount int64, outputTotalAmount int64) *types.Error {
	if inputTotalAmount <= outputTotalAmount {
		return CapacityNotEnoughError
	}
	return nil
}

func validateOutputOperations(operations []*types.Operation) (int64, *types.Error) {
	var outputTotalAmount int64
	outputOperations := operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == "Output"
	})
	if len(outputOperations) == 0 {
		return 0, MissingOutputOperationsError
	}

	for _, operation := range outputOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount <= 0 {
			return 0, InvalidOutputOperationAmountValueError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return 0, AddressParseError
		}
		if isBlake160SighashAllLock(addr) {
			if amount < MinCapacity {
				return 0, LessThanMinCapacityError
			}
		}

		outputTotalAmount += amount
	}
	return outputTotalAmount, nil
}

func validateInputOperations(operations []*types.Operation) (int64, *types.Error) {
	var inputTotalAmount int64
	inputOperations := operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == "Input"
	})

	if len(inputOperations) == 0 {
		return 0, MissingInputOperationsError
	}

	for _, operation := range inputOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount >= 0 {
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

		inputTotalAmount += -amount
	}
	return inputTotalAmount, nil
}
