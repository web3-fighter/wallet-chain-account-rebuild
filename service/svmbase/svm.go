package svmbase

import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	errHTTPError       = errors.New("aptos http error")
	errInvalidAddress  = errors.New("invalid address")
	errInvalidResponse = errors.New("invalid response")
)

var _ SVMClient = (*svmClient)(nil)

type svmClient struct {
	client *resty.Client
}

func (c *svmClient) GetBlockByHash(ctx context.Context, signature string) (*BlockResult, error) {
	tx, err := c.GetTransaction(ctx, signature)
	if err != nil {
		return nil, err
	}
	return c.GetBlockBySlot(ctx, tx.Slot, Full)
}

func (c *svmClient) GetHealth(ctx context.Context) (string, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getHealth",
		"params":  []interface{}{},
	}

	response := &GetHealthResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return "", fmt.Errorf("failed to get health: %w", errHTTPError)
	}

	if response.Error != nil {
		if response.Error.Code == -32005 {
			return HealthBehind, nil
		}
		return HealthUnknown, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	if response.Result == "" {
		return HealthUnknown, fmt.Errorf("invalid response: empty result")
	}

	switch response.Result {
	case HealthOk, HealthBehind:
		return response.Result, nil
	default:
		return HealthUnknown, fmt.Errorf("unknown health status: %s", response.Result)
	}
}

func (c *svmClient) GetAccountInfo(ctx context.Context, inputAddr string) (*AccountInfo, error) {
	dealAddr := strings.TrimSpace(inputAddr)
	if dealAddr == "" {
		return nil, fmt.Errorf("invalid input: empty address")
	}
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getAccountInfo",
		"params": []interface{}{
			dealAddr,
			map[string]string{
				"encoding": "base64",
			},
		},
	}
	response := &GetAccountInfoResponse{}
	resp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get account info: %w", errHTTPError)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	accountInfo := &response.Result.Value
	if accountInfo.Owner == "" {
		return nil, fmt.Errorf("invalid response: empty owner")
	}
	if len(accountInfo.Data) < 2 {
		return nil, fmt.Errorf("invalid response: missing data encoding")
	}
	if accountInfo.Data[1] != "base64" {
		return nil, fmt.Errorf("unexpected data encoding: %s", accountInfo.Data[1])
	}

	return accountInfo, nil
}

func (c *svmClient) GetBalance(ctx context.Context, inputAddr string) (uint64, error) {
	dealAddr := strings.TrimSpace(inputAddr)
	if dealAddr == "" {
		return 0, fmt.Errorf("invalid input: empty address")
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params": []interface{}{
			dealAddr,
		},
	}

	response := &GetBalanceResponse{}
	resp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("HTTP error: status=%d, body=%s",
			resp.StatusCode(),
			resp.String(),
		)
	}
	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}
	if response.Result.Value == 0 {
		log.Printf("Warning: account balance is 0 for address: %s", dealAddr)
	}
	return response.Result.Value, nil
}

func (c *svmClient) GetLatestBlockHash(ctx context.Context, commitmentType CommitmentType) (string, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getLatestBlockhash",
		"params": []interface{}{
			map[string]string{
				"commitment": string(commitmentType),
			},
		},
	}

	response := &GetLatestBlockHashResponse{}
	resp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("failed to get latest blockhash: %w", errHTTPError)
	}

	if response.Error != nil {
		return "", fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	blockHash := response.Result.Value.BlockHash
	if blockHash == "" {
		return "", fmt.Errorf("invalid blockhash response: empty blockhash")
	}

	return blockHash, nil
}

