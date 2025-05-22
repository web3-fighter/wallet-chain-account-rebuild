package evmbase

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

type EVMClient interface {
	BlockHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	BlockHeadersByRange(ctx context.Context, startHeight, endHeight *big.Int, chainId uint) ([]types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*RpcBlock, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*RpcBlock, error)
	LatestSafeBlockHeader(ctx context.Context) (*types.Header, error)
	LatestFinalizedBlockHeader(ctx context.Context) (*types.Header, error)
	TxCountByAddress(ctx context.Context, address common.Address) (hexutil.Uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	SendRawTransaction(ctx context.Context, rawTx string) (*common.Hash, error)
	TxByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error)
	TxReceiptByHash(ctx context.Context, hash common.Hash) (*types.Receipt, error)
	GetStorageHash(ctx context.Context, address common.Address, blockNumber *big.Int) (common.Hash, error)
	EthGetCode(ctx context.Context, address common.Address) (string, error)
	GetBalance(ctx context.Context, address common.Address) (*big.Int, error)
	FilterLogs(ctx context.Context, filterQuery ethereum.FilterQuery, chainId uint) (Logs, error)
	Close(ctx context.Context)
}

type Logs struct {
	Logs          []types.Log
	ToBlockHeader *types.Header
}

type RPC interface {
	Close()
	CallContext(ctx context.Context, result any, method string, args ...any) error
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
}

type TransactionList struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Hash  string `json:"hash"`
	Value string `json:"value"`
}

type RpcBlock struct {
	Hash         common.Hash       `json:"hash"`
	Number       string            `json:"number"`
	Transactions []TransactionList `json:"transactions"`
	BaseFee      string            `json:"baseFeePerGas"`
}

func (b *RpcBlock) NumberUint64() (uint64, error) {
	return hexutil.DecodeUint64(b.Number)
}

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

	Signature string `json:"signature,omitempty"`
}
