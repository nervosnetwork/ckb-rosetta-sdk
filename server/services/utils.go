package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

func operationFilter(arr []*types.Operation, cond func(*types.Operation) bool) []*types.Operation {
	var result []*types.Operation
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func isBlake160SighashAllLock(addr *address.ParsedAddress, cfg *config.Config) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == cfg.Secp256k1Blake160.Script.CodeHash
}

func isBlake160MultisigAllLock(addr *address.ParsedAddress, cfg *config.Config) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == cfg.Secp256k1Blake160Mutisig.Script.CodeHash
}

func wrapErr(rErr *types.Error, err error) *types.Error {
	newErr := &types.Error{
		Code:    rErr.Code,
		Message: rErr.Message,
	}
	if err != nil {
		newErr.Details = map[string]interface{}{
			"context": err.Error(),
		}
	}

	return newErr
}

func separateInputAndOutput(operations []*types.Operation) (inputOperations []*types.Operation, outputOperations []*types.Operation) {
	inputOperations = operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.InputOpType
	})
	outputOperations = operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.OutputOpType
	})

	return
}
