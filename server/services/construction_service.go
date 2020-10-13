package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/factory"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/crypto/blake2b"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	ckbRpc "github.com/nervosnetwork/ckb-sdk-go/rpc"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

// ConstructionAPIService implements the server.ConstructionAPIService interface.
type ConstructionAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
	cfg     *config.Config
}

// NewConstructionAPIService creates a new instance of a ConstructionAPIService.
func NewConstructionAPIService(network *types.NetworkIdentifier, client rpc.Client, cfg *config.Config) server.ConstructionAPIServicer {
	return &ConstructionAPIService{
		network: network,
		client:  client,
		cfg:     cfg,
	}
}

// ConstructionPreprocess implements the /construction/preprocess endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	inputTotalAmount, validateErr := validateInputOperations(request.Operations, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}

	outputTotalAmount, validateErr := validateOutputOperations(request.Operations, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}

	validateErr = validateCapacity(inputTotalAmount, outputTotalAmount)
	if validateErr != nil {
		return nil, validateErr
	}
	constructionType, validateErr := getConstructionType(request.Operations, nil, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}
	txSizeEstimatorFactory := new(factory.TxSizeEstimatorFactory)
	txSizeEstimator := txSizeEstimatorFactory.CreateTxSizeEstimator(constructionType)
	if txSizeEstimator == nil {
		return nil, wrapErr(UnsupportedConstructionTypeError, fmt.Errorf("unsupported construction type: %s", constructionType))
	}
	estimatedTxSize, err := txSizeEstimator.EstimatedTxSize(request.Operations)
	if err != nil {
		return nil, wrapErr(DataParseError, err)
	}

	options, err := types.MarshalMap(&ckb.PreprocessOptions{
		ConstructionType:       constructionType,
		EstimatedTxSize:        estimatedTxSize,
		SuggestedFeeMultiplier: request.SuggestedFeeMultiplier,
	})
	if err != nil {
		return nil, InvalidPreprocessOptionsError
	}

	return &types.ConstructionPreprocessResponse{
		Options: options,
	}, nil
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if request.Options == nil {
		return nil, MissingOptionError
	}

	var options ckb.PreprocessOptions
	if err := types.UnmarshalMap(request.Options, &options); err != nil {
		return nil, InvalidPreprocessOptionsError
	}
	if !SupportedConstructionTypes[options.ConstructionType] {
		return nil, wrapErr(UnsupportedConstructionTypeError, fmt.Errorf("unsupported construction type: %s", options.ConstructionType))
	}
	shannonsPerKB := float64(ckb.MinFeeRate)
	if options.SuggestedFeeMultiplier != nil {
		shannonsPerKB *= *options.SuggestedFeeMultiplier
	}
	shannonsPerB := shannonsPerKB / ckb.BytesInKb
	estimatedFee := shannonsPerB * float64(options.EstimatedTxSize)
	suggestedFee := &types.Amount{
		Value:    fmt.Sprintf("%d", uint64(estimatedFee)),
		Currency: CkbCurrency,
	}

	metadata, err := types.MarshalMap(&ckb.ConstructionMetadata{
		ConstructionType: options.ConstructionType,
	})
	if err != nil {
		return nil, InvalidConstructionMetadataError
	}

	return &types.ConstructionMetadataResponse{
		Metadata:     metadata,
		SuggestedFee: []*types.Amount{suggestedFee},
	}, nil
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	inputTotalAmount, validateErr := validateInputOperations(request.Operations, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}

	outputTotalAmount, validateErr := validateOutputOperations(request.Operations, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}

	validateErr = validateCapacity(inputTotalAmount, outputTotalAmount)
	if validateErr != nil {
		return nil, validateErr
	}

	constructionType, validateErr := validateConstructionType(request.Metadata)
	if validateErr != nil {
		return nil, validateErr
	}
	unsignedTxBuilderFactory := factory.UnsignedTxBuilderFactory{}
	inputOperations, outputOperations := separateInputAndOutput(request.Operations)
	unsignedTxBuilder := unsignedTxBuilderFactory.CreateUnsignedTxBuilder(constructionType, s.cfg, inputOperations, outputOperations)
	if unsignedTxBuilder == nil {
		return nil, wrapErr(UnsupportedConstructionTypeError, fmt.Errorf("unsupported construction type: %s", constructionType))
	}
	unsignedTx, err := unsignedTxBuilder.Build()
	if err != nil {
		return nil, wrapErr(UnsignedTxBuildError, err)
	}
	signingPayloadBuilderFactory := factory.SigningPayloadBuilderFactory{}
	signingPayloadBuilder := signingPayloadBuilderFactory.CreateSigningPayloadBuilder(constructionType)
	payloads, err := signingPayloadBuilder.BuildSigningPayload(inputOperations, unsignedTx)
	if err != nil {
		return nil, wrapErr(SigningPayloadBuildError, err)
	}
	txString, err := ckbRpc.TransactionString(unsignedTx)
	rTxStr, validateErr := rTxStringForPayload(txString, request.Operations)
	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: rTxStr,
		Payloads:            payloads,
	}, nil
}

