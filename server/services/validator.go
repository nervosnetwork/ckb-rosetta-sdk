package services

import (
	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-sdk-go/address"
	"strconv"
)

func validateCapacity(input int64, output int64) *types.Error {
	if input <= output {
		return CapacityNotEnoughError
	}
	return nil
}

func validateVoutOperations(request *types.ConstructionPreprocessRequest) (int64, *types.Error) {
	var output int64
	voutOperations := operationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Vout"
	})
	if len(voutOperations) == 0 {
		return 0, MissingVoutOperationsError
	}

	for _, operation := range voutOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount <= 0 {
			return 0, InvalidVoutOperationAmountValueError
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

		output += amount
	}
	return output, nil
}

func validateVinOperations(request *types.ConstructionPreprocessRequest) (int64, *types.Error) {
	var input int64
	vinOperations := operationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Vin"
	})

	if len(vinOperations) == 0 {
		return 0, MissingVinOperationsError
	}

	for _, operation := range vinOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount >= 0 {
			return 0, InvalidVinOperationAmountValueError
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

		input += -amount
	}
	return input, nil
}
