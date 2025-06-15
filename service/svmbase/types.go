package svmbase

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"time"
)

type SVMClient interface {
	GetHealth(ctx context.Context) (string, error)

	GetAccountInfo(ctx context.Context, inputAddr string) (*AccountInfo, error)
	GetBalance(ctx context.Context, inputAddr string) (uint64, error)
	GetLatestBlockHash(ctx context.Context, commitmentType CommitmentType) (string, error)
	SendTransaction(ctx context.Context, signedTx string, config *SendTransactionRequest) (string, error)
	SimulateTransaction(ctx context.Context, signedTx string, config *SimulateRequest) (*SimulateResult, error)

	GetFeeForMessage(ctx context.Context, message string) (uint64, error)
	GetRecentPrioritizationFees(ctx context.Context) ([]*PrioritizationFee, error)

	GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error)
	GetBlocksWithLimit(ctx context.Context, startSlot uint64, limit uint64) ([]uint64, error)
	GetBlockBySlot(ctx context.Context, slot uint64, detailType TransactionDetailsType) (*BlockResult, error)
	GetBlockByHash(ctx context.Context, signature string) (*BlockResult, error)
	GetTransaction(ctx context.Context, signature string) (*TransactionResult, error)
	GetTransactionRange(ctx context.Context, signatures []string) ([]*TransactionResult, error)
	GetSignaturesForAddress(
		ctx context.Context,
		address string,
		commitment CommitmentType,
		limit uint64,
		beforeSignature string,
		untilSignature string,
	) ([]*SignatureInfo, error)
}

/*
GetSignaturesRequest

		Commitment	区块确认级别，如 finalized, confirmed, processed
	    MinContextSlot  参数主要用于防止客户端读取“过旧的状态” 若节点尚未处理到 MinContextSlot，RPC 会拒绝请求，返回错误；
		Limit	限制返回最多多少条签名（最大 1000）
		Before	向前分页：从这个签名之前开始查找（不包含该签名）
		Until	向后分页：查询到这个签名就停止（包含该签名）
*/
type GetSignaturesRequest struct {
	Commitment     string `json:"commitment,omitempty"`
	MinContextSlot uint64 `json:"minContextSlot,omitempty"`
	Limit          uint64 `json:"limit,omitempty"`
	Before         string `json:"before,omitempty"`
	Until          string `json:"until,omitempty"`
}

type GetSignaturesResponse struct {
	Jsonrpc string           `json:"jsonrpc"`
	ID      int              `json:"id"`
	Error   *RPCError        `json:"error,omitempty"`
	Result  []*SignatureInfo `json:"result"`
}

type GetBlockResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Error   *RPCError   `json:"error,omitempty"`
	Result  BlockResult `json:"result"`
}

/*
GetBlockRequest

	Commitment返回的确认等级，如 "finalized"、"confirmed"、"processed"
	Encoding区块内交易数据的编码方式，常见有 "json"、"jsonParsed"、"base58"、"base64"
	MaxSupportedTransactionVersion	最大支持的交易版本（0 表示 legacy，即仅支持旧版本交易）
		Solana 早期只有 未版本化交易（Legacy Transaction），后来为了增强灵活性，引入了版本化交易（Versioned Transaction），目前主网支持的交易版本包括：
			版本号	名称	简述
			0	Legacy Transaction	旧格式（无版本字段）
			0	Versioned Transaction v0	新格式（有版本字段，版本为 0）
			该字段告诉 RPC 服务端：
				「（客户端）最多能解析到哪一个版本的交易，请不要返回更高版本的交易。」
				👇 取值示例：
				值	意义
				0	只返回 Legacy 交易 或 Versioned v0（取决于 encoding 和实际数据）
				1	支持到 Versioned Transaction v1（目前未在主网上使用）

null（省略）	支持所有版本（一般用于追踪未来兼容性）

	TransactionDetails	指定返回哪些交易信息："full"、"accounts"、"signatures"、"none"
	Rewards	是否返回当前区块的奖励信息（如出块奖励）
*/
type GetBlockRequest struct {
	// slot status
	// Finalized Confirmed Processed
	Commitment CommitmentType `json:"commitment,omitempty"`
	// "json", "jsonParsed", "base58", "base64"
	Encoding string `json:"encoding"`
	// max version
	// Legacy = 0, no other version
	MaxSupportedTransactionVersion int `json:"maxSupportedTransactionVersion"`
	// "full", "accounts", "signatures", "none"
	TransactionDetails string `json:"transactionDetails"`
	// contain rewards
	Rewards bool `json:"rewards"`
}

