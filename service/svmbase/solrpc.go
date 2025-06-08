package svmbase

import (
	"context"
	"errors"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"io"
	"net/http"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrNotConfirmed = errors.New("not confirmed")
)

var _ JSONRpc = (*Client)(nil)

type Client struct {
	rpcURL    string
	rpcClient jsonrpc.RPCClient
}

// Close closes the client.
func (cl *Client) Close() error {
	if cl.rpcClient == nil {
		return nil
	}
	if c, ok := cl.rpcClient.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// CallForInto allows to access the raw RPC client and send custom requests.
func (cl *Client) CallForInto(ctx context.Context, out interface{}, method string, params []interface{}) error {
	return cl.rpcClient.CallForInto(ctx, out, method, params)
}

func (cl *Client) CallWithCallback(
	ctx context.Context,
	method string,
	params []interface{},
	callback func(*http.Request, *http.Response) error,
) error {
	return cl.rpcClient.CallWithCallback(ctx, method, params, callback)
}

func (cl *Client) CallBatch(
	ctx context.Context,
	requests jsonrpc.RPCRequests,
) (jsonrpc.RPCResponses, error) {
	return cl.rpcClient.CallBatch(ctx, requests)
}

func NewBoolean(b bool) *bool {
	return &b
}

func NewTransactionVersion(v uint64) *uint64 {
	return &v
}

// NewSOLRPCClient creates a new Solana RPC client
// with the provided RPC client.
func NewSOLRPCClient(rpcClient jsonrpc.RPCClient) *Client {
	return &Client{
		rpcClient: rpcClient,
	}
}

//// New creates a new Solana JSON RPC client.
//// Client is safe for concurrent use by multiple goroutines.
//func New(rpcEndpoint string) *Client {
//	opts := &jsonrpc.RPCClientOpts{
//		HTTPClient: newHTTP(),
//	}
//
//	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
//	return NewWithCustomRPCClient(rpcClient)
//}
//
//
//// New creates a new Solana JSON RPC client with the provided custom headers.
//// The provided headers will be added to each RPC request sent via this RPC client.
//func NewWithHeaders(rpcEndpoint string, headers map[string]string) *Client {
//	opts := &jsonrpc.RPCClientOpts{
//		HTTPClient:    newHTTP(),
//		CustomHeaders: headers,
//	}
//	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
//	return NewWithCustomRPCClient(rpcClient)
//}

//var (
//	defaultMaxIdleConnsPerHost = 9
//	defaultTimeout             = 5 * time.Minute
//	defaultKeepAlive           = 180 * time.Second
//)
//
//func newHTTPTransport() *http.Transport {
//	return &http.Transport{
//		IdleConnTimeout:     defaultTimeout,
//		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
//		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
//		Proxy:               http.ProxyFromEnvironment,
//		DialContext: (&net.Dialer{
//			Timeout:   defaultTimeout,
//			KeepAlive: defaultKeepAlive,
//			DualStack: true,
//		}).DialContext,
//		ForceAttemptHTTP2: true,
//		// MaxIdleConns:          100,
//		TLSHandshakeTimeout: 10 * time.Second,
//		// ExpectContinueTimeout: 1 * time.Second,
//	}
//}
//
//// newHTTP returns a new Client from the provided config.
//// Client is safe for concurrent use by multiple goroutines.
//func newHTTP() *http.Client {
//	tr := newHTTPTransport()
//
//	return &http.Client{
//		Timeout:   defaultTimeout,
//		Transport: gzhttp.Transport(tr),
//	}
//}