// SendTransaction
/*
功能：向 Solana 主网广播交易
入参
	signedTx string：Base58 编码的已签名交易
	config *SendTransactionRequest：发送配置，如是否跳过前检查等

行为
	构造 sendTransaction 的 JSON-RPC 请求
	将交易发送给 RPC 节点进行广播

返回
	成功返回交易哈希（txid）
	失败返回网络错误、RPC 错误、或交易签名为空错误
*/
func (c *svmClient) SendTransaction(ctx context.Context, signedTx string, config *SendTransactionRequest) (string, error) {
	if signedTx == "" {
		return "", fmt.Errorf("invalid input: empty transaction")
	}
	if config == nil {
		config = &SendTransactionRequest{
			Commitment: string(Finalized),
			Encoding:   "base58",
		}
	}
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "sendTransaction",
		"params": []interface{}{
			signedTx,
			config,
		},
	}

	resp := &SendTransactionResponse{}

	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(&resp).
		Post("/")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return "", fmt.Errorf("failed to send transaction: %w", errHTTPError)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("RPC error: code=%d, message=%s",
			resp.Error.Code,
			resp.Error.Message,
		)
	}

	if resp.Result == "" {
		return "", fmt.Errorf("empty transaction signature returned")
	}

	return resp.Result, nil
}

// SimulateTransaction
/*
功能：模拟交易执行，不上链，用于预估结果
	入参
		signedTx string：Base64 编码的已签名交易（注意：和上面不同）
		config *SimulateRequest：模拟配置，如 commitment、是否返回日志等

	行为
		构造 simulateTransaction 的 JSON-RPC 请求
		提交模拟执行，不会广播到链上
	返回
		成功返回模拟执行结果，包括 logs、units_consumed、错误信息
		失败返回模拟失败原因或 RPC 错误

	使用建议
		在发送高价值交易前，建议先调用 SimulateTransaction，确保不会失败或报错。
		模拟成功后再调用 SendTransaction 广播，避免实际失败浪费 gas。
*/
/*
两者区别对比
对比项	SendTransaction	SimulateTransaction
	功能	广播真实交易	模拟交易执行
	编码要求	base58 编码	base64 编码
	是否上链	✅ 是	❌ 否
	是否花费费用	✅ 是（可能消耗 lamports）	❌ 否
		在 Solana 区块链中，lamports 是 SOL 的最小单位，类似于以太坊的 wei、比特币的 satoshi。
		✅ 一、基本概念
			单位	数量	说明
			1 SOL	= 1,000,000,000 lamports	1 SOL = 10⁹ lamports
			lamports	最小单位	不能再拆分
		因此，如果你看到一笔交易消耗了 5000 lamports，这相当于：
			5000 / 1_000_000_000 = 0.000005 SOL
		✅ 二、lamports 常见用途
			交易手续费（Transaction Fee）	每笔交易都会消耗少量 lamports，通常为 5000～10000 lamports（约 0.000005～0.00001 SOL）
			租赁机制（Rent）	Solana 账户占用空间需要支付租金（除非存入足够 lamports 成为 “rent-exempt”）
			创建账户	创建新账户时需预存一定 lamports 保证账户存在
			程序部署	部署合约（Program）时也需支付 lamports 以存储代码
	返回内容	txid 字符串	模拟结果结构体（日志、单元消耗等）
	常见用途	发送真实转账、部署合约等	检查是否成功、调试合约
*/
func (c *svmClient) SimulateTransaction(ctx context.Context, signedTx string, config *SimulateRequest) (*SimulateResult, error) {
	if signedTx == "" {
		return nil, fmt.Errorf("invalid input: empty transaction")
	}
	if config == nil {
		// Solana RPC 内部也倾向于使用 base64 作为默认交易模拟（simulate）返回的编码格式；
		// 如你使用了 simulateTransaction 方法，返回的 data 默认格式就是 base64。
		config = &SimulateRequest{
			Commitment: string(Finalized),
			Encoding:   "base64",
		}
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "simulateTransaction",
		"params": []interface{}{
			signedTx,
			config,
		},
	}

	resp := &SimulateTransactionResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(resp).
		Post("/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to simulate transaction: %w", errHTTPError)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			resp.Error.Code,
			resp.Error.Message,
		)
	}
	if resp.Result.Err != nil {
		return nil, fmt.Errorf("simulation failed: %v", resp.Result.Err)
	}
	if resp.Result.UnitsConsumed == 0 && len(resp.Result.Logs) == 0 {
		return nil, fmt.Errorf("empty simulation result")
	}
	return &resp.Result, nil
}