type GetTransactionResponse struct {
	Jsonrpc string            `json:"jsonrpc"`
	ID      int               `json:"id"`
	Error   *RPCError         `json:"error,omitempty"`
	Result  TransactionResult `json:"result"`
}

// GetBlocksWithLimitResponse represents the response structure
type GetBlocksWithLimitResponse struct {
	JsonRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	Result  []uint64  `json:"result"`
}

type GetSlotRequest struct {
	Commitment CommitmentType `json:"commitment,omitempty"`
	// 参数主要用于防止客户端读取“过旧的状态” 若节点尚未处理到 MinContextSlot，RPC 会拒绝请求，返回错误；
	MinContextSlot uint64 `json:"minContextSlot,omitempty"`
}

type GetSlotResponse struct {
	JsonRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	// slot
	Result uint64 `json:"result"`
}

type getRecentPrioritizationFeesResponse struct {
	Jsonrpc string               `json:"jsonrpc"`
	ID      int                  `json:"id"`
	Error   *RPCError            `json:"error,omitempty"`
	Result  []*PrioritizationFee `json:"result"`
}

/*
GetFeeForMessageRequest

✅ 一、最小单位对比（1 主币 = 多少最小单位）

	主币名称	主币单位	最小单位	数量（主币 : 最小单位）
	比特币	BTC	satoshi	1 BTC = 100,000,000 sat
	以太坊	ETH	wei	1 ETH = 1,000,000,000,000,000,000 wei (1e18)
	Solana	SOL	lamports	1 SOL = 1,000,000,000 lamports (1e9)

✅ 二、设计理念差异

	项目	设计理念
	BTC	交易频率低，以安全性为核心，satoshi 是比特币的最小支付单位，主要用于 UTXO 精确找零。
	ETH	支持智能合约、DeFi，设计高精度 wei 是为应对 Gas 计算和高频交易。
	SOL	高并发链设计，lamports 精度够用，又不会像 wei 那样太细，兼顾性能和易用性。

✅ 三、最小单位是否可再拆分？

	单位	是否可拆分？	精度用途说明
	satoshi	❌ 不能再拆	保证全局 UTXO 一致性
	wei	❌ 不能再拆	用于精细计算 Gas、DeFi 金额等
	lamports	❌ 不能再拆	简化内存和计算需求，链上表现更高效

✅ 四、实际交易精度对比（举例）

	交易	          比特币	          以太坊	          Solana
	交易手续费	通常 1k～5k sat	几万到几百万 wei	通常 5000～10000 lamports
	小额支付	      最少 1 sat	     最少 1 wei	      最少 1 lamport

✅ 五、在钱包或系统开发中的应用建议

	区块链	 建议展示	     建议存储
	BTC	   BTC（带8位小数）   satoshi（整数）
	ETH	   ETH（带18位小数）	wei（整数）
	SOL	   SOL（带9位小数）	lamports（整数）

✅ 总结一句话

	satoshi、wei 和 lamports 都是区块链的最小支付单位，分别服务于 BTC、ETH、SOL 的不同设计理念：BTC 追求安全与稳定，ETH 注重合约与精度，Solana 强调高性能与吞吐。
*/
type GetFeeForMessageRequest struct {
	// Commitment	指定读取链上状态的确认级别，如 processed、confirmed、finalized，值越高越稳定但越慢。
	Commitment string `json:"commitment,omitempty"`
	// MinContextSlot	可选，用于指定估算时的最小 slot，防止节点回滚或使用旧状态。一般不需要设。
	MinContextSlot uint64 `json:"minContextSlot,omitempty"`
}

