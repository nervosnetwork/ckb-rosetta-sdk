package builder

import (
	"bytes"
	"github.com/coinbase/rosetta-sdk-go/types"
	ckbRpc "github.com/nervosnetwork/ckb-sdk-go/rpc"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

type SignedTxBuilderInterface interface {
	Combine(txStr string, signatures []*types.Signature) (signedTxStr string, err error)
}

func NewSignedTxCombinerSecp256k1Blake160() *SignedTxCombinerSecp256k1Blake160 {
	return &SignedTxCombinerSecp256k1Blake160{}
}

type SignedTxCombinerSecp256k1Blake160 struct{}

func (c SignedTxCombinerSecp256k1Blake160) Combine(unsignedTxStr string, signatures []*types.Signature) (string, error) {
	tx, err := ckbRpc.TransactionFromString(unsignedTxStr)
	emptyWitnessArg := make([]byte, 85)
	if err != nil {
		return "", err
	}
	sIndex := 0
	for i, witness := range tx.Witnesses {
		if bytes.Compare(witness, emptyWitnessArg) == 0 {
			witnessArgs := &ckbTypes.WitnessArgs{
				Lock: signatures[sIndex].Bytes,
			}
			serializedWitness, err := witnessArgs.Serialize()
			if err != nil {
				return "", err
			}
			tx.Witnesses[i] = serializedWitness
			sIndex++
		}
	}
	signedTxStr, err := ckbRpc.TransactionString(tx)
	if err != nil {
		return "", err
	}
	return signedTxStr, nil
}
