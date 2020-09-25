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
	var metadata ckb.PreprocessMetadata
	if err := types.UnmarshalMap(request.Metadata, &metadata); err != nil {
		return nil, InvalidPreprocessMetadataError
	}

	txSizeEstimatorFactory := new(factory.TxSizeEstimatorFactory)
	txSizeEstimator := txSizeEstimatorFactory.CreateTxSizeEstimator(metadata.TxType)
	if txSizeEstimator == nil {
		return nil, wrapErr(UnsupportedTxTypeError, fmt.Errorf("unsupported tx type: %s", metadata.TxType))
	}
	estimatedTxSize, err := txSizeEstimator.EstimatedTxSize(request.Operations)
	if err != nil {
		return nil, wrapErr(DataParseError, err)
	}

	options, err := types.MarshalMap(&ckb.PreprocessOptions{
		TxType:                 metadata.TxType,
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
	if !SupportedTxTypes[options.TxType] {
		return nil, wrapErr(UnsupportedTxTypeError, fmt.Errorf("unsupported tx type: %s", options.TxType))
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
		TxType: options.TxType,
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

	txType, validateErr := validateTxType(request.Metadata)
	if validateErr != nil {
		return nil, validateErr
	}
	unsignedTxBuilderFactory := factory.UnsignedTxBuilderFactory{}
	inputOperations, outputOperations := separateInputAndOutput(request.Operations)
	unsignedTxBuilder := unsignedTxBuilderFactory.CreateUnsignedTxBuilder(txType, s.cfg, inputOperations, outputOperations)
	if unsignedTxBuilder == nil {
		return nil, wrapErr(UnsupportedTxTypeError, fmt.Errorf("unsupported tx type: %s", txType))
	}
	unsignedTx, err := unsignedTxBuilder.Build()
	if err != nil {
		return nil, wrapErr(UnsignedTxBuildError, err)
	}
	signingPayloadBuilderFactory := SigningPayloadBuilderFactory{}
	signingPayloadBuilder := signingPayloadBuilderFactory.CreateSigningPayloadBuilder(txType)
	payloads, err := signingPayloadBuilder.BuildSigningPayload(inputOperations, unsignedTx)
	if err != nil {
		return nil, wrapErr(SigningPayloadBuildError, err)
	}
	txString, err := ckbRpc.TransactionString(unsignedTx)
	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: txString,
		Payloads:            payloads,
	}, nil
}

// ConstructionCombine implements the /construction/combine endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	tx, err := ckbRpc.TransactionFromString(request.UnsignedTransaction)
	if err != nil {
		return nil, &types.Error{
			Code:      11,
			Message:   fmt.Sprintf("can not decode transaction string: %s", request.UnsignedTransaction),
			Retriable: false,
		}
	}
	index := 0
	for i, witness := range tx.Witnesses {
		if len(witness) != 0 {
			tx.Witnesses[i] = request.Signatures[index].Bytes
			index++
		}
	}

	txString, err := ckbRpc.TransactionString(tx)
	if err != nil {
		return nil, ServerError
	}
	return &types.ConstructionCombineResponse{
		SignedTransaction: txString,
	}, nil
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	tx, err := ckbRpc.TransactionFromString(request.Transaction)
	if err != nil {
		return nil, &types.Error{
			Code:      11,
			Message:   fmt.Sprintf("can not decode transaction string: %s", request.Transaction),
			Retriable: false,
		}
	}

	signers := make(map[string]bool)
	index := int64(0)
	operations := make([]*types.Operation, 0)
	for _, input := range tx.Inputs {
		ptx, err := s.client.GetTransaction(ctx, input.PreviousOutput.TxHash)
		if err != nil {
			return nil, ServerError
		}

		addr := GenerateAddress(s.network, ptx.Transaction.Outputs[input.PreviousOutput.Index].Lock)
		signers[addr] = true
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: index,
			},
			Type:   "Transfer",
			Status: "Success",
			Account: &types.AccountIdentifier{
				Address: addr,
			},
			Amount: &types.Amount{
				Value:    fmt.Sprintf("-%d", ptx.Transaction.Outputs[input.PreviousOutput.Index].Capacity),
				Currency: CkbCurrency,
			},
		})
		index++
	}
	for _, output := range tx.Outputs {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: index,
			},
			Type:   "Transfer",
			Status: "Success",
			Account: &types.AccountIdentifier{
				Address: GenerateAddress(s.network, output.Lock),
			},
			Amount: &types.Amount{
				Value:    fmt.Sprintf("%d", output.Capacity),
				Currency: CkbCurrency,
			},
		})
		index++
	}

	addresses := make([]string, 0, len(signers))
	for addr := range signers {
		addresses = append(addresses, addr)
	}

	return &types.ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: nil,
	}, nil
}

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	tx, err := ckbRpc.TransactionFromString(request.SignedTransaction)
	if err != nil {
		return nil, &types.Error{
			Code:      11,
			Message:   fmt.Sprintf("can not decode transaction string: %s", request.SignedTransaction),
			Retriable: false,
		}
	}
	hash, err := tx.ComputeHash()
	if err != nil {
		return nil, &types.Error{
			Code:      12,
			Message:   fmt.Sprintf("compute hash error: %v", err),
			Retriable: false,
		}
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
		return nil, &types.Error{
			Code:      4,
			Message:   fmt.Sprintf("submit transaction error: %v", err),
			Retriable: true,
		}
	}

	hash, err := s.client.SendTransaction(ctx, tx)
	if err != nil {
		return nil, &types.Error{
			Code:      4,
			Message:   fmt.Sprintf("submit transaction error: %v", err),
			Retriable: true,
		}
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
		return nil, &types.Error{
			Code:      5,
			Message:   fmt.Sprintf("server error: %v", err),
			Retriable: true,
		}
	}

	prefix := address.Mainnet
	if s.network.Network != "Mainnet" {
		prefix = address.Testnet
	}

	_, err = address.Generate(prefix, &ckbTypes.Script{
		CodeHash: ckbTypes.HexToHash("0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8"),
		HashType: ckbTypes.HashTypeType,
		Args:     args,
	})
	if err != nil {
		return nil, &types.Error{
			Code:      5,
			Message:   fmt.Sprintf("server error: %v", err),
			Retriable: true,
		}
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: nil,
	}, nil
}