type GetFeeForMessageResponse struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	Result  struct {
		Context struct {
			Slot uint64 `json:"slot"`
		} `json:"context"`
		Value *uint64 `json:"value"`
	} `json:"result"`
}

type SimulateTransactionResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Error   *RPCError      `json:"error,omitempty"`
	Result  SimulateResult `json:"result"`
}

type SendTransactionResponse struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  string    `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
}

type GetLatestBlockHashResponse struct {
	JsonRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	Result  struct {
		Context struct {
			Slot uint64 `json:"slot"`
		} `json:"context"`
		Value struct {
			BlockHash            string `json:"blockhash"`
			LastValidBlockHeight uint64 `json:"lastValidBlockHeight"`
		} `json:"value"`
	} `json:"result"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GetAccountInfoResponse struct {
	JsonRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	Result  struct {
		Context struct {
			// now slot
			Slot uint64 `json:"slot"`
		} `json:"context"`
		Value AccountInfo `json:"value"`
	} `json:"result"`
}

type GetHealthResponse struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  string    `json:"result"`
	Error   *RPCError `json:"error,omitempty"`
}

type CreateNonceAccountRequest struct {
	// payer privateKey
	Payer solana.PrivateKey
	// nonce account Auth, PublicKey
	Authority solana.PublicKey
}

type CreateNonceAccountResponse struct {
	// nonce account PublicKey
	NonceAccount solana.PublicKey
	// nonce
	Nonce string
	// nonce
	Signature string
}

type BlockResult struct {
	ParentSlot        uint64              `json:"parentSlot"`        // 父区块 slot，高度比当前块小 1
	BlockTime         int64               `json:"blockTime"`         // 区块产生时间戳（Unix 秒）
	BlockHeight       uint64              `json:"blockHeight"`       // 区块链的总高度（主网启动以来的区块数）
	BlockHash         string              `json:"blockhash"`         // 当前区块哈希
	PreviousBlockhash string              `json:"previousBlockhash"` // 父区块的哈希
	Signatures        []string            `json:"signatures"`        // 当前区块中所有交易的签名（base58）
	Transactions      []TransactionDetail `json:"transactions"`      // 当前区块中每笔交易的详细信息
}

type TransactionDetail struct {
	Signature       string           `json:"signature"`       // 交易签名
	Slot            uint64           `json:"slot"`            // 交易所在 slot（等于区块 slot）
	BlockTime       int64            `json:"blockTime"`       // 交易时间戳（等于区块时间）
	Meta            *TransactionMeta `json:"meta"`            // 交易的执行结果及元数据
	Version         any              `json:"version"`         // 交易版本（"legacy" 或 int，比如 0）
	Message         interface{}      `json:"message"`         // 交易消息（使用 interface{} 可能为 null 或结构体）
	RecentBlockhash string           `json:"recentBlockhash"` // 消耗的 recent blockhash（即签名时使用的）
}

type TransactionMeta struct {
	Err               interface{}     `json:"err"`               // 错误信息，成功为 null，失败为错误结构
	Fee               uint64          `json:"fee"`               // 本次交易消耗的 fee（单位 lamports）
	PreBalances       []uint64        `json:"preBalances"`       // 各账户在执行前的余额
	PostBalances      []uint64        `json:"postBalances"`      // 各账户在执行后的余额
	InnerInstructions []interface{}   `json:"innerInstructions"` // 内部指令（如 CPI 调用）
	PreTokenBalances  []interface{}   `json:"preTokenBalances"`  // 执行前的 SPL Token 余额
	PostTokenBalances []interface{}   `json:"postTokenBalances"` // 执行后的 SPL Token 余额
	LogMessages       []string        `json:"logMessages"`       // 程序日志（可用于调试）
	LoadedAddresses   LoadedAddresses `json:"loadedAddresses"`   // 从地址表中加载的地址
	Status            struct {
		Ok interface{} `json:"Ok"` // 成功状态，若非 null 则表示执行成功
	} `json:"status"`
	Rewards              interface{} `json:"rewards"`              // 本次交易产生的奖励（可能为 null）
	ComputeUnitsConsumed uint64      `json:"computeUnitsConsumed"` // 消耗的计算单元（类似 EVM 的 gas）
}

