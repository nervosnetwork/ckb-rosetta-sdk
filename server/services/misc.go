package services

import (
	"fmt"
	"log"
	"math"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	typesCKB "github.com/nervosnetwork/ckb-sdk-go/types"
)

var (
	NoImplementError = &types.Error{
		Code:      1,
		Message:   "Not implemented.",
		Retriable: false,
	}

	RpcError = &types.Error{
		Code:      2,
		Message:   "RPC error.",
		Retriable: true,
	}

	AddressParseError = &types.Error{
		Code:      3,
		Message:   "Address parse error.",
		Retriable: false,
	}

	SubmitError = &types.Error{
		Code:      4,
		Message:   "Submit transaction error.",
		Retriable: true,
	}

	ServerError = &types.Error{
		Code:      5,
		Message:   "Server error.",
		Retriable: true,
	}

	UnsupportedCurveTypeError = &types.Error{
		Code:      6,
		Message:   "Unsupported curve type error.",
		Retriable: false,
	}

	MissingInputOperationsError = &types.Error{
		Code:      7,
		Message:   "Must have Input type operations.",
		Retriable: false,
	}

	MissingOutputOperationsError = &types.Error{
		Code:      8,
		Message:   "Must have Output type operations.",
		Retriable: false,
	}

	InvalidInputOperationAmountValueError = &types.Error{
		Code:      9,
		Message:   "Input operation amount value must be negative.",
		Retriable: false,
	}

	InvalidCoinChangeError = &types.Error{
		Code:      10,
		Message:   "Invalid CoinChange Error.",
		Retriable: false,
	}

	InvalidOutputOperationAmountValueError = &types.Error{
		Code:      11,
		Message:   "Input operation amount value must be positive.",
		Retriable: false,
	}

	NotSupportMultisigAllLockError = &types.Error{
		Code:      12,
		Message:   "Don't support sent to multisig all lock.",
		Retriable: false,
	}

	LessThanMinCapacityError = &types.Error{
		Code:      13,
		Message:   fmt.Sprintf("Transfer amount must greater than %d CKB", MinCapacity/int64(math.Pow10(8))),
		Retriable: false,
	}

	CapacityNotEnoughError = &types.Error{
		Code:      14,
		Message:   "Capacity not enough.",
		Retriable: false,
	}

	CoinIdentifierInvalidError = &types.Error{
		Code:      15,
		Message:   "Coin identifier is invalid.",
		Retriable: false,
	}

	MissingOptionError = &types.Error{
		Code:      16,
		Message:   "Must set option in ConstructionMetadataRequest.",
		Retriable: false,
	}

	LiveCellMetadataHasDeadCellsError = &types.Error{
		Code:      17,
		Message:   "Coin identifier has dead cells.",
		Retriable: false,
	}

	CkbCurrency = &types.Currency{
		Symbol:   "CKB",
		Decimals: 8,
	}

	SupportedOperationTypes = []string{
		"Input",
		"Output",
		"Fee",
		"Reward",
	}

	MinCapacity   int64 = 6100000000
	AllErrorTypes       = []*types.Error{
		NoImplementError,
		RpcError,
		AddressParseError,
		SubmitError,
		ServerError,
		UnsupportedCurveTypeError,
		MissingInputOperationsError,
		MissingOutputOperationsError,
		InvalidInputOperationAmountValueError,
		InvalidCoinChangeError,
		InvalidOutputOperationAmountValueError,
		NotSupportMultisigAllLockError,
		LessThanMinCapacityError,
		CapacityNotEnoughError,
		CoinIdentifierInvalidError,
		LiveCellMetadataHasDeadCellsError,
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