// ConstructionCombine implements the /construction/combine endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	unsignedTxCombinerFactory := factory.SignedTxBuilder{}
	constructionType, validateErr := getConstructionType(nil, request.Signatures, s.cfg)
	if validateErr != nil {
		return nil, validateErr
	}
	signedTxBuilder := unsignedTxCombinerFactory.CreateSignedTxBuilder(constructionType)
	signedTxStr, err := signedTxBuilder.Combine(request.UnsignedTransaction, request.Signatures)
	if err != nil {
		return nil, wrapErr(SignedTxBuildError, err)
	}
	rTx, err := rosettaTransactionFromString(request.UnsignedTransaction)
	if err != nil {
		return nil, wrapErr(TransactionParseError, err)
	}
	signedTx, err := ckbRpc.TransactionFromString(signedTxStr)
	if err != nil {
		return nil, wrapErr(TransactionParseError, err)
	}
	rTx.Witnesses = signedTx.Witnesses
	rTxStr, err := rTxString(rTx)
	if err != nil {
		return nil, wrapErr(TransactionParseError, err)
	}

	signedRtx, validateErr := rTxStringForCombine(rTxStr, request.Signatures)
	return &types.ConstructionCombineResponse{
		SignedTransaction: signedRtx,
	}, nil
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	return s.parseTransaction(request)
}

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	tx, err := ckbRpc.TransactionFromString(request.SignedTransaction)
	if err != nil {
		return nil, TransactionParseError
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, wrapErr(ComputeHashError, fmt.Errorf("error computing hash: %v", err))
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: hash.String(),
		},
	}, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (s *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	tx, err := ToTransaction(request.SignedTransaction)
	if err != nil {
		return nil, wrapErr(SubmitError, err)
	}

	hash, err := s.client.SendTransaction(ctx, tx)
	if err != nil {
		return nil, wrapErr(SubmitError, err)
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: hash.String(),
		},
	}, nil
}

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructionAPIService) ConstructionDerive(
	ctx context.Context,
	request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	if request.PublicKey.CurveType != types.Secp256k1 {
		return nil, UnsupportedCurveTypeError
	}

	args, err := blake2b.Blake160(request.PublicKey.Bytes)
	if err != nil {
		return nil, wrapErr(ServerError, err)
	}

	if _, ok := SupportedNetworks[s.network.Network]; !ok {
		return nil, wrapErr(UnsupportedNetworkError, fmt.Errorf("network %s not supported", s.network.Network))
	}
	prefix := address.Mainnet
	if s.network.Network != "mainnet" {
		prefix = address.Testnet
	}

	var script *ckbTypes.Script
	var lockType string
	if request.Metadata != nil {
		var metadata ckb.DeriveMetadata
		if err := types.UnmarshalMap(request.Metadata, &metadata); err != nil {
			return nil, wrapErr(InvalidDeriveMetadataError, err)
		}
		script, err = toScript(metadata.Script)
		if err != nil {
			return nil, wrapErr(ServerError, err)
		}
		lockType = getLockType(script, s.cfg)
	} else {
		script = &ckbTypes.Script{
			CodeHash: ckbTypes.HexToHash(s.cfg.Secp256k1Blake160.Script.CodeHash),
			HashType: ckbTypes.ScriptHashType(s.cfg.Secp256k1Blake160.Script.HashType),
			Args:     args,
		}
		lockType = ckb.Secp256k1Blake160Lock.String()
	}

	addr, err := address.Generate(prefix, script)
	if err != nil {
		return nil, wrapErr(ServerError, err)
	}

	metadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
		LockType: lockType,
	})

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address:  addr,
			Metadata: metadata,
		},
	}, nil
}

func (s *ConstructionAPIService) parseTransaction(request *types.ConstructionParseRequest) (*types.ConstructionParseResponse, *types.Error) {
	signedTx, err := rosettaTransactionFromString(request.Transaction)
	if err != nil {
		return nil, TransactionParseError
	}
	var operations []*types.Operation

	for i, input := range signedTx.Inputs {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{Index: int64(len(operations))},
			Type:                ckb.InputOpType,
			Account:             signedTx.InputAccounts[i],
			Amount:              signedTx.InputAmounts[i],
			CoinChange: &types.CoinChange{
				CoinIdentifier: getCoinIdentifier(input.PreviousOutput),
				CoinAction:     types.CoinSpent,
			},
		})
	}
	for i := range signedTx.Outputs {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(len(operations)),
			},
			Type:    ckb.OutputOpType,
			Account: signedTx.OutputAccounts[i],
			Amount:  signedTx.OutputAmounts[i],
		})
	}

	return &types.ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signedTx.AccountIdentifierSigners,
	}, nil
}
