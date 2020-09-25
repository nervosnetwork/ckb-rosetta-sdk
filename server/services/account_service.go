package services

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
)

// AccountAPIService implements the server.AccountAPIServicer interface.
type AccountAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
	cfg     *config.Config
}

// NewAccountAPIService creates a new instance of a AccountAPIService.
func NewAccountAPIService(network *types.NetworkIdentifier, client rpc.Client, cfg *config.Config) server.AccountAPIServicer {
	return &AccountAPIService{
		network: network,
		client:  client,
		cfg:     cfg,
	}
}

// AccountBalance implements the /account/balance endpoint.
func (s *AccountAPIService) AccountBalance(
	ctx context.Context,
	request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
	addr, err := address.Parse(request.AccountIdentifier.Address)
	if err != nil {
		return nil, AddressParseError
	}

	capacity, err := s.client.GetCellsCapacity(context.Background(), &indexer.SearchKey{
		Script:     addr.Script,
		ScriptType: indexer.ScriptTypeLock,
	})
	if err != nil {
		return nil, RpcError
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(capacity.BlockNumber),
			Hash:  capacity.BlockHash.String(),
		},
		Balances: []*types.Amount{
			{
				Value:    fmt.Sprintf("%d", capacity.Capacity),
				Currency: CkbCurrency,
			},
		},
	}, nil
}
