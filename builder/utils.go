package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
	"strconv"
	"strings"
)

func GenerateOutPointFromCoinIdentifier(identifier string) (*ckbTypes.OutPoint, error) {
	splittedIdentifier := strings.Split(identifier, ":")
	index, err := strconv.ParseUint(splittedIdentifier[1], 10, 32)
	if err != nil {
		return nil, err
	}
	return &ckbTypes.OutPoint{
		TxHash: ckbTypes.HexToHash(splittedIdentifier[0]),
		Index:  uint(index),
	}, nil
}

func BuildIndexGroups(inputOperations []*types.Operation) ([][]int, error) {
	lockScriptHashes := make(map[ckbTypes.Hash][]int)
	for i, operation := range inputOperations {
		parsedAddress, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, err
		}
		lockHash, err := parsedAddress.Script.Hash()
		if err != nil {
			return nil, err
		}
		if _, ok := lockScriptHashes[lockHash]; !ok {
			lockScriptHashes[lockHash] = append(lockScriptHashes[lockHash], i)
		}
	}

	indexGroups := make([][]int, 0, len(lockScriptHashes))
	for _, index := range lockScriptHashes {
		indexGroups = append(indexGroups, index)
	}
	return indexGroups, nil
}
