package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTransaction "github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
	"strings"
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

func isBlake160SighashAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == ckbTransaction.SECP256K1_BLAKE160_SIGHASH_ALL_TYPE_HASH
}

func isBlake160MultisigAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == ckbTransaction.SECP256K1_BLAKE160_MULTISIG_ALL_TYPE_HASH
}

func generateInputOutPointsOption(request *types.ConstructionPreprocessRequest) ([]outPoint, *types.Error) {
	inputOperations := operationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Input"
	})
	var outPoints []outPoint
	for _, operation := range inputOperations {
		identifier := strings.Split(operation.CoinChange.CoinIdentifier.Identifier, ":")
		index, err := strconv.ParseUint(identifier[1], 10, 64)
		if err != nil {
			return nil, CoinIdentifierInvalidError
		}
		outPoint := outPoint{TxHash: ckbTypes.HexToHash(identifier[0]), Index: hexutil.Uint(index)}
		outPoints = append(outPoints, outPoint)
	}
	return outPoints, nil
}
