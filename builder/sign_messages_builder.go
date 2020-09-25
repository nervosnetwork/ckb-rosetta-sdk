package builder

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SignMessagesBuilderInterface interface {
	BuildSignMessages(tx *ckbTypes.Transaction, inputOperations []*types.Operation) ([][]byte, error)
}

func NewSecp256k1Blake160SignMessagesBuilder() *Secp256k1Blake160SignMessagesBuilder {
	return &Secp256k1Blake160SignMessagesBuilder{}
}

type Secp256k1Blake160SignMessagesBuilder struct{}

func (s Secp256k1Blake160SignMessagesBuilder) BuildSignMessages(tx *ckbTypes.Transaction, inputOperations []*types.Operation) ([][]byte, error) {
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
