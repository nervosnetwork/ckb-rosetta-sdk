package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/asserter"
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

// ConstructionPreprocess implements the /construction/preprocess endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	var input int64
	var output int64
	vinOperations := operationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Vin"
	})
	voutOperations := operationFilter(request.Operations, func(operation *types.Operation) bool {
		return operation.Type == "Vout"
	})

	if len(vinOperations) == 0 {
		return nil, MissingVinOperationsError
	}

	for _, operation := range vinOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount >= 0 {
			return nil, InvalidVinOperationAmountValueError
		}
		err = asserter.CoinChange(operation.CoinChange)
		if err != nil {
			return nil, InvalidCoinChangeError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, AddressParseError
		}
		// do not support send to multisig all lock
		if isBlake160MultisigAllLock(addr) {
			return nil, NotSupportMultisigAllLockError
		}

		input += -amount
	}

	if len(voutOperations) == 0 {
		return nil, MissingVoutOperationsError
	}

	for _, operation := range voutOperations {
		amount, err := strconv.ParseInt(operation.Amount.Value, 10, 64)
		if err != nil || amount <= 0 {
			return nil, InvalidVoutOperationAmountValueError
		}
		addr, err := address.Parse(operation.Account.Address)
		if err != nil {
			return nil, AddressParseError
		}
		if isBlake160SighashAllLock(addr) {
			if amount < MinCapacity {
				return nil, LessThanMinCapacityError
			}
		}

		output += amount
	}

	if input <= output {
		return nil, CapacityNotEnoughError
	}

	return &types.ConstructionPreprocessResponse{}, nil
}

func isBlake160SighashAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == "0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8"
}

func isBlake160MultisigAllLock(addr *address.ParsedAddress) bool {
	return addr.Script.HashType == ckbTypes.HashTypeType &&
		addr.Script.CodeHash.String() == "0x5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8"
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

func operationFilter(arr []*types.Operation, cond func(*types.Operation) bool) []*types.Operation {
	var result []*types.Operation
	for i := range arr {
		if cond(arr[i]) {
			result = append(result, arr[i])
		}
	}
	return result
}
