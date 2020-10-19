package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
	"math"
	"math/big"
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
	parsedAddr, err := address.Parse(request.AccountIdentifier.Address)
	var accountMetadata ckb.AccountIdentifierMetadata
	if request.AccountIdentifier.Metadata != nil {
		err := types.UnmarshalMap(request.AccountIdentifier.Metadata, &accountMetadata)
		if err != nil {
			return nil, wrapErr(DataParseError, err)
		}
	} else {
		accountMetadata = ckb.AccountIdentifierMetadata{LockType: ckb.Secp256k1Blake160Lock.String()}
	}
	if err != nil {
		return nil, AddressParseError
	}
	var cursor string
	var ckbBalance uint64
	var coins []*types.Coin
	var availableCkbBalance uint64
	udtBalance := make(map[string]*types.Amount)
	for {
		liveCells, err := s.client.GetCells(context.Background(), &indexer.SearchKey{
			Script:     parsedAddr.Script,
			ScriptType: indexer.ScriptTypeLock,
		}, indexer.SearchOrderAsc, ckb.SearchLimit, cursor)
		if err != nil {
			return nil, wrapErr(ServerError, err)
		}
		for _, cell := range liveCells.Objects {
			if accountMetadata.LockType != getLockType(cell.Output.Lock, s.cfg) {
				continue
			}
			ckbBalance += cell.Output.Capacity
			var currentReservedCkbBalance uint64
			if cell.Output.Type == nil && len(cell.OutputData) == 0 {
				availableCkbBalance += cell.Output.Capacity
			} else {
				currentReservedCkbBalance = cell.Output.OccupiedCapacity(cell.OutputData) * uint64(math.Pow10(8))
				availableCkbBalance += cell.Output.Capacity - currentReservedCkbBalance
				// is sUDT cell
				if cell.Output.Type != nil && cell.Output.Type.CodeHash.String() == s.cfg.UDT.Script.CodeHash {
					uuid := "0x" + hex.EncodeToString(cell.Output.Type.Args)
					if token, ok := s.cfg.UDT.Tokens[uuid]; ok {
						if b, ok := udtBalance[uuid]; ok {
							value, ok := big.NewInt(0).SetString(b.Value, 10)
							if !ok {
								return nil, SudtAmountInvalidError
							}
							amount, err := utils.ParseSudtAmount(cell.OutputData)
							if err != nil {
								return nil, SudtAmountInvalidError
							}
							b.Value = amount.Add(amount, value).String()
						} else {
							amount, err := utils.ParseSudtAmount(cell.OutputData)
							if err != nil {
								return nil, SudtAmountInvalidError
							}
							udtBalance[uuid] = &types.Amount{
								Value: amount.String(),
								Currency: &types.Currency{
									Symbol:   token.Symbol,
									Decimals: int32(token.Decimal),
								},
							}
						}
					}
				}
			}

			coins = append(coins, &types.Coin{
				CoinIdentifier: &types.CoinIdentifier{Identifier: fmt.Sprintf("%s:%d", cell.OutPoint.TxHash, cell.OutPoint.Index)},
				Amount: &types.Amount{
					Value:    fmt.Sprintf("%d", cell.Output.Capacity),
					Currency: CkbCurrency,
				},
			})
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

	var balances []*types.Amount
	metadata, err := types.MarshalMap(&ckb.AmountMetadata{
		AvailableCkbBalance: availableCkbBalance,
	})
	if err != nil {
		return nil, InvalidAmountMetadataError
	}
	balances = []*types.Amount{
		{
			Value:    fmt.Sprintf("%d", ckbBalance),
			Currency: CkbCurrency,
			Metadata: metadata,
		},
	}

	if len(udtBalance) > 0 {
		for _, udtAmount := range udtBalance {
			balances = append(balances, udtAmount)
		}
	}
	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: int64(tipHeader.Number),
			Hash:  tipHeader.Hash.String(),
		},
		Balances: balances,
		Coins:    coins,
	}, nil
}
