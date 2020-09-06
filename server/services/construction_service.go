package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ququzone/ckb-rich-sdk-go/indexer"
	"github.com/ququzone/ckb-rich-sdk-go/rpc"
	"github.com/ququzone/ckb-sdk-go/address"
	"github.com/ququzone/ckb-sdk-go/crypto/blake2b"
	ckbRpc "github.com/ququzone/ckb-sdk-go/rpc"
	ckbTransaction "github.com/ququzone/ckb-sdk-go/transaction"
	ckbTypes "github.com/ququzone/ckb-sdk-go/types"
	"github.com/ququzone/ckb-sdk-go/utils"
)

// ConstructionAPIService implements the server.ConstructionAPIService interface.
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

// ConstructionPreprocess implements the /construction/preprocess endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	inputTotalAmount, err := validateInputOperations(request)
	if err != nil {
		return nil, err
	}

	outputTotalAmount, err := validateOutputOperations(request)
	if err != nil {
		return nil, err
	}

	err = validateCapacity(inputTotalAmount, outputTotalAmount)
	if err != nil {
		return nil, err
	}

	outPointsOption, err := generateInputOutPointsOption(request)
	if err != nil {
		return nil, err
	}
	response := &types.ConstructionPreprocessResponse{
		Options: make(map[string]interface{}),
	}
	response.Options["outPoints"] = outPointsOption

	return response, nil
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	// TODO use batchRequest get liveCells, if exists unknown cell return error
	return &types.ConstructionMetadataResponse{
		Metadata: map[string]interface{}{},
	}, nil
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {

	systemScripts, err := utils.NewSystemScripts(s.client)
	if err != nil {
		return nil, ServerError
	}

	tx := ckbTransaction.NewSecp256k1SingleSigTx(systemScripts)
	payloads := make([]*types.SigningPayload, 0)
	for _, operation := range request.Operations {
		addr, err := address.Parse(operation.Account.Address)
		if err != nil || addr.Script.HashType != ckbTypes.HashTypeType ||
			addr.Script.CodeHash.String() != "0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8" {
			return nil, &types.Error{
				Code:      7,
				Message:   fmt.Sprintf("error address: %s", operation.Account.Address),
				Retriable: true,
			}
		}

		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount == 0 {
			return nil, &types.Error{
				Code:      8,
				Message:   fmt.Sprintf("error amount: %s", operation.Amount.Value),
				Retriable: true,
			}
		}
		if amount > 0 {
			if amount < MinCapacity {
				return nil, &types.Error{
					Code:      9,
					Message:   fmt.Sprintf("to small amount: %s", operation.Amount.Value),
					Retriable: true,
				}
			}

			tx.Outputs = append(tx.Outputs, &ckbTypes.CellOutput{
				Capacity: uint64(amount),
				Lock:     addr.Script,
			})
			tx.OutputsData = append(tx.OutputsData, []byte{})
		} else {
			amount = -amount

			liveCells, err := s.client.GetCells(context.Background(), &indexer.SearchKey{
				Script:     addr.Script,
				ScriptType: indexer.ScriptTypeLock,
			}, indexer.SearchOrderAsc, 1000, "")
			if err != nil {
				return nil, ServerError
			}

			fromCapacity := int64(0)
			inputs := make([]*ckbTypes.Cell, 0)
			for _, cell := range liveCells.Objects {
				if cell.Output.Type == nil && len(cell.OutputData) == 0 {
					fromCapacity += int64(cell.Output.Capacity)
					inputs = append(inputs, &ckbTypes.Cell{
						OutPoint: &ckbTypes.OutPoint{
							TxHash: cell.OutPoint.TxHash,
							Index:  cell.OutPoint.Index,
						},
					})
					if fromCapacity < amount {
						continue
					}
					if fromCapacity-amount >= MinCapacity {
						tx.Outputs = append(tx.Outputs, &ckbTypes.CellOutput{
							Capacity: uint64(fromCapacity - amount),
							Lock:     addr.Script,
						})
						tx.OutputsData = append(tx.OutputsData, []byte{})
					}
				}
			}

			group, witnessArgs, err := ckbTransaction.AddInputsForTransaction(tx, inputs)
			if err != nil {
				return nil, ServerError
			}

			// TODO remove to transaction generated
			payload, err := ckbTransaction.SingleSegmentSignMessage(tx, group[0], group[0]+len(group), witnessArgs)
			if err != nil {
				return nil, ServerError
			}
			payloads = append(payloads, &types.SigningPayload{
				Address:       operation.Account.Address,
				Bytes:         payload,
				SignatureType: types.EcdsaRecovery,
			})
		}
	}

	txString, err := ckbRpc.TransactionString(tx)
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
		Operations: operations,
		Signers:    addresses,
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
