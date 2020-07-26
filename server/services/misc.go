package services

import (
	"log"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-sdk-go/address"
	typesCKB "github.com/ququzone/ckb-sdk-go/types"
)

var (
	NoImplementError = &types.Error{
		Code:      1,
		Message:   "not implemented",
		Retriable: false,
	}

	RpcError = &types.Error{
		Code:      2,
		Message:   "rpc error",
		Retriable: true,
	}

	AddressError = &types.Error{
		Code:      3,
		Message:   "address error",
		Retriable: false,
	}

	SubmitError = &types.Error{
		Code:      4,
		Message:   "submit transaction error",
		Retriable: true,
	}

	ServerError = &types.Error{
		Code:      5,
		Message:   "server error",
		Retriable: true,
	}

	CkbCurrency = &types.Currency{
		Symbol:   "CKB",
		Decimals: 8,
	}
)

func GenerateAddress(network *types.NetworkIdentifier, script *typesCKB.Script) string {
	var mode = address.Mainnet
	if network.Network != "Mainnet" {
		mode = address.Testnet
	}

	addr, err := address.Generate(mode, script)
	if err != nil {
		log.Fatalf("generate address error: %v", err)
	}

	return addr
}