// GetFeeForMessage 估算一笔 已构造好但尚未签名的交易消息（Message） 的交易费用（Lamports），
// 这是 Solana 中一种轻量级交易费估算方式，不需要实际签名或广播交易。
func (c *svmClient) GetFeeForMessage(ctx context.Context, message string) (uint64, error) {
	if message == "" {
		return 0, fmt.Errorf("invalid input: empty message")
	}
	config := GetFeeForMessageRequest{
		Commitment: string(Finalized),
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getFeeForMessage",
		"params":  []interface{}{message, config},
	}

	resp := &GetFeeForMessageResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(resp).
		Post("/")
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return 0, fmt.Errorf("failed to get fee for message: %w", errHTTPError)
	}
	if resp.Error != nil {
		return 0, fmt.Errorf("RPC error: code=%d, message=%s",
			resp.Error.Code,
			resp.Error.Message,
		)
	}
	if resp.Result.Value == nil {
		return 0, fmt.Errorf("invalid message or unable to estimate fee")
	}

	return *resp.Result.Value, nil
}

// GetRecentPrioritizationFees
/*
	该方法用于从 Solana 节点获取一批区块中不同交易优先级（priority level）对应的实际费用（fee），帮助钱包或交易平台动态评估：
		当前链上拥堵程度
		设置合适的优先级费用
		估算加速交易的额外成本
	可用于 交易费用推荐、交易调速策略、动态加速广播服务。
*/
func (c *svmClient) GetRecentPrioritizationFees(ctx context.Context) ([]*PrioritizationFee, error) {
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getRecentPrioritizationFees",
		"params":  []interface{}{},
	}

	resp := &getRecentPrioritizationFeesResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(resp).
		Post("/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to get prioritization fees: %w", errHTTPError)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			resp.Error.Code,
			resp.Error.Message,
		)
	}
	if resp.Result == nil || len(resp.Result) == 0 {
		return nil, errors.New("invalid response: empty RecentPrioritizationFees data")
	}

	return resp.Result, nil
}

// GetSlot
/*
方法名： GetSlot
目标： 获取 Solana 网络中，某一特定确认等级下的最新 Slot（即区块号）
用途：
	确定链的最新进展高度
	结合 Slot 做区块/交易的时间戳估计
	数据同步、分布式比对、容灾分析
*/
func (c *svmClient) GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error) {
	config := GetSlotRequest{
		// 传不同的 commitment，节点可能返回不同的 slot
		Commitment: commitment,
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSlot",
		"params":  []interface{}{config},
	}

	response := &GetSlotResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")

	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return 0, fmt.Errorf("failed to get slot: %w", errHTTPError)
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	if response.Result == 0 {
		return 0, fmt.Errorf("invalid slot number: got 0")
	}

	return response.Result, nil
}

func (c *svmClient) GetBlocksWithLimit(ctx context.Context, startSlot uint64, limit uint64) ([]uint64, error) {
	if startSlot == 0 {
		return nil, fmt.Errorf("invalid input: start slot cannot be 0")
	}
	if limit == 0 {
		return nil, fmt.Errorf("invalid input: limit cannot be 0")
	}
	if limit > blockLimit {
		return nil, fmt.Errorf("limit must not exceed %d blocks", blockLimit)
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBlocksWithLimit",
		"params":  []uint64{startSlot, limit},
	}

	response := &GetBlocksWithLimitResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to get blocks with limit: %w", errHTTPError)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}

	if response.Result == nil {
		return []uint64{}, nil
	}

	if len(response.Result) == 0 {
		log.Printf("Warning: no blocks found for slot range %d to %d",
			startSlot, startSlot+limit-1)
	}

	if uint64(len(response.Result)) > limit {
		return nil, fmt.Errorf("received more blocks than requested limit: got %d, want <= %d",
			len(response.Result), limit)
	}

	return response.Result, nil
}

