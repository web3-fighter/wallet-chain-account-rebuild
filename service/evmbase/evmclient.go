package evmbase

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/pkg/helpers"
	"github.com/web3-fighter/wallet-chain-account/pkg/retry"
	"math/big"
	"sync"
	"time"
)

const (
	defaultDialTimeout    = 5 * time.Second
	defaultDialAttempts   = 5
	defaultRequestTimeout = 10 * time.Second
)

var _ EVMClient = (*evmClient)(nil)

type evmClient struct {
	evmRpc RPC
}

func (c *evmClient) BlockHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	tCtx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var header *types.Header
	err := c.evmRpc.CallContext(tCtx, &header, "eth_getBlockByNumber", toBlockNumArg(number), false)
	if err != nil {
		log.Error("Call eth_getBlockByNumber method fail", "err", err)
		return nil, err
	} else if header == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}

	return header, nil
}

func (c *evmClient) BlockHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var header *types.Header
	err := c.evmRpc.CallContext(ctxwt, &header, "eth_getBlockByHash", hash, false)
	if err != nil {
		return nil, err
	} else if header == nil {
		return nil, ethereum.NotFound
	}

	if header.Hash() != hash {
		return nil, errors.New("header mismatch")
	}

	return header, nil
}

// BlockHeadersByRange 兼容多链的 批量获取区块头的 RPC 调用实现
func (c *evmClient) BlockHeadersByRange(ctx context.Context, startHeight, endHeight *big.Int, chainId uint) ([]types.Header, error) {
	if startHeight.Cmp(endHeight) == 0 {
		header, err := c.BlockHeaderByNumber(ctx, startHeight)
		if err != nil {
			return nil, err
		}
		return []types.Header{*header}, nil
	}

	count := new(big.Int).Sub(endHeight, startHeight).Uint64() + 1
	headers := make([]types.Header, count)
	batchElems := make([]rpc.BatchElem, count)
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	// ZkFair 相关链
	if chainId == uint(domain.ZkFairSepoliaChainId) ||
		chainId == uint(domain.ZkFairChainId) {
		groupSize := 100
		var wg sync.WaitGroup
		numGroups := (int(count)-1)/groupSize + 1
		wg.Add(numGroups)
		// 每 groupSize 个 block 启动一个 goroutine 查询
		for i := 0; i < int(count); i += groupSize {
			start := i
			end := i + groupSize - 1
			if end > int(count) {
				end = int(count) - 1
			}
			go func(start, end int) {
				defer wg.Done()
				for j := start; j <= end; j++ {
					height := new(big.Int).Add(startHeight, new(big.Int).SetUint64(uint64(j)))
					//对 batchElems 的并发写入是安全的，因为每个 goroutine 只写入自己负责的区间（不重叠）；
					batchElems[j] = rpc.BatchElem{
						Method: "eth_getBlockByNumber",
						Result: new(types.Header),
						Error:  nil,
					}
					header := new(types.Header)
					batchElems[j].Error = c.evmRpc.CallContext(ctxwt, header, batchElems[j].Method, toBlockNumArg(height), false)
					batchElems[j].Result = header
				}
			}(start, end)
		}
		wg.Wait()
	} else {
		for i := uint64(0); i < count; i++ {
			height := new(big.Int).Add(startHeight, new(big.Int).SetUint64(i))
			batchElems[i] = rpc.BatchElem{Method: "eth_getBlockByNumber", Args: []interface{}{toBlockNumArg(height), false}, Result: &headers[i]}
		}
		// 普通链：使用以太坊 RPC 的 批量调用（batch）机制 请求所有区块头
		err := c.evmRpc.BatchCallContext(ctxwt, batchElems)
		if err != nil {
			return nil, err
		}
	}
	size := 0
	for i, batchElem := range batchElems {
		header, ok := batchElem.Result.(*types.Header)
		if !ok {
			return nil, fmt.Errorf("unable to transform rpc response %v into types.Header", batchElem.Result)
		}
		headers[i] = *header
		size = size + 1
	}
	headers = headers[:size]

	return headers, nil
}

func (c *evmClient) BlockByNumber(ctx context.Context, number *big.Int) (*RpcBlock, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var block *RpcBlock
	err := c.evmRpc.CallContext(ctxwt, &block, "eth_getBlockByNumber", toBlockNumArg(number), true)
	if err != nil {
		log.Error("Call eth_getBlockByNumber method fail", "err", err)
		return nil, err
	} else if block == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	return block, nil
}

func (c *evmClient) BlockByHash(ctx context.Context, hash common.Hash) (*RpcBlock, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var block *RpcBlock
	err := c.evmRpc.CallContext(ctxwt, &block, "eth_getBlockByHash", hash, true)
	if err != nil {
		log.Error("Call eth_getBlockByHash method fail", "err", err)
		return nil, err
	} else if block == nil {
		log.Warn("header not found")
		return nil, ethereum.NotFound
	}
	return block, nil
}

