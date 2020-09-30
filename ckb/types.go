package ckb

import ckbTypes "github.com/nervosnetwork/ckb-sdk-go/types"

const (
	InputOpType              = "INPUT"
	OutputOpType             = "OUTPUT"
	RewardOpType             = "Reward"
	BaseTxSize               = 68 // empty cellDeps + empty headerDeps + empty inputs + empty outputs + empty outputs_data + empty witnesses + version
	InputSize                = 44
	HeaderDepSize            = 32
	CellDepSize              = 37
	SerializedOffsetByteSize = 4
	BytesInKb                = 1000
	MinFeeRate               = 1000 // shannons/KB
	TransferCKB              = "TransferCKB"
)

const (
	Secp256k1Blake160Lock LockType = iota
)

func (l LockType) String() string {
	return [...]string{"Secp256k1Blake160Lock"}[l]
}

type LockType int

type PreprocessOptions struct {
	ConstructionType       string   `json:"construction_type"`
	EstimatedTxSize        uint64   `json:"estimated_tx_size"`
	SuggestedFeeMultiplier *float64 `json:"suggested_fee_multiplier"`
}

type ConstructionMetadata struct {
	ConstructionType string `json:"construction_type"`
}

type OperationMetadata struct {
	Data string           `json:"data"`
	Type *ckbTypes.Script `json:"type"`
}

type AccountIdentifierMetadata struct {
	LockType string `json:"lock_type"`
}

type Script struct {
	CodeHash string `json:"code_hash"`
	HashType string `json:"hash_type"`
	Args     string `json:"args"`
}
type DeriveMetadata struct {
	Script `json:"script"`
}
