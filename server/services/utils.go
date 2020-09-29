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
	inputOperations = getInputOperations(operations)
	outputOperations = getOutputOperations(operations)

	return
}

func getInputOperations(operations []*types.Operation) []*types.Operation {
	return operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.InputOpType
	})
}

func getOutputOperations(operations []*types.Operation) []*types.Operation {
	return operationFilter(operations, func(operation *types.Operation) bool {
		return operation.Type == ckb.OutputOpType
	})
}

func getConstructionType(operations []*types.Operation, signatures []*types.Signature, cfg *config.Config) (string, *types.Error) {
	inputOperations := getInputOperations(operations)
	outputOperations := getOutputOperations(operations)
	if ok, err := isTransferCKB(inputOperations, outputOperations, signatures, cfg); ok {
		if err != nil {
			return "", err
		}
		return ckb.TransferCKB, nil
	} else {
		return "", UnsupportedConstructionTypeError
	}
}

func isTransferCKB(inputOperations []*types.Operation, outputOperations []*types.Operation, signatures []*types.Signature, cfg *config.Config) (bool, *types.Error) {
	if signatures == nil {
		for _, operation := range inputOperations {
			parsedAddress, err := address.Parse(operation.Account.Address)
			if err != nil {
				return false, AddressParseError
			}
			if parsedAddress.Script.CodeHash.String() != cfg.Secp256k1Blake160.Script.CodeHash || string(parsedAddress.Script.HashType) != cfg.Secp256k1Blake160.Script.HashType {
				return false, nil
			}
		}

		for _, operation := range outputOperations {
			var metadata ckb.OperationMetadata
			err := types.UnmarshalMap(operation.Metadata, &metadata)
			if err != nil {
				return false, wrapErr(DataParseError, err)
			}
			if metadata.Type != nil {
				return false, nil
			}
		}
	} else {
		var metadata ckb.AccountIdentifierMetadata
		for _, signature := range signatures {
			err := types.UnmarshalMap(signature.SigningPayload.AccountIdentifier.Metadata, &metadata)
			if err != nil {
				return false, wrapErr(InvalidAccountIdentifierMetadataError, err)
			}
			if metadata.LockType != ckb.Secp256k1Blake160Lock.String() {
				return false, nil
			}
		}
	}

	return true, nil
}