type TokenBalance struct {
	AccountIndex  int    `json:"accountIndex"` // 对应账户索引
	Mint          string `json:"mint"`         // Token mint 地址
	Owner         string `json:"owner"`        // Token 拥有者
	UITokenAmount struct {
		Amount         string `json:"amount"`
		Decimals       int    `json:"decimals"`
		UIAmountString string `json:"uiAmountString"`
	} `json:"uiTokenAmount"`
}

/*
LoadedAddresses TODO

	在 Solana 中，“加载为只读/可写的地址” 出现在交易结构中的 loadedAddresses.readonly 和
	loadedAddresses.writable 字段，它们来自 Address Lookup Table (地址查找表) 机制。
	📌 背景知识：什么是 Address Lookup Table（地址查找表）？
		Solana 的交易结构对 账户数量有字节限制（传统上是最多 32 个账户），
		为了在不增加交易体积的前提下支持更多地址，Solana 引入了 Address Lookup Table（ALT）：
		ALT 允许交易通过“索引引用”外部表中的账户地址，而不是直接写入交易数据。
		这能让你在交易中引用 多个账户地址而不超出交易大小限制。
	🧠 举个例子
		假设你通过一个 ALT 使用了以下两个地址：
			TokenProgram 账户：只需要读它（判断 Token 类型），不会修改
			UserWallet 账户：要给它转账，涉及写操作
			那么：
			"loadedAddresses": {
			  "readonly": [
				"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
			  ],
			  "writable": [
				"FvZ9P7yfxUmmvxd4XLQK6a..."
			  ]
			}
	🧱 为什么要区分只读和可写？
		Solana 是并发执行交易的链，可读和可写的账户决定了并发是否可能：
			如果两个交易写同一个账户 → 存在冲突，不能并发
			如果两个交易都只读同一个账户 → 可以并发执行
			所以区分是否可写，有助于 并发优化和账户锁策略
	🚀 开发者角度：何时关注这个字段？
		通常不需要手动处理 loadedAddresses，但当你：
			调用复杂合约、CPI（跨程序调用）
			使用地址查找表构造大型交易
			优化吞吐量、调试并发问题
			这些信息就变得非常关键。
*/
type LoadedAddresses struct {
	Readonly []string `json:"readonly"` // 通过地址表加载的 只读 账户地址（只会读取，不会修改状态）（通过 Address Lookup Table）
	Writable []string `json:"writable"` // 通过地址表加载的 可写 账户地址（可能被修改或转账）
}

type TransactionResult struct {
	Slot        uint64          `json:"slot"`        // 所在 Slot
	Version     any             `json:"version"`     // 交易版本 ("legacy" 或 0)
	BlockTime   *int64          `json:"blockTime"`   // 区块时间戳（可选）
	Transaction Transaction     `json:"transaction"` // 原始交易数据
	Meta        TransactionMeta `json:"meta"`        // 执行后的元信息
}

type Transaction struct {
	Message    TransactionMessage `json:"message"`    // 解码后的消息
	Signatures []string           `json:"signatures"` // 交易签名
}

