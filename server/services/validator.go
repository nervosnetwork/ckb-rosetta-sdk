package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
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
	outputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
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
	inputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
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

func ValidateCellDeps(operations []*types.Operation) (map[string]ckbTypes.CellDep, *types.Error) {
	cellDeps := make(map[string]ckbTypes.CellDep)
	err := validateCellDepsOnInputOperation(operations, cellDeps)
	if err != nil {
		return nil, err
	}

	err = validateCellDepsOnOutputOperation(operations, cellDeps)
	if err != nil {
		return nil, err
	}
	return cellDeps, nil
}

func validateCellDepsOnOutputOperation(operations []*types.Operation, cellDeps map[string]ckbTypes.CellDep) *types.Error {
	outputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == "Output"
	})
	for _, operation := range outputOperations {
		if mCellDep, ok := operation.Metadata["cell_dep"]; ok {
			decodedCellDep, err := base64.StdEncoding.DecodeString(mCellDep.(string))
			if err != nil {
				return InvalidCellDepError
			}
			var cellDep ckbTypes.CellDep
			err = json.Unmarshal(decodedCellDep, &cellDep)
			if err != nil {
				return InvalidCellDepError
			}
			key := fmt.Sprintf("%s:%d", cellDep.OutPoint.TxHash, cellDep.OutPoint.Index)
			if _, ok := cellDeps[key]; !ok {
				cellDeps[key] = cellDep
			}
		}
	}
	return nil
}

func validateCellDepsOnInputOperation(operations []*types.Operation, cellDeps map[string]ckbTypes.CellDep) *types.Error {
	inputOperations := OperationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == "Input"
	})
	for _, operation := range inputOperations {
		mCellDep, ok := operation.Metadata["cell_dep"]
		if !ok || mCellDep == nil {
			return MissingCellDepsOnOperationError
		}
		decodedCellDep, err := base64.StdEncoding.DecodeString(mCellDep.(string))
		if err != nil {
			return InvalidCellDepError
		}
		var cellDep ckbTypes.CellDep
		err = json.Unmarshal(decodedCellDep, &cellDep)
		if err != nil {
			return InvalidCellDepError
		}
		key := fmt.Sprintf("%s:%d", cellDep.OutPoint.TxHash, cellDep.OutPoint.Index)
		if _, ok := cellDeps[key]; !ok {
			cellDeps[key] = cellDep
		}
	}
	return nil
}

func validateInputsMetadata(metadata map[string]interface{}) (string, *types.Error) {
	inputs, ok := metadata["inputs"]
	if !ok || inputs == nil {
		return "", MissingInputsOnConstructionPayloadsRequestError
	}
	return inputs.(string), nil
}
