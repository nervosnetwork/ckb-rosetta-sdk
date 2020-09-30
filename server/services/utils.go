package services

import (
	"encoding/json"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbRpc "github.com/nervosnetwork/ckb-sdk-go/rpc"
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

func toRosettaTransaction(rTx inRosettaTransaction) *rosettaTransaction {
	return &rosettaTransaction{
		Version:                  uint(rTx.Version),
		Hash:                     ckbTypes.Hash{},
		CellDeps:                 toCellDeps(rTx.CellDeps),
		HeaderDeps:               rTx.HeaderDeps,
		Inputs:                   toInputs(rTx.Inputs),
		Outputs:                  toOutputs(rTx.Outputs),
		OutputsData:              toBytesArray(rTx.OutputsData),
		Witnesses:                toBytesArray(rTx.Witnesses),
		InputAmounts:             rTx.InputAmounts,
		InputAccounts:            rTx.InputAccounts,
		OutputAmounts:            rTx.OutputAmounts,
		OutputAccounts:           rTx.OutputAccounts,
		AccountIdentifierSigners: rTx.AccountIdentifierSigners,
	}
}

func rosettaTransactionFromString(tx string) (*rosettaTransaction, error) {
	var rTx inRosettaTransaction
	err := json.Unmarshal([]byte(tx), &rTx)
	if err != nil {
		return nil, err
	}

	return toRosettaTransaction(rTx), nil
}

func rTxStringForPayload(txStr string, operations []*types.Operation) (string, *types.Error) {
	var inputAmounts []*types.Amount
	var inputAccounts []*types.AccountIdentifier
	inputOperations := getInputOperations(operations)
	for _, inputOperation := range inputOperations {
		inputAmounts = append(inputAmounts, inputOperation.Amount)
		inputAccounts = append(inputAccounts, inputOperation.Account)
	}
	var outputAmounts []*types.Amount
	var outputAccounts []*types.AccountIdentifier
	outputOperations := getOutputOperations(operations)
	for _, outputOperation := range outputOperations {
		outputAmounts = append(outputAmounts, outputOperation.Amount)
		outputAccounts = append(outputAccounts, outputOperation.Account)
	}
	tx, err := ckbRpc.TransactionFromString(txStr)
	if err != nil {
		return "", TransactionParseError
	}
	rTx := inRosettaTransaction{
		Version:                  hexutil.Uint(tx.Version),
		Hash:                     tx.Hash,
		CellDeps:                 fromCellDeps(tx.CellDeps),
		HeaderDeps:               tx.HeaderDeps,
		Inputs:                   fromInputs(tx.Inputs),
		Outputs:                  fromOutputs(tx.Outputs),
		OutputsData:              fromBytesArray(tx.OutputsData),
		Witnesses:                fromBytesArray(tx.Witnesses),
		InputAmounts:             inputAmounts,
		InputAccounts:            inputAccounts,
		OutputAmounts:            outputAmounts,
		OutputAccounts:           outputAccounts,
		AccountIdentifierSigners: []*types.AccountIdentifier{},
	}
	bytes, err := json.Marshal(rTx)
	if err != nil {
		return "", wrapErr(TransactionParseError, err)
	}

	return string(bytes), nil
}

func rTxStringForCombine(txStr string, signatures []*types.Signature) (string, *types.Error) {
	rTx, err := rosettaTransactionFromString(txStr)
	if err != nil {
		return "", TransactionParseError
	}
	var accountIdentifierSigners []*types.AccountIdentifier
	for _, signature := range signatures {
		accountIdentifierSigners = append(accountIdentifierSigners, signature.SigningPayload.AccountIdentifier)
	}
	inRtx := inRosettaTransaction{
		Version:                  hexutil.Uint(rTx.Version),
		Hash:                     rTx.Hash,
		CellDeps:                 fromCellDeps(rTx.CellDeps),
		HeaderDeps:               rTx.HeaderDeps,
		Inputs:                   fromInputs(rTx.Inputs),
		Outputs:                  fromOutputs(rTx.Outputs),
		OutputsData:              fromBytesArray(rTx.OutputsData),
		Witnesses:                fromBytesArray(rTx.Witnesses),
		InputAmounts:             rTx.InputAmounts,
		InputAccounts:            rTx.InputAccounts,
		OutputAmounts:            rTx.OutputAmounts,
		OutputAccounts:           rTx.OutputAccounts,
		AccountIdentifierSigners: accountIdentifierSigners,
	}

	bytes, err := json.Marshal(inRtx)
	if err != nil {
		return "", wrapErr(TransactionParseError, err)
	}

	return string(bytes), nil
}

func rTxString(rTx *rosettaTransaction) (string, error) {
	iRtx := inRosettaTransaction{
		Version:                  hexutil.Uint(rTx.Version),
		Hash:                     rTx.Hash,
		CellDeps:                 fromCellDeps(rTx.CellDeps),
		HeaderDeps:               rTx.HeaderDeps,
		Inputs:                   fromInputs(rTx.Inputs),
		Outputs:                  fromOutputs(rTx.Outputs),
		OutputsData:              fromBytesArray(rTx.OutputsData),
		Witnesses:                fromBytesArray(rTx.Witnesses),
		InputAmounts:             rTx.InputAmounts,
		InputAccounts:            rTx.InputAccounts,
		OutputAmounts:            rTx.OutputAmounts,
		OutputAccounts:           rTx.OutputAccounts,
		AccountIdentifierSigners: []*types.AccountIdentifier{},
	}

	bytes, err := json.Marshal(iRtx)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