type TransactionMessage struct {
	AccountKeys []string `json:"accountKeys"` // 所有用到的账户地址（按 index 顺序）
	/*
		AddressTableLookups（ALT）作用
			如果是 legacy 交易，该字段为 null 或空；
			如果是 version: 0，且交易中引用了 ALT，此字段包含：
				lookupTableAccount：ALT 的地址
				writableIndexes：从 ALT 中引用的可写地址索引
				readonlyIndexes：从 ALT 中引用的只读地址索引
			这是 Solana 扩展地址空间的方式，用来避免 AccountKeys 超长。
	*/
	AddressTableLookups []interface{}     `json:"addressTableLookups"` // 如果是 v0 交易，这里包含 ALT 使用情况
	Header              TransactionHeader `json:"header"`              // 用于执行权限检查
	Instructions        []Instruction     `json:"instructions"`        // 实际指令集（调用哪个合约、传参）
	RecentBlockhash     string            `json:"recentBlockhash"`     // 防重放用的 recent blockhash
}

/*
TransactionHeader
举个例子：

	一个交易的 AccountKeys 有 5 个地址
	NumRequiredSignatures = 2：前两个必须签名
	NumReadonlySignedAccounts = 1：第 2 个签名账户是只读的
	NumReadonlyUnsignedAccounts = 1：最后一个账户是只读、未签名

可以推断：

	签名写账户 = 1 个
	签名读账户 = 1 个
	未签名写账户 = 2 个
	未签名读账户 = 1 个
*/
type TransactionHeader struct {
	NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`   // 签名过 但只读的账户数量
	NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"` // 未签名 但只读的账户数量
	NumRequiredSignatures       int `json:"numRequiredSignatures"`       // 必须签名的账户数量
}

/*
Instruction
示例：

	ProgramIdIndex = 2 → 表示此指令调用第 3 个 accountKey 指定的合约程序；
	Accounts = [0,1,3] → 表示参数中涉及第 1、2、4 个账户地址；
*/
type Instruction struct {
	Accounts       []int       `json:"accounts"`       // 引用的 accountKeys 下标
	Data           string      `json:"data"`           // 指令数据（base58/hex 编码）
	ProgramIdIndex int         `json:"programIdIndex"` // 调用哪个合约（也是 accountKeys 中的下标）
	StackHeight    interface{} `json:"stackHeight"`    // 内部调用栈高度（一般用于 inner instruction）
}

type PrioritizationFee struct {
	Slot              uint64 `json:"slot"`
	PrioritizationFee uint64 `json:"prioritizationFee"`
}

type SignatureInfo struct {
	Signature          string      `json:"signature"`          // 交易签名（hash）
	Slot               uint64      `json:"slot"`               // 区块 Slot
	Error              interface{} `json:"err"`                // 交易是否出错（null 表示成功）
	Memo               *string     `json:"memo"`               // 可选 memo
	BlockTime          *int64      `json:"blockTime"`          // 区块时间戳（Unix 秒）
	ConfirmationStatus *string     `json:"confirmationStatus"` // 确认状态（如 "confirmed"）
}

type SimulateResult struct {
	Err           interface{} `json:"err"`
	Logs          []string    `json:"logs"`
	UnitsConsumed uint64      `json:"unitsConsumed"`
	Accounts      []struct {
		Executable bool     `json:"executable"`
		Lamports   uint64   `json:"lamports"`
		Owner      string   `json:"owner"`
		RentEpoch  uint64   `json:"rentEpoch"`
		Data       []string `json:"data"`
	} `json:"accounts,omitempty"`
	ReturnData *struct {
		ProgramId string   `json:"programId"`
		Data      []string `json:"data"`
	} `json:"returnData,omitempty"`
	InnerInstructions []struct {
		Index        uint16 `json:"index"`
		Instructions []struct {
			ProgramIdIndex uint8   `json:"programIdIndex"`
			Accounts       []uint8 `json:"accounts"`
			Data           string  `json:"data"`
		} `json:"instructions"`
	} `json:"innerInstructions,omitempty"`
}

