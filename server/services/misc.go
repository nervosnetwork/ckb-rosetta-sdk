package services

import (
	"fmt"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
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
		Message:   "Output operation amount value must be positive.",
		Retriable: false,
	}

	NotSupportMultisigAllLockError = &types.Error{
		Code:      12,
		Message:   "Don't support sent to multisig all lock.",
		Retriable: false,
	}

	LessThanMinCapacityError = &types.Error{
		Code:      13,
		Message:   fmt.Sprintf("Transfer amount must greater than %d CKB", MinCapacity/uint64(math.Pow10(8))),
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

	InvalidTypeScriptError = &types.Error{
		Code:      18,
		Message:   "Invalid type script error.",
		Retriable: false,
	}

	InvalidOutputDataError = &types.Error{
		Code:      19,
		Message:   "Invalid output data error.",
		Retriable: false,
	}

	MissingSigningTypeError = &types.Error{
		Code:      20,
		Message:   "Missing signing type error.",
		Retriable: false,
	}

	InvalidPreprocessMetadataError = &types.Error{
		Code:      21,
		Message:   "invalid preprocess metadata error.",
		Retriable: false,
	}

	InvalidPreprocessOptionsError = &types.Error{
		Code:      22,
		Message:   "invalid preprocess options error.",
		Retriable: false,
	}

	InvalidConstructionMetadataError = &types.Error{
		Code:      23,
		Message:   "invalid construction metadata error.",
		Retriable: false,
	}

	InvalidOperationMetadataError = &types.Error{
		Code:      24,
		Message:   "invalid operation metadata error.",
		Retriable: false,
	}

	DataParseError = &types.Error{
		Code:      25,
		Message:   "invalid operation metadata error.",
		Retriable: false,
	}

	UnsupportedConstructionTypeError = &types.Error{
		Code:      26,
		Message:   "unsupported construction type error.",
		Retriable: false,
	}

	ScriptHashComputedError = &types.Error{
		Code:      27,
		Message:   "script hash computed error.",
		Retriable: false,
	}

	UnsignedTxBuildError = &types.Error{
		Code:      28,
		Message:   "unsigned tx build error.",
		Retriable: false,
	}

	SignMessagesBuildError = &types.Error{
		Code:      29,
		Message:   "signing messages build error.",
		Retriable: false,
	}

	SigningPayloadBuildError = &types.Error{
		Code:      30,
		Message:   "signing payload build error.",
		Retriable: false,
	}

	SignedTxBuildError = &types.Error{
		Code:      31,
		Message:   "signed tx build error.",
		Retriable: false,
	}

	TransactionParseError = &types.Error{
		Code:      32,
		Message:   "transaction parse error.",
		Retriable: false,
	}

	InvalidAccountIdentifierMetadataError = &types.Error{
		Code:      33,
		Message:   "invalid account identifier metadata error.",
		Retriable: false,
	}

	ComputeHashError = &types.Error{
		Code:      34,
		Message:   "compute hash error.",
		Retriable: false,
	}

	AddressGenerationError = &types.Error{
		Code:      35,
		Message:   "Address generation error.",
		Retriable: false,
	}

	InvalidDeriveMetadataError = &types.Error{
		Code:      36,
		Message:   "invalid derive metadata error.",
		Retriable: false,
	}

	UnsupportedNetworkError = &types.Error{
		Code:      37,
		Message:   "unsupported network error.",
		Retriable: false,
	}

	SudtAmountInvalidError = &types.Error{
		Code:      38,
		Message:   "sudt amount invalid error.",
		Retriable: false,
	}

	InvalidAmountMetadataError = &types.Error{
		Code:      39,
		Message:   "invalid amount metadata error.",
		Retriable: false,
	}

	CkbCurrency = &types.Currency{
		Symbol:   "CKB",
		Decimals: 8,
	}

	SupportedOperationTypes = []string{
		ckb.InputOpType,
		ckb.OutputOpType,
		ckb.RewardOpType,
	}

	SupportedNetworks = map[string]bool{
		"mainnet": true,
		"testnet": true,
		"dev":     true,
	}

	SupportedConstructionTypes = map[string]bool{
		ckb.TransferCKB: true,
	}

	MinCapacity   uint64 = 6100000000
	AllErrorTypes        = []*types.Error{
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
		MissingOptionError,
		InvalidTypeScriptError,
		InvalidOutputDataError,
		MissingSigningTypeError,
		InvalidPreprocessMetadataError,
		InvalidPreprocessOptionsError,
		InvalidConstructionMetadataError,
		InvalidOperationMetadataError,
		DataParseError,
		UnsupportedConstructionTypeError,
		ScriptHashComputedError,
		UnsignedTxBuildError,
		SignMessagesBuildError,
		SigningPayloadBuildError,
		SignedTxBuildError,
		TransactionParseError,
		InvalidAccountIdentifierMetadataError,
		AddressGenerationError,
		UnsupportedNetworkError,
		SudtAmountInvalidError,
		InvalidAmountMetadataError,
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