// GetBlockBySlot
/*
获取指定 slot 的区块信息，可选是否包含交易详情、区块奖励、编码格式、交易版本支持等。
对应 Solana RPC 方法： getBlock
*/
func (c *svmClient) GetBlockBySlot(ctx context.Context, slot uint64, detailType TransactionDetailsType) (*BlockResult, error) {
	config := GetBlockRequest{
		Commitment:                     Finalized,
		Encoding:                       "json",
		MaxSupportedTransactionVersion: 0,
		TransactionDetails:             string(detailType),
		Rewards:                        false,
	}
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBlock",
		"params":  []interface{}{slot, config},
	}
	resp := &GetBlockResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(resp).
		Post("/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to get block: %w", errHTTPError)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: (code: %d) %s", resp.Error.Code, resp.Error.Message)
	}

	return &resp.Result, nil
}

// GetTransaction 这个方法用于调用 Solana RPC 接口 getTransaction，
// 获取某笔交易的完整信息（包括原始交易内容、执行元信息、签名、时间戳等）。
// signature：交易哈希（base58 编码）
func (c *svmClient) GetTransaction(ctx context.Context, signature string) (*TransactionResult, error) {
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return nil, fmt.Errorf("invalid input: empty signature")
	}
	if len(signature) < 88 || len(signature) > 90 {
		return nil, fmt.Errorf("invalid signature length: expected 88-90 chars, got %d", len(signature))
	}
	/*
		encoding: "json"	string	返回结构为 JSON 格式（还有 base58、base64）
		commitment: Finalized	string	表示查询已达成 Finalized 状态的交易（不可回滚）
		maxSupportedTransactionVersion: 0	int	表示客户端最多只支持 Version 0 的交易（即不支持未来版本，legacy 视为 version "null"）
	*/
	config := map[string]interface{}{
		"encoding":                       "json",
		"commitment":                     Finalized,
		"maxSupportedTransactionVersion": 0,
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params":  []interface{}{signature, config},
	}

	response := &GetTransactionResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(response).
		Post("/")

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to get transaction: %w", errHTTPError)
	}
	if response.Error != nil {
		if response.Error.Code == -32004 {
			return nil, fmt.Errorf("transaction not found: %s", signature)
		}
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			response.Error.Code,
			response.Error.Message,
		)
	}
	if response.Result.Transaction.Signatures == nil {
		return nil, fmt.Errorf("invalid response: empty transaction data")
	}

	return &TransactionResult{
		Slot:        response.Result.Slot,
		Version:     response.Result.Version,
		BlockTime:   response.Result.BlockTime,
		Transaction: response.Result.Transaction,
		Meta:        response.Result.Meta,
	}, nil
}

