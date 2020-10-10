package services

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
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
	var cursor string
	var ckbBalance uint64
	var ckbCoins []*types.Coin
	for {
		liveCells, err := s.client.GetCells(context.Background(), &indexer.SearchKey{
			Script:     addr.Script,
			ScriptType: indexer.ScriptTypeLock,
		}, indexer.SearchOrderAsc, ckb.SearchLimit, cursor)
		if err != nil {
			return nil, wrapErr(ServerError, err)
		}
		for _, cell := range liveCells.Objects {
			ckbBalance += cell.Output.Capacity
			if cell.Output.Type == nil && len(cell.OutputData) == 0 {
				ckbCoins = append(ckbCoins, &types.Coin{
					CoinIdentifier: &types.CoinIdentifier{Identifier: fmt.Sprintf("%s:%d", cell.OutPoint.TxHash, cell.OutPoint.Index)},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("%d", cell.Output.Capacity),
						Currency: CkbCurrency,
					},
				})
			}
		}
		if len(liveCells.Objects) < ckb.SearchLimit || liveCells.LastCursor == "" {
			break
		}
		cursor = liveCells.LastCursor
	}

	tipHeader, err := s.client.GetTipHeader(context.Background())
	if err != nil {
		return nil, wrapErr(RpcError, err)
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(tipHeader.Number),
			Hash:  tipHeader.Hash.String(),
		},
		Balances: []*types.Amount{
			{
				Value:    fmt.Sprintf("%d", ckbBalance),
				Currency: CkbCurrency,
			},
		},
		Coins: ckbCoins,
	}, nil
}
