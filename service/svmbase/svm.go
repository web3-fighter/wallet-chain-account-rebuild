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

// SendTransaction TODO
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

// SimulateTransaction TODO
func (c *svmClient) SimulateTransaction(ctx context.Context, signedTx string, config *SimulateRequest) (*SimulateResult, error) {
	if signedTx == "" {
		return nil, fmt.Errorf("invalid input: empty transaction")
	}
	if config == nil {
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

// GetFeeForMessage TODO
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

// GetRecentPrioritizationFees TODO
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

func (c *svmClient) GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error) {
	config := GetSlotRequest{
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

func (c *svmClient) GetTransaction(ctx context.Context, signature string) (*TransactionResult, error) {
	signature = strings.TrimSpace(signature)
	if signature == "" {
		return nil, fmt.Errorf("invalid input: empty signature")
	}
	if len(signature) < 88 || len(signature) > 90 {
		return nil, fmt.Errorf("invalid signature length: expected 88-90 chars, got %d", len(signature))
	}
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

func (c *svmClient) GetTxForAddress(ctx context.Context, address string, commitment CommitmentType, limit uint64, beforeSignature string, untilSignature string) ([]*SignatureInfo, error) {
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
