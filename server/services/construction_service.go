package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"
	"github.com/ququzone/ckb-sdk-go/address"
	"github.com/ququzone/ckb-sdk-go/crypto/blake2b"
	ckbTypes "github.com/ququzone/ckb-sdk-go/types"
)

// ConstructionAPIService implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
}

// NewConstructionAPIService creates a new instance of a ConstructionAPIService.
func NewConstructionAPIService(network *types.NetworkIdentifier, client rpc.Client) server.ConstructionAPIServicer {
	return &ConstructionAPIService{
		network: network,
		client:  client,
	}
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	context.Context,
	*types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	return &types.ConstructionMetadataResponse{
		Metadata: map[string]interface{}{},
	}, nil
}

// ConstructionCombine implements the /construction/combine endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	context.Context,
	*types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	panic("implement me")
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

	addr, err := address.Generate(prefix, &ckbTypes.Script{
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
		Address: addr,
	}, nil
}

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructionAPIService) ConstructionHash(
	context.Context,
	*types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	panic("implement me")
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	context.Context,
	*types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	panic("implement me")
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	context.Context,
	*types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	panic("implement me")
}

// ConstructionPreprocess implements the /construction/preprocess endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	context.Context,
	*types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	panic("implement me")
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
