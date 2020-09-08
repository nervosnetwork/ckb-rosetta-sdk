package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTransaction "github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
	"strings"
)

func OperationFilter(arr []*types.Operation, cond func(*types.Operation) bool) []*types.Operation {
	var result []*types.Operation
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}

func isBlake160SighashAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == ckbTransaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH
}

func isBlake160MultisigAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == ckbTransaction.SECP256K1_BLAKE160_MULTISIG_ALL_TYPE_HASH
}

func generateCoinIdentifiersOption(request *types.ConstructionPreprocessRequest) ([]string, *types.Error) {
	inputOperations := OperationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Input"
	})
	var identifiers []string
	for _, operation := range inputOperations {
		identifiers = append(identifiers, operation.CoinChange.CoinIdentifier.Identifier)
	}
	return identifiers, nil
}

func fetchLiveCells(ctx context.Context, request *types.ConstructionMetadataRequest, s *ConstructionAPIService) ([]ckbTypes.CellInfo, *types.Error) {
	var items []ckbTypes.BatchLiveCellItem
	for _, option := range request.Options["out_points"].([]interface{}) {
		identifier := strings.Split(option.(string), ":")
		index, err := strconv.ParseUint(identifier[1], 10, 32)
		if err != nil {
			return nil, CoinIdentifierInvalidError
		}
		outPoint := ckbTypes.OutPoint{
			TxHash: ckbTypes.HexToHash(identifier[0]),
			Index:  uint(index),
		}
		items = append(items, ckbTypes.BatchLiveCellItem{OutPoint: outPoint, WithData: false})
	}

	err := s.client.BatchLiveCells(ctx, items)
	if err != nil {
		return nil, CoinIdentifierInvalidError
	}
	var cellInfos []ckbTypes.CellInfo
	for _, item := range items {
		cell := item.Result.Cell
		if item.Result.Status != "live" {
			return nil, LiveCellMetadataHasDeadCellsError
		}
		cellInfos = append(cellInfos, ckbTypes.CellInfo{Output: cell.Output, Data: cell.Data})
	}
	return cellInfos, nil
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

func parseInputCellsFromMetadata(inputs string) ([]ckbTypes.CellInfo, *types.Error) {
	var inputCells []ckbTypes.CellInfo
	decodedInputs, err := base64.StdEncoding.DecodeString(inputs)
	if err != nil {
		return nil, InvalidLiveCellsError
	}
	err = json.Unmarshal(decodedInputs, &inputCells)
	if err != nil {
		return nil, InvalidLiveCellsError
	}
	return inputCells, nil
}
