package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-sdk-go/address"
	ckbTransaction "github.com/ququzone/ckb-sdk-go/transaction"
	ckbTypes "github.com/ququzone/ckb-sdk-go/types"
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
