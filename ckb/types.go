package ckb

const (
	InputOpType              = "INPUT"
	OutputOpType             = "OUTPUT"
	RewardOpType             = "Reward"
	BaseTxSize               = 68 // empty cellDeps + empty headerDeps + empty inputs + empty outputs + empty outputs_data + empty witnesses + version
	InputSize                = 44
	HeaderDepSize            = 32
	CellDepSize              = 37
	SerializedOffsetByteSize = 4
	Secp256k1Tx              = "Secp256k1Tx"
)

type PreprocessMetadata struct {
	TxType string `json:"tx_type"`
}

type PreprocessOptions struct {
	TxType                 string   `json:"tx_type"`
	EstimatedTxSize        uint64   `json:"estimated_tx_size"`
	SuggestedFeeMultiplier *float64 `json:"suggested_fee_multiplier"`
}

type ConstructionMetadata struct {
	TxType string `json:"tx_type"`
}

type OperationMetadata struct {
	Data string `json:"data"`
}