/*
SimulateRequest

	Commitment	string	指定模拟交易读取的链上状态确认级别，常见值：processed（默认）、confirmed、finalized。值越高，模拟时数据越稳定但越慢。
	SigVerify	bool	是否在模拟过程中执行签名验证。默认为 false，模拟性能更好；设置为 true 更接近真实发送情况。
	ReplaceRecentBlockhash	bool	是否将交易中使用的 recentBlockhash 替换为当前最新的（防止因旧 blockhash 失败）。
	MinContextSlot	uint64	指定最小 slot，高于这个 slot 才接受此请求，避免旧状态数据影响模拟。通常用于并发和数据一致性控制。
	Encoding	string	交易序列化编码，常见值：base64（默认）、base58，模拟交易建议用 base64（更紧凑，解码快）。
	Accounts	*AccountsInfo	可选，指定要模拟读取的账户列表及其编码方式，会在返回结果中提供这些账户的模拟状态（如余额变动）。
*/
type SimulateRequest struct {
	Commitment             string        `json:"commitment,omitempty"`
	SigVerify              bool          `json:"sigVerify,omitempty"`
	ReplaceRecentBlockhash bool          `json:"replaceRecentBlockhash,omitempty"`
	MinContextSlot         uint64        `json:"minContextSlot,omitempty"`
	Encoding               string        `json:"encoding,omitempty"`
	Accounts               *AccountsInfo `json:"accounts,omitempty"`
}

type GetBalanceResponse struct {
	JsonRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Error   *RPCError `json:"error,omitempty"`
	Result  struct {
		Context struct {
			Slot uint64 `json:"slot"`
		} `json:"context"`
		Value uint64 `json:"value"`
	} `json:"result"`
}

/*
AccountsInfo
Addresses	[]string	要在模拟中附加返回状态信息的账户地址列表。例如用于观察 token account 的变化情况。
Encoding	string	账户数据返回格式，支持：base64、jsonParsed 等。通常选择 jsonParsed 更利于阅读和调试。
*/
type AccountsInfo struct {
	Addresses []string `json:"addresses"`
	Encoding  string   `json:"encoding,omitempty"`
}

type SendTransactionRequest struct {
	/*
			Encoding
			类型：string
			作用：指定签名交易的编码格式
			常见取值：
				"base58"（较常用）

				"base64"（更紧凑）
					base64 编码每 3 字节生成 4 个字符，效率更高；
		            base58 为了避免易混淆字符（如 0 和 O），牺牲了一定编码效率；
		            所以 同样一笔签名交易，base64 编码后的体积通常比 base58 小约 20% 左右。
					在客户端和节点之间传输数据时，base64 可以更快地编码与解码；
					特别是在大量签名交易批量提交时，base64 减少了 JSON 请求的整体字节数，提升带宽利用效率；
					节点解析效率也略高，因为 base64 是标准编码，Go/Rust/JS 均原生支持。
				"json"（调试或查看字段结构用）

			Solana 通常使用 base58 进行交易签名编码。*/
	Encoding string `json:"encoding,omitempty"`
	/*
		Commitment
			类型：string
			作用：客户端希望交易被认为“已确认”的最低区块确认级别
			可选值：
				processed	节点处理过但未确认，最快，但不安全
				confirmed	在一个确认的区块中存在
				finalized	被多数验证者确认，是最安全的状态
				例子：如果你设置了 "finalized"，那么 RPC 返回的结果会等到交易被多个节点共同确认。
	*/
	Commitment string `json:"commitment,omitempty"`
	/*
			前检查（Preflight Check）是 Solana 节点在真正发送交易前 模拟执行 的一种机制，用来提前验证交易是否可能失败。
			它主要包括：
				检查交易的签名是否合法
				检查账户是否存在、余额是否足够
				检查指令是否合法、验证 blockhash 是否有效、账户状态是否正确
				验证交易是否有机会成功上链
			这就像是「上链前做一遍 dry-run」，可以避免：
				错误交易消耗 gas
				交易因为账户状态问题直接被丢弃
				不必要的节点负担
			false：交易会被预先模拟执行（推荐）
		SkipPreflight
			true：直接跳过前检查发送（高性能或对延迟敏感时使用）
			⚠️ 设置为 true 虽然能减少发送延迟，但风险更高（交易失败不会提前发现）。
	*/
	SkipPreflight bool `json:"skipPreflight,omitempty"`
	/*
		 PreflightCommitment
			类型：string
			含义：前检查时使用的区块确认级别
			常见的值有：
				值	含义
				processed	最新状态，响应最快但不稳定
				confirmed	至少被一个确认区块包含
				finalized	被超级多数确认，最安全
	*/
	PreflightCommitment string `json:"preflightCommitment,omitempty"`
	/*
		 MaxRetries
			类型：uint64
			作用：当交易暂时无法上链时，最大重试次数
			用途：
				节点会尝试在后续的 slot 中重新广播这笔交易
				防止因暂时的 slot 拥堵、区块未生成而丢失交易
				适合用在高网络延迟或交易挤压严重的情况。默认值因客户端实现可能为 0（不重试）或无限重试。
	*/
	MaxRetries uint64 `json:"maxRetries,omitempty"`
	/*
		 MinContextSlot
			类型：uint64
			作用：确保在特定 Slot（区块编号）之后才广播交易
			应用场景：
				防止基于旧状态发送交易（如使用旧的 blockhash）
				与 blockhash 的有效性检查结合使用，确保节点最新状态
				举例：如果你设置了 MinContextSlot = 12345，则只有节点同步到了 大于等于 12345 的 slot，才会广播这笔交易。
	*/
	MinContextSlot uint64 `json:"minContextSlot,omitempty"`
}

