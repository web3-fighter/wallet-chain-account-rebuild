package ethereum

type Eip1559DynamicFeeTx struct {
	ChainId     string `json:"chain_id"`
	Nonce       uint64 `json:"nonce"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	GasLimit    uint64 `json:"gas_limit"`
	Gas         uint64 `json:"Gas"`

	MaxFeePerGas         string `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`

	// eth/erc20 amount
	Amount string `json:"amount"`
	// erc20 erc721 erc1155 contract_address
	ContractAddress string `json:"contract_address"`
}

type Eip1559TransactionInfo struct {
	Hash                 string `json:"hash"`
	FromAddress          string `json:"from_address"`
	ToAddress            string `json:"to_address"`
	Value                string `json:"value"`
	GasLimit             uint64 `json:"gas_limit"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`

	Nonce   uint64 `json:"nonce"`
	Data    string `json:"data"`
	Type    uint8  `json:"type"`
	ChainId string `json:"chain_id"`

	// eth/erc20 amount
	Amount string `json:"amount"`
	// erc20 erc721 erc1155 contract_address
	ContractAddress string `json:"contract_address"`
}
