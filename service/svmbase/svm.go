package svmbase

import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
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
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetBalance(ctx context.Context, inputAddr string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetLatestBlockhash(ctx context.Context, commitmentType CommitmentType) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) SendTransaction(ctx context.Context, signedTx string, config *SendTransactionRequest) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) SimulateTransaction(ctx context.Context, signedTx string, config *SimulateRequest) (*SimulateResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetFeeForMessage(ctx context.Context, message string) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetRecentPrioritizationFees(ctx context.Context) ([]*PrioritizationFee, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetBlocksWithLimit(ctx context.Context, startSlot uint64, limit uint64) ([]uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetBlockBySlot(ctx context.Context, slot uint64, detailType TransactionDetailsType) (*BlockResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetTransaction(ctx context.Context, signature string) (*TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetTransactionRange(ctx context.Context, signatures []string) ([]*TransactionResult, error) {
	//TODO implement me
	panic("implement me")
}

func (c *svmClient) GetTxForAddress(ctx context.Context, address string, commitment CommitmentType, limit uint64, beforeSignature string, untilSignature string) ([]*SignatureInfo, error) {
	//TODO implement me
	panic("implement me")
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
