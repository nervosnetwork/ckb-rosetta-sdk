package builder

import (
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
