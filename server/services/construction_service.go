package services

import (
	"context"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"
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

// ConstructionSubmit implements the /construction/submit endpoint.
func (s *ConstructionAPIService) ConstructionSubmit(
	ctx context.Context,
	request *types.ConstructionSubmitRequest,
) (*types.ConstructionSubmitResponse, *types.Error) {
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

	return &types.ConstructionSubmitResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: hash.String(),
		},
	}, nil
}
