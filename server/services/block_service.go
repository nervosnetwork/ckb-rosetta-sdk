package services

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-rosetta-sdk/ckb"
	"github.com/nervosnetwork/ckb-rosetta-sdk/server/config"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"
)

// BlockAPIService implements the server.BlockAPIServicer interface.
type BlockAPIService struct {
	network *types.NetworkIdentifier
	client  rpc.Client
	cfg     *config.Config
}

// NewBlockAPIService creates a new instance of a BlockAPIService.
func NewBlockAPIService(network *types.NetworkIdentifier, client rpc.Client, cfg *config.Config) server.BlockAPIServicer {
	return &BlockAPIService{
		network: network,
		client:  client,
		cfg:     cfg,
	}
}

// Block implements the /block endpoint.
func (s *BlockAPIService) Block(
	ctx context.Context,
	request *types.BlockRequest,
) (*types.BlockResponse, *types.Error) {
	var block *ckbTypes.Block
	var err error
	if request.BlockIdentifier.Hash == nil || *request.BlockIdentifier.Hash == "" {
		if *request.BlockIdentifier.Index < 0 {
			*request.BlockIdentifier.Index = 0
		}
		block, err = s.client.GetBlockByNumber(context.Background(), uint64(*request.BlockIdentifier.Index))
	} else {
		block, err = s.client.GetBlock(context.Background(), ckbTypes.HexToHash(*request.BlockIdentifier.Hash))
	}
	if err != nil {
		return nil, RpcError
	}

	result := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Header.Number),
				Hash:  block.Header.Hash.String(),
			},
			ParentBlockIdentifier: &types.BlockIdentifier{
				Index: int64(block.Header.Number),
				Hash:  block.Header.Hash.String(),
			},
			Timestamp:    int64(block.Header.Timestamp),
			Transactions: []*types.Transaction{},
		},
	}

	if block.Header.Number > 0 {
		result.Block.ParentBlockIdentifier = &types.BlockIdentifier{
			Index: int64(block.Header.Number) - 1,
			Hash:  block.Header.ParentHash.String(),
		}
	}

	batchReq := make([]ckbTypes.BatchTransactionItem, 0)
	txHashCache := make(map[string]bool)
	for i, tx := range block.Transactions {
		if i != 0 {
			for _, input := range tx.Inputs {
				if _, ok := txHashCache[input.PreviousOutput.TxHash.String()]; !ok {
					txHashCache[input.PreviousOutput.TxHash.String()] = true
					batchReq = append(batchReq, ckbTypes.BatchTransactionItem{
						Hash:   input.PreviousOutput.TxHash,
						Result: &ckbTypes.TransactionWithStatus{},
					})
				}
			}
		}
	}
	inputTxCache := make(map[string]*ckbTypes.TransactionWithStatus)
	if len(batchReq) > 0 {
		count := len(batchReq) / 2000
		if len(batchReq)%2000 != 0 {
			count++
		}

		for i := 0; i < count; i++ {
			start := i * 2000
			end := start + 2000
			if i == count-1 {
				end = len(batchReq)
			}
			err = s.client.BatchTransactions(context.Background(), batchReq[start:end])
			if err != nil {
				return nil, RpcError
			}

		}
	}

	for _, req := range batchReq {
		if req.Error != nil || req.Result.Transaction == nil {
			return nil, RpcError
		}
		inputTxCache[req.Hash.String()] = req.Result
	}

	for i, tx := range block.Transactions {
		var transaction *types.Transaction
		optIndex := int64(0)
		if i == 0 {
			if len(tx.Outputs) > 0 {
				transaction = &types.Transaction{
					TransactionIdentifier: &types.TransactionIdentifier{
						Hash: tx.Hash.String(),
					},
					Operations: []*types.Operation{},
				}
				for _, output := range tx.Outputs {
					accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
						LockType: getLockType(output.Lock, s.cfg),
					})
					if err != nil {
						return nil, wrapErr(InvalidAccountIdentifierMetadataError, err)
					}
					transaction.Operations = append(transaction.Operations, &types.Operation{
						OperationIdentifier: &types.OperationIdentifier{
							Index: optIndex,
						},
						Type:   "Reward",
						Status: "Success",
						Account: &types.AccountIdentifier{
							Address:  GenerateAddress(s.network, output.Lock),
							Metadata: accountMetadata,
						},
						Amount: &types.Amount{
							Value:    fmt.Sprintf("%d", output.Capacity),
							Currency: CkbCurrency,
						},
					})
					optIndex++
				}
			}
		} else {
			transaction = &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: tx.Hash.String(),
				},
				Operations: []*types.Operation{},
			}
			for _, input := range tx.Inputs {
				tx := inputTxCache[input.PreviousOutput.TxHash.String()]
				if tx == nil {
					return nil, ServerError
				}

				accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
					LockType: getLockType(tx.Transaction.Outputs[input.PreviousOutput.Index].Lock, s.cfg),
				})
				if err != nil {
					return nil, wrapErr(InvalidAccountIdentifierMetadataError, err)
				}

				transaction.Operations = append(transaction.Operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: optIndex,
					},
					Type:   ckb.InputOpType,
					Status: "Success",
					Account: &types.AccountIdentifier{
						Address:  GenerateAddress(s.network, tx.Transaction.Outputs[input.PreviousOutput.Index].Lock),
						Metadata: accountMetadata,
					},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("-%d", tx.Transaction.Outputs[input.PreviousOutput.Index].Capacity),
						Currency: CkbCurrency,
					},
					CoinChange: &types.CoinChange{
						CoinIdentifier: getCoinIdentifier(input.PreviousOutput),
						CoinAction:     types.CoinSpent,
					},
				})
				optIndex++
			}
			for i, output := range tx.Outputs {
				accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
					LockType: getLockType(output.Lock, s.cfg),
				})
				if err != nil {
					return nil, wrapErr(InvalidAccountIdentifierMetadataError, err)
				}
				transaction.Operations = append(transaction.Operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: optIndex,
					},
					Type:   ckb.OutputOpType,
					Status: "Success",
					Account: &types.AccountIdentifier{
						Address:  GenerateAddress(s.network, output.Lock),
						Metadata: accountMetadata,
					},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("%d", output.Capacity),
						Currency: CkbCurrency,
					},
					CoinChange: &types.CoinChange{
						CoinIdentifier: getCoinIdentifier(&ckbTypes.OutPoint{
							TxHash: tx.Hash,
							Index:  uint(i),
						}),
						CoinAction: types.CoinCreated,
					},
				})
				optIndex++
			}
		}
		if transaction != nil {
			result.Block.Transactions = append(result.Block.Transactions, transaction)
		}
	}

	return result, nil
}