// LatestSafeBlockHeader 获取最新的 “安全区块头”（Safe Block Header）
// "safe" 区块：指链上经过一定确认数的区块，可能还没最终确定，但大概率不会被回滚。
func (c *evmClient) LatestSafeBlockHeader(ctx context.Context) (*types.Header, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var header *types.Header
	// eth_getBlockByNumber 方法支持传 "safe" 和 "finalized" 作为区块编号参数。
	err := c.evmRpc.CallContext(ctxwt, &header, "eth_getBlockByNumber", "safe", false)
	if err != nil {
		return nil, err
	} else if header == nil {
		return nil, ethereum.NotFound
	}

	return header, nil
}

// LatestFinalizedBlockHeader 获取最新的 “最终确定区块头”（Finalized Block Header）
// "finalized" 区块：已经最终确定的区块，不可能被回滚，安全性最高。
func (c *evmClient) LatestFinalizedBlockHeader(ctx context.Context) (*types.Header, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var header *types.Header
	err := c.evmRpc.CallContext(ctxwt, &header, "eth_getBlockByNumber", "finalized", false)
	if err != nil {
		return nil, err
	} else if header == nil {
		return nil, ethereum.NotFound
	}

	return header, nil
}

func (c *evmClient) TxCountByAddress(ctx context.Context, address common.Address) (hexutil.Uint64, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var nonce hexutil.Uint64
	err := c.evmRpc.CallContext(ctxwt, &nonce, "eth_getTransactionCount", address, "latest")
	if err != nil {
		log.Error("Call eth_getTransactionCount method fail", "err", err)
		return 0, err
	}
	log.Info("get nonce by address success", "nonce", nonce)
	return nonce, err
}

// SuggestGasPrice 获取当前网络推荐的 传统 Gas 单价（单位是 Wei），主要用于 非 EIP-1559（legacy）交易 的定价
// 交易发送者直接设置 gasPrice，全额支付给矿工。
func (c *evmClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var hex hexutil.Big
	if err := c.evmRpc.CallContext(ctxwt, &hex, "eth_gasPrice"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

// SuggestGasTipCap 调用以太坊节点的 eth_maxPriorityFeePerGas 接口，获取当前建议的矿工小费（Tip），以支持 EIP-1559 格式的交易定价机制。
/*
对比项	SuggestGasTipCap	SuggestGasPrice
使用的 RPC 方法	eth_maxPriorityFeePerGas	eth_gasPrice
用途	EIP-1559 交易的 小费（tip） 建议	传统（legacy）交易的 总 gas 单价
推荐给矿工的内容	只是 Tip（矿工小费）	传统 Gas 总价，矿工全部收入
是否包含 base fee	❌ 不包含，需要你自己加上 base fee 才能得到 maxFeePerGas	✅ 已包含全部费用（因为 legacy 没有 base fee）
适用交易类型	EIP-1559 类型交易（有 maxFeePerGas, maxPriorityFeePerGas 字段）	legacy 交易（只有 gasPrice 字段）

{
  "maxFeePerGas": "30000000000",          // 最多愿意出 30 gwei
  "maxPriorityFeePerGas": "2000000000"    // 给矿工的小费 2 gwei
}
maxFeePerGas = baseFee + priorityFee，新交易（EIP-1559）需要你再手动加上 base fee 才能计算 maxFeePerGas
base fee 会烧掉，priority fee 给矿工。
*/
func (c *evmClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	var hex hexutil.Big
	if err := c.evmRpc.CallContext(ctxwt, &hex, "eth_maxPriorityFeePerGas"); err != nil {
		return nil, err
	}
	return (*big.Int)(&hex), nil
}

func (c *evmClient) SendRawTransaction(ctx context.Context, rawTx string) (*common.Hash, error) {
	var txHash common.Hash
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()
	if err := c.evmRpc.CallContext(ctxwt, &txHash, "eth_sendRawTransaction", rawTx); err != nil {
		return nil, err
	}
	log.Info("send tx to ethereum success", "txHash", txHash.Hex())
	return &txHash, nil
}

// TxByHash 根据交易哈希 `hash` 查询一笔 **交易的详细信息（但不包含执行结果）**。
// 这是交易被打包到链上的原始交易内容，不包含执行状态、消耗的 gas、日志等。
func (c *evmClient) TxByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var tx *types.Transaction
	err := c.evmRpc.CallContext(ctxwt, &tx, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, err
	} else if tx == nil {
		return nil, ethereum.NotFound
	}

	return tx, nil
}

// TxReceiptByHash 根据交易哈希 `hash` 查询这笔交易的 **执行结果、状态、事件日志等信息**。
func (c *evmClient) TxReceiptByHash(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var txReceipt *types.Receipt
	err := c.evmRpc.CallContext(ctxwt, &txReceipt, "eth_getTransactionReceipt", hash)
	if err != nil {
		return nil, err
	} else if txReceipt == nil {
		return nil, ethereum.NotFound
	}

	return txReceipt, nil
}

// GetStorageHash 调用以太坊的 eth_getProof 接口，获取指定合约地址在某个区块的 storageHash（即存储槽 Merkle 树的根哈希）。
func (c *evmClient) GetStorageHash(ctx context.Context, address common.Address, blockNumber *big.Int) (common.Hash, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	proof := struct{ StorageHash common.Hash }{}
	err := c.evmRpc.CallContext(ctxwt, &proof, "eth_getProof", address, nil, toBlockNumArg(blockNumber))
	if err != nil {
		return common.Hash{}, err
	}

	return proof.StorageHash, nil
}

// EthGetCode 判断一个以太坊地址是 EOA（Externally Owned Account，外部账户） 还是 合约账户（Contract Account） 的工具方法。
// 如果返回为空（0x），说明是 EOA；否则是合约。
func (c *evmClient) EthGetCode(ctx context.Context, address common.Address) (string, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var result hexutil.Bytes
	err := c.evmRpc.CallContext(ctxwt, &result, "eth_getCode", address, "latest")
	if err != nil {
		return "", err
	}
	// 如果是普通地址（EOA，Externally Owned Account）：返回空字符串（0x）；
	if result.String() == "0x" {
		return "eoa", nil
	} else {
		// 如果是合约地址：装成了 "contract" 返回
		return "contract", nil
	}
}

// GetBalance 使用以太坊 JSON-RPC 接口 eth_getBalance 查询指定地址在最新区块上的余额，并返回一个 *big.Int 类型的余额值（单位为 Wei）。
func (c *evmClient) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	var result hexutil.Big
	err := c.evmRpc.CallContext(ctxwt, &result, "eth_getBalance", address, "latest")
	if err != nil {
		return nil, fmt.Errorf("get balance failed: %w", err)
	}

	balance := (*big.Int)(&result)
	return balance, nil
}