// GetTransactionRange 批量获取 Solana 交易详情 的函数，名称为 GetTransactionRange。
// 它封装在一个 svmClient 客户端中， 调用前面定义好的单个 GetTransaction 方法，
// 并支持高并发请求、限速、超时控制、错误重试等机制。
func (c *svmClient) GetTransactionRange(ctx context.Context, inputSignatureList []string) ([]*TransactionResult, error) {
	if len(inputSignatureList) == 0 {
		return nil, fmt.Errorf("empty signatures")
	}

	for i, sig := range inputSignatureList {
		inputSignatureList[i] = strings.TrimSpace(sig)
		if inputSignatureList[i] == "" {
			return nil, fmt.Errorf("invalid input: empty signature at index %d", i)
		}
		if len(inputSignatureList[i]) < 88 || len(inputSignatureList[i]) > 90 {
			return nil, fmt.Errorf("invalid signature length at index %d: expected 88-90 chars, got %d",
				i, len(inputSignatureList[i]))
		}
	}

	if len(inputSignatureList) == 1 {
		tx, err := c.GetTransaction(ctx, inputSignatureList[0])
		if err != nil {
			return nil, fmt.Errorf("failed to get single transaction: %w", err)
		}
		return []*TransactionResult{tx}, nil
	}

	const (
		maxConcurrent   = 20
		requestInterval = 100 * time.Millisecond
		timeout         = 5 * time.Minute
		maxRetries      = 3
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChannel := make(chan *TransactionResult, len(inputSignatureList))
	errorChannel := make(chan error, len(inputSignatureList))

	rateLimiter := time.NewTicker(requestInterval)
	defer rateLimiter.Stop()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)

	for i, sig := range inputSignatureList {
		wg.Add(1)
		go func(index int, signature string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			var tx *TransactionResult
			var err error

			for retry := 0; retry < maxRetries; retry++ {
				select {
				case <-ctx.Done():
					errorChannel <- ctx.Err()
					return
				case <-rateLimiter.C:
					tx, err = c.GetTransaction(ctx, signature)
					if err == nil {
						resultChannel <- tx
						return
					}

					if retry < maxRetries-1 && strings.Contains(err.Error(), "request failed") {
						time.Sleep(time.Second * time.Duration(retry+1))
						continue
					}

					errorChannel <- fmt.Errorf("failed to get transaction %s: %w", signature, err)
					return
				}
			}

			if err != nil {
				errorChannel <- fmt.Errorf("max retries exceeded for %s: %w", signature, err)
			}
		}(i, sig)
	}

	wg.Wait()
	close(resultChannel)
	close(errorChannel)

	var errorList []string
	for err := range errorChannel {
		if err != nil {
			errorList = append(errorList, err.Error())
		}
	}
	if len(errorList) > 0 {
		return nil, fmt.Errorf("multiple errors occurred: %s", strings.Join(errorList, "; "))
	}

	validResults := make([]*TransactionResult, 0, len(resultChannel))
	for result := range resultChannel {
		if result != nil && result.Transaction.Signatures != nil {
			validResults = append(validResults, result)
		} else {
			log.Println("Skipping invalid transaction", "result", result)
		}
	}

	if len(validResults) == 0 {
		return nil, fmt.Errorf("no valid transactions found")
	}

	return validResults, nil
}