// BlockTransaction implements the /block/transaction endpoint.
func (s *BlockAPIService) BlockTransaction(
	ctx context.Context,
	request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	tx, err := s.client.GetTransaction(context.Background(), ckbTypes.HexToHash(request.TransactionIdentifier.Hash))
	if err != nil {
		return nil, RpcError
	}
	var transaction *types.Transaction
	optIndex := int64(0)
	if tx.Transaction.Inputs[0].PreviousOutput.TxHash.String() == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		if len(tx.Transaction.Outputs) > 0 {
			transaction = &types.Transaction{
				TransactionIdentifier: &types.TransactionIdentifier{
					Hash: tx.Transaction.Hash.String(),
				},
				Operations: []*types.Operation{},
			}
			for _, output := range tx.Transaction.Outputs {
				accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
					LockType: getLockType(output.Lock, s.cfg),
				})
				if err != nil {
					return nil, wrapErr(InvalidAccountIdentifierMetadataError, err)
				}
				transaction.Operations = append(transaction.Operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: optIndex,
					},
					Type:   "Reward",
					Status: "Success",
					Account: &types.AccountIdentifier{
						Address:  GenerateAddress(s.network, output.Lock),
						Metadata: accountMetadata,
					},
					Amount: &types.Amount{
						Value:    fmt.Sprintf("%d", output.Capacity),
						Currency: CkbCurrency,
					},
				})
				optIndex++
			}
		}
	} else {
		transaction = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: tx.Transaction.Hash.String(),
			},
			Operations: []*types.Operation{},
		}
		index, err := s.processTxInputs(tx.Transaction.Inputs, optIndex, transaction)
		if err != nil {
			return nil, RpcError
		}
		optIndex = index

		for i, output := range tx.Transaction.Outputs {
			accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
				LockType: getLockType(output.Lock, s.cfg),
			})
			if err != nil {
				return nil, wrapErr(InvalidAccountIdentifierMetadataError, err)
			}
			transaction.Operations = append(transaction.Operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: optIndex,
				},
				Type:   ckb.OutputOpType,
				Status: "Success",
				Account: &types.AccountIdentifier{
					Address:  GenerateAddress(s.network, output.Lock),
					Metadata: accountMetadata,
				},
				Amount: &types.Amount{
					Value:    fmt.Sprintf("%d", output.Capacity),
					Currency: CkbCurrency,
				},
				CoinChange: &types.CoinChange{
					CoinIdentifier: getCoinIdentifier(&ckbTypes.OutPoint{
						TxHash: tx.Transaction.Hash,
						Index:  uint(i),
					}),
					CoinAction: types.CoinCreated,
				},
			})
			optIndex++
		}
	}

	if transaction == nil {
		transaction = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: tx.Transaction.Hash.String(),
			},
			Operations: []*types.Operation{},
		}
	}

	return &types.BlockTransactionResponse{
		Transaction: transaction,
	}, nil
}

func (s *BlockAPIService) processTxInputs(inputs []*ckbTypes.CellInput, optIndex int64, transaction *types.Transaction) (int64, error) {
	batchReq := make([]ckbTypes.BatchTransactionItem, len(inputs))
	for i, input := range inputs {
		batchReq[i] = ckbTypes.BatchTransactionItem{
			Hash:   input.PreviousOutput.TxHash,
			Result: &ckbTypes.TransactionWithStatus{},
		}
	}

	err := s.client.BatchTransactions(context.Background(), batchReq)
	if err != nil {
		return 0, err
	}

	for i, input := range inputs {
		req := batchReq[i]
		if req.Error != nil {
			return 0, req.Error
		}
		accountMetadata, err := types.MarshalMap(&ckb.AccountIdentifierMetadata{
			LockType: getLockType(req.Result.Transaction.Outputs[input.PreviousOutput.Index].Lock, s.cfg),
		})
		if err != nil {
			return 0, err
		}

		transaction.Operations = append(transaction.Operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: optIndex,
			},
			Type:   ckb.InputOpType,
			Status: "Success",
			Account: &types.AccountIdentifier{
				Address:  GenerateAddress(s.network, req.Result.Transaction.Outputs[input.PreviousOutput.Index].Lock),
				Metadata: accountMetadata,
			},
			Amount: &types.Amount{
				Value:    fmt.Sprintf("-%d", req.Result.Transaction.Outputs[input.PreviousOutput.Index].Capacity),
				Currency: CkbCurrency,
			},
			CoinChange: &types.CoinChange{
				CoinIdentifier: getCoinIdentifier(input.PreviousOutput),
				CoinAction:     types.CoinSpent,
			},
		})
		optIndex++
	}

	return optIndex, nil
}
