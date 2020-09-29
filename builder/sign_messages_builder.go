package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SignMessagesBuilder interface {
	BuildSignMessages(tx *ckbTypes.Transaction, inputOperations []*types.Operation) ([][]byte, error)
}

func NewSignMessagesBuilderSecp256k1Blake160() *SignMessagesBuilderSecp256k1Blake160 {
	return &SignMessagesBuilderSecp256k1Blake160{}
}

type SignMessagesBuilderSecp256k1Blake160 struct{}

func (s SignMessagesBuilderSecp256k1Blake160) BuildSignMessages(tx *ckbTypes.Transaction, inputOperations []*types.Operation) ([][]byte, error) {
	indexGroups, err := BuildIndexGroups(inputOperations)
	if err != nil {
		return nil, err
	}
	var messages [][]byte
	for _, indexGroup := range indexGroups {
		message, err := transaction.SingleSegmentSignMessage(tx, indexGroup[0], indexGroup[0]+len(indexGroup), transaction.EmptyWitnessArg)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}