// GetSignaturesForAddress
/*
	查询某个 Solana 地址的 交易签名记录列表。
	返回的是交易摘要信息（不包含交易详情），可用于后续调用 getTransaction 获取完整交易。
	对接 Solana 的 getSignaturesForAddress JSON-RPC 接口。

参数名	说明
	ctx	上下文，用于控制请求超时、取消
	address	要查询的地址（Base58 编码的账户公钥）
	commitment	区块确认级别，如 finalized, confirmed, processed
	limit	限制返回最多多少条签名（最大 1000）
	beforeSignature	向前分页：从这个签名之前开始查找（不包含该签名）
	untilSignature	向后分页：查询到这个签名就停止（包含该签名）
*/
/*
在 Solana 区块链中，交易签名（signature）就是该交易的唯一标识，也可视为该交易的 Hash。
✅ Solana 中的交易签名（Signature）详解：
	交易签名（signature）	是对交易数据进行 Ed25519 签名后的 Base58 编码字符串，长度通常为 88～90 字符。
	作用	是交易的唯一标识，可以用来查交易详情、追踪状态、分页定位等
	与 Ethereum 的 tx hash 类比	在作用上等同于以太坊的 transaction hash，但生成方式不同（Solana 是签名而不是哈希）
	可用于 RPC 查询	例如 getTransaction, getConfirmedTransaction, getSignaturesForAddress 等接口都用它作为索引
🧠 举例：
	{
	  "signature": "3htd98zMre...LZJyyud54WJTP",
	  ...
	}
你可以拿这个 signature 去调用：
	curl https://api.mainnet-beta.solana.com -X POST \
	  -H "Content-Type: application/json" \
	  -d '{
		"jsonrpc":"2.0",
		"id":1,
		"method":"getTransaction",
		"params":["3htd98zMre...LZJyyud54WJTP", {"encoding": "json"}]
	  }'
	即可获得该交易的详细信息。

🧩 补充：为什么不叫 hash？
	在以太坊，交易是用 keccak256(rlp(transaction)) 哈希生成的哈希值来唯一标识；
	在 Solana，交易是通过 第一个签名者对交易数据签名（使用 Ed25519），并将该签名作为交易的标识；
	所以它不是一个纯粹的哈希，而是签名后的结果（但同样是唯一且可验证的）。
*/
func (c *svmClient) GetSignaturesForAddress(ctx context.Context, address string, commitment CommitmentType, limit uint64, beforeSignature string, untilSignature string) ([]*SignatureInfo, error) {
	config := &GetSignaturesRequest{
		Commitment: string(commitment),
		Limit:      limit,
		Before:     beforeSignature,
		Until:      untilSignature,
	}

	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSignaturesForAddress",
		"params":  []interface{}{address, config},
	}

	resp := &GetSignaturesResponse{}
	httpResp, err := c.client.R().SetContext(ctx).
		SetBody(requestBody).
		SetResult(resp).
		Post("/")
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("failed to get signatures: %w", errHTTPError)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: code=%d, message=%s",
			resp.Error.Code,
			resp.Error.Message,
		)
	}

	if resp.Result == nil {
		return nil, errors.New("invalid response: empty signatures data")
	}

	return resp.Result, nil
}

func NewSVMClient(client *resty.Client) SVMClient {
	return &svmClient{client: client}
}

//func NewSVMHttpClient(baseUrl string) (SVMClient, error) {
//	return NewSVMHttpClientAll(baseUrl, defaultWithDebug)
//}
//
//func NewSVMHttpClientAll(baseUrl string, withDebug bool) (SVMClient, error) {
//	grestyClient := resty.New()
//	grestyClient.SetBaseURL(baseUrl)
//	grestyClient.SetTimeout(defaultRequestTimeout)
//	grestyClient.SetRetryCount(defaultRetryCount)
//	grestyClient.SetRetryWaitTime(defaultRetryWaitTime)
//	grestyClient.SetRetryMaxWaitTime(defaultRetryMaxWaitTime)
//	grestyClient.SetDebug(withDebug)
//
//	// Retry Condition
//	//grestyClient.AddRetryCondition(func(r *gresty.Response, err error) bool {
//	//	return err != nil || r.StatusCode() >= 500
//	//})
//
//	grestyClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
//		log.Printf("Making request to %s (Attempt %d)", r.URL, r.Attempt)
//		return nil
//	})
//
//	grestyClient.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
//		statusCode := r.StatusCode()
//		attempt := r.Request.Attempt
//		method := r.Request.Method
//		url := r.Request.URL
//		log.Printf("Response received: Method=%s, URL=%s, Status=%d, Attempt=%d",
//			method, url, statusCode, attempt)
//
//		if statusCode >= 400 {
//			if statusCode == 404 {
//				return fmt.Errorf("%d resource not found %s %s: %w",
//					statusCode, method, url, errHTTPError)
//			}
//			if statusCode >= 500 {
//				return fmt.Errorf("%d server error %s %s: %w",
//					statusCode, method, url, errHTTPError)
//			}
//			return fmt.Errorf("%d cannot %s %s: %w",
//				statusCode, method, url, errHTTPError)
//		}
//		return nil
//	})
//	return NewSVMClient(grestyClient), nil
//}