// FilterLogs 根据指定的过滤条件（如区块范围、合约地址、事件等）查询日志信息，并附带日志所在的 ToBlock 的区块头，适配了evm系列不同链（如 ZkFair）的调用方式。
func (c *evmClient) FilterLogs(ctx context.Context, filterQuery ethereum.FilterQuery, chainId uint) (Logs, error) {
	arg, err := toFilterArg(filterQuery)
	if err != nil {
		return Logs{}, err
	}

	var logs []types.Log
	var header types.Header

	batchElems := make([]rpc.BatchElem, 2)
	batchElems[0] = rpc.BatchElem{Method: "eth_getBlockByNumber", Args: []interface{}{toBlockNumArg(filterQuery.ToBlock), false}, Result: &header}
	batchElems[1] = rpc.BatchElem{Method: "eth_getLogs", Args: []interface{}{arg}, Result: &logs}
	ctxwt, cancel := context.WithTimeout(ctx, defaultRequestTimeout*10)
	defer cancel()
	if chainId == uint(domain.ZkFairSepoliaChainId) ||
		chainId == uint(domain.ZkFairChainId) {

		batchElems[0].Error = c.evmRpc.CallContext(ctxwt, &header, batchElems[0].Method, toBlockNumArg(filterQuery.ToBlock), false)
		batchElems[1].Error = c.evmRpc.CallContext(ctxwt, &logs, batchElems[1].Method, arg)
	} else {
		err = c.evmRpc.BatchCallContext(ctxwt, batchElems)
		if err != nil {
			return Logs{}, err
		}
	}
	if batchElems[0].Error != nil {
		return Logs{}, fmt.Errorf("unable to query for the `FilterQuery#ToBlock` header: %w", batchElems[0].Error)
	}
	if batchElems[1].Error != nil {
		return Logs{}, fmt.Errorf("unable to query logs: %w", batchElems[1].Error)
	}
	return Logs{Logs: logs, ToBlockHeader: &header}, nil
}

func (c *evmClient) Close(_ context.Context) {
	c.evmRpc.Close()
}

func DialEthClient(ctx context.Context, rpcUrl string) (EVMClient, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()

	bOff := retry.Exponential()
	rpcCli, err := retry.Do(ctx, defaultDialAttempts, bOff, func() (*rpc.Client, error) {
		if !helpers.IsURLAvailable(rpcUrl) {
			return nil, fmt.Errorf("address unavailable (%s)", rpcUrl)
		}

		client, err := rpc.DialContext(ctx, rpcUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to dial address (%s): %w", rpcUrl, err)
		}

		return client, nil
	})

	if err != nil {
		return nil, err
	}

	return &evmClient{evmRpc: NewRPC(rpcCli)}, nil
}