type AccountInfo struct {
	// account now balance
	Lamports uint64 `json:"lamports"`
	Owner    string `json:"owner"`
	// slice index = 0, data
	// slice index = 1, encode = base58, and other
	Data       []string `json:"data"`
	Executable bool     `json:"executable"`
	RentEpoch  uint64   `json:"rentEpoch"`
	Space      uint64   `json:"space"`
}

type CommitmentType string

func (t CommitmentType) ToString() string {
	return string(t)
}

const (
	// Finalized Confirmed Processed
	// Finalized wait 32 slot
	Finalized CommitmentType = "finalized"
	// Confirmed wait 2-3 slot
	Confirmed CommitmentType = "confirmed"
	// Processed wait 0 slot
	Processed CommitmentType = "processed"
)

type TransactionDetailsType string

const (
	Full       TransactionDetailsType = "full"
	Accounts   TransactionDetailsType = "accounts"
	Signatures TransactionDetailsType = "signatures"
	None       TransactionDetailsType = "none"
)

const (
	HealthOk      = "ok"
	HealthBehind  = "behind"
	HealthUnknown = "unknown"
)

const (
	defaultRequestTimeout   = 30 * time.Second
	defaultRetryCount       = 3
	defaultRetryWaitTime    = 10 * time.Second
	defaultRetryMaxWaitTime = 30 * time.Second
	defaultWithDebug        = false

	blockLimit = 50_0000
)

type TransferType int32

func (t TransferType) ToInt32() int32 {
	return int32(t)
}

/*
类型	判断依据
SOL_TRANSFER	ProgramId == SystemProgram (111111...) 且指令是转账，通常指令 data 为空或 0x02（系统转账）
SPL_TRANSFER	ProgramId == TokenProgram (Tokenkeg...) 且 data[0] == 3 且 decimals > 0
SPL_NFT_TRANSFER	同 SPL_TRANSFER，但 mint 对应 metadata 存在，且 decimals == 0
CONTRACT_CALL	除以上之外的程序调用，如 ProgramId != System/Token，尤其是 Metaplex、Bubblegum 等程序
*/
const (
	TypeSolTransfer  TransferType = 1 // "SOL_TRANSFER"
	TypeSplTransfer  TransferType = 2 // "SPL_TRANSFER"
	TypeNftTransfer  TransferType = 3 // "SPL_NFT_TRANSFER"
	TypeContractCall TransferType = 4 //"CONTRACT_CALL"
)
