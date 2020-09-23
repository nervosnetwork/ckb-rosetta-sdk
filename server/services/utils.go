package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
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

func generateCoinIdentifiersOption(request *types.ConstructionPreprocessRequest) ([]byte, *types.Error) {
	inputOperations := OperationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Input"
	})
	var identifiers []string
	for _, operation := range inputOperations {
		identifiers = append(identifiers, operation.CoinChange.CoinIdentifier.Identifier)
	}
	jsonIdentifiers, err := json.Marshal(identifiers)
	if err != nil {
		return nil, CoinIdentifierInvalidError
	}
	return jsonIdentifiers, nil
}

func fetchLiveCells(ctx context.Context, options map[string]interface{}, s *ConstructionAPIService) ([]ckbTypes.CellInfo, *types.Error) {
	var items []ckbTypes.BatchLiveCellItem
	strCoinIdentifier, ok := options["coin_identifiers"].(string)
	if !ok {
		return nil, MissingCoinIdentifiersError
	}
	decodedCoinIdentifiers, err := base64.StdEncoding.DecodeString(strCoinIdentifier)
	var coinIdentifiers []string
	err = json.Unmarshal(decodedCoinIdentifiers, &coinIdentifiers)
	if err != nil {
		return nil, CoinIdentifierInvalidError
	}

	outPoints, validateErr := generateOutPointFromCoinIdentifiers(coinIdentifiers)
	if validateErr != nil {
		return nil, validateErr
	}
	for _, outPoint := range outPoints {
		items = append(items, ckbTypes.BatchLiveCellItem{OutPoint: outPoint, WithData: false})
	}

	err = s.client.BatchLiveCells(ctx, items)
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

func generateOutPointFromCoinIdentifiers(coinIdentifiers []string) ([]ckbTypes.OutPoint, *types.Error) {
	var outPoints []ckbTypes.OutPoint
	for _, option := range coinIdentifiers {
		identifier := strings.Split(option, ":")
		index, err := strconv.ParseUint(identifier[1], 10, 32)
		if err != nil {
			return nil, CoinIdentifierInvalidError
		}
		outPoint := ckbTypes.OutPoint{
			TxHash: ckbTypes.HexToHash(identifier[0]),
			Index:  uint(index),
		}
		outPoints = append(outPoints, outPoint)
	}

	return outPoints, nil
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
