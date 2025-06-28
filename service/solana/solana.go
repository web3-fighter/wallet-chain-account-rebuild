package solana

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/log"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/web3-fighter/chain-explorer-api/types"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
	"github.com/web3-fighter/wallet-chain-account/service/svmbase"
	"github.com/web3-fighter/wallet-chain-account/service/unimplemente"
	"math"
	"strconv"
)

const ChainName = "Solana"

const (
	MaxBlockRange = 1000
)

var _ service.WalletAccountService = (*SOLNodeService)(nil)

type SOLNodeService struct {
	svmClient svmbase.SVMClient
	sdkClient *rpc.Client
	solData   *svmbase.SolData
	unimplemente.UnimplementedService
}

//func (s *SOLNodeService) GetSupportChains(ctx context.Context, param domain.SupportChainsParam) (bool, error) {
//	//TODO implement me
//	panic("implement me")
//}

func (s *SOLNodeService) ConvertAddress(_ context.Context, param domain.ConvertAddressParam) (string, error) {
	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		err := fmt.Errorf("GetSupportChains validateChainAndNetwork fail, err msg = %s", msg)
		log.Error("err", err)
		return "", err
	}
	pubKeyHex := param.PublicKey
	if ok, msg := validatePublicKey(pubKeyHex); !ok {
		err := fmt.Errorf("ConvertAddress validatePublicKey fail, err msg = %s", msg)
		log.Error("err", err)
		return "", err
	}
	accountAddress, err := svmbase.PubKeyHexToAddress(pubKeyHex)
	if err != nil {
		err = fmt.Errorf("ConvertAddress PubKeyHexToAddress failed: %w", err)
		log.Error("err", err)
		return "", err
	}
	return accountAddress, nil

}

func (s *SOLNodeService) ValidAddress(_ context.Context, param domain.ValidAddressParam) (bool, error) {
	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		err := fmt.Errorf("ValidAddress validateChainAndNetwork failed: %s", msg)
		log.Error("err", err)
		return false, err
	}
	address := param.Address
	if len(address) == 0 {
		err := fmt.Errorf("ValidAddress address is empty")
		log.Error("err", err)
		return false, err
	}
	if len(address) != 43 && len(address) != 44 {
		err := fmt.Errorf("invalid Solana address length: expected 43 or 44 characters, got %d", len(address))
		return false, err
	}
	return true, nil
}

func (s *SOLNodeService) GetBlockByNumber(ctx context.Context, param domain.BlockNumberParam) (domain.Block, error) {
	response := domain.Block{}

	if ok, msg := validateChainAndNetwork(param.Chain, ""); !ok {
		err := fmt.Errorf("GetBlockByNumber validateChainAndNetwork failed: %s", msg)
		log.Error("err", err)
		return response, err
	}
	resultSlot := uint64(param.Height)
	if param.Height == 0 {
		latestSlot, err := s.svmClient.GetSlot(ctx, svmbase.Finalized)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetSlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		resultSlot = latestSlot
	}

	blockResult := &svmbase.BlockResult{}
	if param.ViewTx {
		tempBlockBySlot, err := s.svmClient.GetBlockBySlot(ctx, resultSlot, svmbase.Signatures)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetBlockBySlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		blockResult = tempBlockBySlot
	} else {
		tempBlockBySlot, err := s.svmClient.GetBlockBySlot(ctx, resultSlot, svmbase.None)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetBlockBySlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		blockResult = tempBlockBySlot
	}

	// 填充基本字段
	response.Hash = blockResult.BlockHash
	response.Height = int64(resultSlot)
	// 如果只展示 tx hash，可直接返回
	if param.ViewTx {
		// 遍历区块内每笔交易并解析
		for _, tx := range blockResult.Transactions {
			bt := parseBlockTransaction(tx)
			response.Transactions = append(response.Transactions, bt)
		}
	}
	return response, nil
}

func (s *SOLNodeService) GetBlockByHash(ctx context.Context, param domain.BlockHashParam) (domain.Block, error) {
	response := domain.Block{}
	if ok, msg := validateChainAndNetwork(param.Chain, ""); !ok {
		err := fmt.Errorf("GetBlockByHash validateChainAndNetwork fail, err msg = %s", msg)
		return response, err
	}

	blockResult, err := s.svmClient.GetBlockByHash(ctx, param.Hash)
	if err != nil {
		return response, err
	}
	// 填充基本字段
	response.Hash = blockResult.BlockHash
	response.Height = int64(blockResult.BlockHeight)
	// 如果只展示 tx hash，可直接返回
	if param.ViewTx {
		// 遍历区块内每笔交易并解析
		for _, tx := range blockResult.Transactions {
			bt := parseBlockTransaction(tx)
			response.Transactions = append(response.Transactions, bt)
		}
	}
	return response, nil
}

func (s *SOLNodeService) GetBlockHeaderByNumber(ctx context.Context, param domain.BlockHeaderNumberParam) (domain.BlockHeader, error) {
	response := domain.BlockHeader{}
	if ok, msg := validateChainAndNetwork(param.Chain, ""); !ok {
		err := fmt.Errorf("GetBlockHeaderByNumber validateChainAndNetwork failed: %s", msg)
		log.Error("err", err)
		return response, err
	}

	resultSlot := uint64(param.Height)
	if param.Height == 0 {
		latestSlot, err := s.svmClient.GetSlot(ctx, svmbase.Finalized)
		if err != nil {
			err = fmt.Errorf("GetBlockHeaderByNumber GetSlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		resultSlot = latestSlot
	}

	blockResult, err := s.svmClient.GetBlockBySlot(ctx, resultSlot, svmbase.None)
	if err != nil {
		err = fmt.Errorf("GetBlockHeaderByNumber GetBlockBySlot failed: %w", err)
		log.Error("err", err)
		return response, err
	}
	response.Hash = blockResult.BlockHash
	response.Number = strconv.FormatUint(resultSlot, 10)
	response.ParentHash = blockResult.PreviousBlockhash
	response.Time = uint64(blockResult.BlockTime)
	return response, nil
}

func (s *SOLNodeService) GetBlockHeaderByHash(ctx context.Context, param domain.BlockHeaderHashParam) (domain.BlockHeader, error) {
	response := domain.BlockHeader{}
	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		err := fmt.Errorf("GetBlockByHash validateChainAndNetwork fail, err msg = %s", msg)
		return response, err
	}

	blockResult, err := s.svmClient.GetBlockByHash(ctx, param.Hash)
	if err != nil {
		return response, err
	}
	response.Hash = blockResult.BlockHash
	response.Number = strconv.FormatUint(blockResult.BlockHeight, 10)
	response.ParentHash = blockResult.PreviousBlockhash
	response.Time = uint64(blockResult.BlockTime)
	return response, nil
}

func (s *SOLNodeService) ListBlockHeaderByRange(ctx context.Context, param domain.BlockHeaderByRangeParam) ([]domain.BlockHeader, error) {
	if err := validateBlockRangeParam(param); err != nil {
		return nil, err
	}
	startSlot, _ := strconv.ParseUint(param.Start, 10, 64)
	endSlot, _ := strconv.ParseUint(param.End, 10, 64)

	resBlockHeaders := make([]domain.BlockHeader, 0, endSlot-startSlot+1)
	for slot := startSlot; slot <= endSlot; slot++ {
		blockResult, err := s.svmClient.GetBlockBySlot(ctx, slot, svmbase.Signatures)
		if err != nil {
			if len(resBlockHeaders) > 0 {
				return resBlockHeaders, fmt.Errorf("partial success, stopped at slot %d: %v", slot, err)
			}
			return resBlockHeaders, fmt.Errorf("failed to get signatures for slot %d: %v", slot, err)
		}
		if len(blockResult.Signatures) == 0 {
			continue
		}
		txResults, err := s.svmClient.GetTransactionRange(ctx, blockResult.Signatures)
		if err != nil {
			if len(resBlockHeaders) > 0 {
				return resBlockHeaders, fmt.Errorf("partial success, stopped at slot %d: %v", slot, err)
			}
			return resBlockHeaders, fmt.Errorf("failed to get transactions for slot %d: %v", slot, err)
		}
		blockHeaders, err := organizeTransactionsByBlock(txResults)
		if err != nil {
			if len(resBlockHeaders) > 0 {
				return resBlockHeaders, fmt.Errorf("partial success, stopped at slot %d: %v", slot, err)
			}
			return resBlockHeaders, fmt.Errorf("failed to organize transactions for slot %d: %v", slot, err)
		}

		if len(blockHeaders) > 0 {
			resBlockHeaders = append(resBlockHeaders, blockHeaders...)
		}
	}

	if len(resBlockHeaders) == 0 {
		return nil, errors.New("no transactions found in range")
	}

	return resBlockHeaders, nil
}

func (s *SOLNodeService) GetAccount(ctx context.Context, param domain.AccountParam) (domain.Account, error) {
	response := domain.Account{}
	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		return response, fmt.Errorf("GetAccount validateChainAndNetwork fail, err msg = %s", msg)
	}
	accountInfoResp, err := s.svmClient.GetAccountInfo(ctx, param.Address)

	if err != nil {
		err = fmt.Errorf("GetAccount GetAccountInfo failed: %w", err)
		log.Error("err", err)
		return response, fmt.Errorf("GetAccount GetAccountInfo failed: %w", err)
	}
	latestBlockHashResponse, err := s.svmClient.GetLatestBlockHash(ctx, svmbase.Finalized)
	if err != nil {
		err = fmt.Errorf("GetAccount GetLatestBlockhash failed: %w", err)
		log.Error("err", err)
		return response, err
	}

	response.Sequence = latestBlockHashResponse
	response.Network = param.Network
	response.Balance = strconv.FormatUint(accountInfoResp.Lamports, 10)
	return response, nil
}

func (s *SOLNodeService) GetFee(ctx context.Context, param domain.FeeParam) (domain.Fee, error) {
	response := domain.Fee{}
	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		return response, fmt.Errorf("GetFee validateChainAndNetwork fail, err msg = %s", msg)
	}
	baseFee, err := s.svmClient.GetFeeForMessage(ctx, param.RawTx)
	if err != nil {
		err = fmt.Errorf("GetFee GetFeeForMessage failed: %w", err)
		log.Error("err", err)
		return response, err
	}
	priorityFees, err := s.svmClient.GetRecentPrioritizationFees(ctx)
	if err != nil {
		err = fmt.Errorf("GetFee GetRecentPrioritizationFees failed: %w", err)
		log.Error("err", err)
		return response, err
	}
	priorityFee := svmbase.GetSuggestedPriorityFee(priorityFees)
	slowFee := baseFee + uint64(float64(priorityFee)*0.75)
	normalFee := baseFee + priorityFee
	fastFee := baseFee + uint64(float64(priorityFee)*1.25)

	response.SlowFee = domain.GasFee{GasPrice: strconv.FormatUint(slowFee, 10)}
	response.NormalFee = domain.GasFee{GasPrice: strconv.FormatUint(normalFee, 10)}
	response.FastFee = domain.GasFee{GasPrice: strconv.FormatUint(fastFee, 10)}

	return response, nil
}

func (s *SOLNodeService) SendTx(ctx context.Context, param domain.SendTxParam) (string, error) {
	if param.RawTx == "" {
		return "", errors.New("invalid input: empty transaction")
	}
	// Send the transaction
	txHash, err := s.svmClient.SendTransaction(ctx, param.RawTx, nil)
	if err != nil {
		log.Error("Failed to send transaction", "err", err)
		return "", err
	}

	return txHash, nil
}

func (s *SOLNodeService) ListTxByAddress(ctx context.Context, param domain.TxAddressParam) ([]domain.TxMessage, error) {
	var resp *types.TransactionResponse[types.AccountTxResponse]
	var err error
	if param.ContractAddress != "0x00" && param.ContractAddress != "" {
		log.Info("Spl token transfer record")
		resp, err = s.solData.GetTxByAddress(uint64(param.Page), uint64(param.PageSize), param.Address, "spl")
	} else {
		log.Info("Sol transfer record")
		resp, err = s.solData.GetTxByAddress(uint64(param.Page), uint64(param.PageSize), param.Address, "sol")
	}
	if err != nil {
		log.Error("get GetTxByAddress error", "err", err)
		return nil, errors.New("get tx list fail")
	} else {
		txs := resp.TransactionList
		txMessages := make([]domain.TxMessage, 0, len(txs))
		for i := 0; i < len(txs); i++ {
			txMessages = append(txMessages, domain.TxMessage{
				Hash:   txs[i].TxId,
				Tos:    []string{txs[i].To},
				Froms:  []string{txs[i].From},
				Fee:    txs[i].TxId,
				Status: domain.TxStatus_Success,
				Values: []string{txs[i].Amount},
				Type:   1,
				Height: txs[i].Height,
			})
		}
		return txMessages, nil
	}
}

func (s *SOLNodeService) GetTxByHash(ctx context.Context, param domain.GetTxByHashParam) (domain.TxMessage, error) {
	if err := validateParam(param); err != nil {
		return domain.TxMessage{}, err
	}
	txResult, err := s.svmClient.GetTransaction(ctx, param.Hash)
	if err != nil {
		log.Error("GetTransaction failed", "error", err)
		return domain.TxMessage{}, err
	}

	txMessage, err := buildTxMessage(txResult)
	if err != nil {
		return txMessage, err
	}

	return txMessage, nil
}

// CreateUnSignTransaction 根据用户提供的 Base64 编码交易参数，
// 构造一个未签名的 Solana 原始交易（支持 SOL 和 SPL Token 转账）并返回其十六进制格式的签名消息体。
func (s *SOLNodeService) CreateUnSignTransaction(ctx context.Context, param domain.UnSignTransactionParam) (string, error) {
	// Decode the base64 transaction string
	jsonBytes, err := base64.StdEncoding.DecodeString(param.Base64Tx)
	if err != nil {
		log.Error("Failed to decode base64 string", "err", err)
		return "", err
	}

	// Unmarshal JSON into TxStructure
	var data TxStructure
	// 解析 base64 编码并转换为结构体
	if err = json.Unmarshal(jsonBytes, &data); err != nil {
		log.Error("Failed to parse JSON", "err", err)
		return "", err
	}

	// 计算转账金额 先将 string 转 float，再根据精度转换为整数：
	//编辑
	// Parse the value from string to float
	valueFloat, err := strconv.ParseFloat(data.Value, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse value: %w", err)
	}
	value := uint64(valueFloat * 1000000000)

	// 地址转换为 PublicKey
	// Convert from address to public key
	fromPubKey, err := solana.PublicKeyFromBase58(data.FromAddress)
	if err != nil {
		return "", err
	}

	// 地址转换为 PublicKey
	// Convert to address to public key
	toPubKey, err := solana.PublicKeyFromBase58(data.ToAddress)
	if err != nil {
		return "", err
	}

	// 判断是 SOL 转账还是 SPL Token 转账
	var tx *solana.Transaction
	if isSOLTransfer(data.ContractAddress) {
		// Create a new SOL transfer transaction
		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					value,
					fromPubKey,
					toPubKey,
				).Build(),
			},
			solana.MustHashFromBase58(data.Nonce),
			solana.TransactionPayer(fromPubKey),
		)
	} else {
		// Handle SPL token transfer
		mintPubKey := solana.MustPublicKeyFromBase58(data.ContractAddress)

		// 查找关联账户（ATA）fromTokenAccount
		fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			fromPubKey,
			mintPubKey,
		)
		if err != nil {
			return "", fmt.Errorf("failed to find associated token address: %w", err)
		}
		// 查找关联账户（ATA）toTokenAccount
		toTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			toPubKey,
			mintPubKey,
		)
		if err != nil {
			return "", fmt.Errorf("failed to find associated token address: %w", err)
		}

		// 获取 token 的信息
		tokenInfo, err := s.GetTokenSupply(ctx, mintPubKey)
		if err != nil {
			return "", fmt.Errorf("failed to get token info: %w", err)
		}
		// 获取 token 的 decimals 精度
		decimals := tokenInfo.Value.Decimals

		actualValue := uint64(valueFloat * math.Pow10(int(decimals)))

		// 构造转账指令
		transferInstruction := token.NewTransferInstruction(
			actualValue,
			fromTokenAccount,
			toTokenAccount,
			fromPubKey,
			[]solana.PublicKey{},
		).Build()

		//  查找目标 关联 token account
		accountInfo, err := s.GetAccountInfo(ctx, toTokenAccount)

		//  检查目标 关联 token account 是否存在，不存在则创建
		if err != nil || accountInfo.Value == nil {
			/*
				在 Solana 的 SPL Token 标准中，ATA（Associated Token Account） 是每个钱包地址针对某个 Token 所专
				属的账户。你不能直接用钱包地址去接收 Token，必须创建一个 Associated Token Account (ATA)，就像是：
				💡“Token 的银行子账户，用来存储某种特定 Token 的余额。”

				📦 什么是 ATA？
					每个钱包地址 + 每个 Token → 唯一的一个 Token Account（ATA）
					这个 ATA 是用来接收和持有某个 SPL Token 的
					ATA 是可以通过标准算法计算得出（无需在链上查询）

				❓为什么需要构造 ATA？
					在你给别人转 SPL Token 时，如果 目标地址还没有 ATA，Token 就没地方存，转账会失败。
					所以需要判断一下目标地址是否有 ATA，如果没有，就先创建它。
			*/
			// Create associated token account if it doesn't exist
			// 构造一个创建「关联 Token 子账户（ATA）」的链上指令，在执行交易前确保目标地址能正确接收 SPL Token。
			createATAInstruction := associatedtokenaccount.NewCreateInstruction(
				fromPubKey,
				toPubKey,
				mintPubKey,
			).Build()

			tx, err = solana.NewTransaction(
				[]solana.Instruction{createATAInstruction, transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
		} else {
			// 直接构造转账
			// Directly create transfer transaction
			tx, err = solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
		}
	}

	// Log the transaction details
	log.Info("Transaction:", tx.String())

	// Serialize the transaction message
	txm, _ := tx.Message.MarshalBinary()
	signingMessageHex := hex.EncodeToString(txm)

	// Return the unsigned transaction response
	return signingMessageHex, nil
}

// GetAccountInfo retrieves account information for a given token account
func (s *SOLNodeService) GetAccountInfo(ctx context.Context, tokenAccount solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
	accountInfo, err := s.sdkClient.GetAccountInfo(ctx, tokenAccount)
	if err != nil {
		log.Info("Failed to get account info", "err", err)
		return nil, err
	}
	return accountInfo, nil
}

// GetTokenSupply retrieves the token supply for a given mint public key
func (s *SOLNodeService) GetTokenSupply(ctx context.Context, mintPubKey solana.PublicKey) (*rpc.GetTokenSupplyResult, error) {
	tokenInfo, err := s.sdkClient.GetTokenSupply(ctx, mintPubKey, rpc.CommitmentFinalized)
	if err != nil {
		log.Info("Failed to get token supply", "err", err)
		return nil, err
	}
	return tokenInfo, nil
}

// BuildSignedTransaction 实现了 Solana 的交易构造 + 签名绑定过程
/*
✅ 第 1 步：创建待签名交易结构（在线系统）
方法：CreateUnSignTransaction()
🟢 你做了什么：
	构造了包含如下信息的交易结构体 TxStructure：
		FromAddress
		ToAddress
		Nonce
		Value
		ContractAddress（判断是否是 SPL Token）
	SPL 情况下还处理 ATA、decimals、构造 transfer 指令等
	将这个结构 base64 编码，作为签名输入
	序列化 Message，用于签名（注意：不是整个交易体）

✅ 为什么这么做：
	冷钱包不能访问网络，交易需要热端构造好 Message 数据
	冷端只需对 message 签名，不需要知道链的状态
	Message 是交易的 digest 部分，签名仅对其做 Ed25519 签名

✅ 第 2 步：离线端签名（冷钱包 / 离线系统）
🟢 离线钱包做了什么：
	对 base64Tx 解码，恢复出交易结构
	解析出待签名的 Message（来自 MarshalBinary）
	使用私钥对 Message 做 Ed25519 签名
	返回 hex 格式的签名（64字节）

✅ 为什么这么做：
	避免私钥暴露，离线设备不可联网
	所有签名过程必须在离线设备完成
	hex 格式方便传输回热端

✅ 第 3 步：热端绑定签名 + 构造完整交易
方法：BuildSignedTransaction()

🟢 你做了什么：
	再次恢复 base64Tx → TxStructure
	构造完整的 Transaction
	将 Signature 插入到 Transaction.Signatures[0]
	可选：验证签名是否正确（tx.VerifySignatures()）

✅ 为什么这么做：
	Solana 交易签名结构是：
		Transaction {
		  Signatures []Signature
		  Message    Message
		}
	你只绑定签名而不更改 Message，保证签名有效。

✅ 第 4 步：序列化并编码为 base58Tx
	🟢 你做了什么：
		tx.MarshalBinary() → 二进制交易
		base58.Encode() → 返回可广播格式

	✅ 为什么这么做：
		Solana 网络要求 sendTransaction 入参是 base58 编码的完整二进制交易体
		和 Metamask 签名结构类似，但是不同链的格式

✅ 第 5 步：广播交易到网络
	方法：sendTransaction(signedTxBase58)
	🟢 你做了什么：
		调用 RPC，如：
		curl https://api.mainnet-beta.solana.com \
		  -X POST \
		  -H "Content-Type: application/json" \
		  -d '{"jsonrpc":"2.0","id":1,"method":"sendTransaction","params":["<base58Tx>"]}'
		✅ 为什么这么做：
		这是 Solana 唯一接受完整签名交易的广播接口

		一次广播后交易进入 mempool



*/
func (s *SOLNodeService) BuildSignedTransaction(ctx context.Context, param domain.SignedTransactionParam) (domain.SignedTransaction, error) {
	signedTransaction := domain.SignedTransaction{}
	// Decode the base64 transaction string
	jsonBytes, err := base64.StdEncoding.DecodeString(param.Base64Tx)
	if err != nil {
		log.Error("Failed to decode base64 string", "err", err)
		return signedTransaction, err
	}

	// Unmarshal JSON into TxStructure
	var data TxStructure
	if err = json.Unmarshal(jsonBytes, &data); err != nil {
		log.Error("Failed to parse JSON", "err", err)
		return signedTransaction, err
	}

	// Parse the value from string to float
	valueFloat, err := strconv.ParseFloat(data.Value, 64)
	if err != nil {
		return signedTransaction, fmt.Errorf("failed to parse value: %w", err)
	}
	value := uint64(valueFloat * 1000000000)

	// Convert from address to public key
	fromPubKey, err := solana.PublicKeyFromBase58(data.FromAddress)
	if err != nil {
		return signedTransaction, err
	}

	// Convert to address to public key
	toPubKey, err := solana.PublicKeyFromBase58(data.ToAddress)
	if err != nil {
		return signedTransaction, err
	}

	var tx *solana.Transaction
	if isSOLTransfer(data.ContractAddress) {
		// Create a new SOL transfer transaction
		tx, err = solana.NewTransaction(
			[]solana.Instruction{
				system.NewTransferInstruction(
					value,
					fromPubKey,
					toPubKey,
				).Build(),
			},
			solana.MustHashFromBase58(data.Nonce),
			solana.TransactionPayer(fromPubKey),
		)
	} else {
		// Handle SPL token transfer
		mintPubKey := solana.MustPublicKeyFromBase58(data.ContractAddress)

		fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			fromPubKey,
			mintPubKey,
		)
		if err != nil {
			return signedTransaction, fmt.Errorf("failed to find associated token address: %w", err)
		}

		toTokenAccount, _, err := solana.FindAssociatedTokenAddress(
			toPubKey,
			mintPubKey,
		)
		if err != nil {
			return signedTransaction, fmt.Errorf("failed to find associated token address: %w", err)
		}

		//tokenInfo, err := c.sdkClient.GetTokenSupply(context.Background(), mintPubKey, rpc.CommitmentFinalized)
		tokenInfo, err := s.GetTokenSupply(ctx, mintPubKey)
		if err != nil {
			return signedTransaction, fmt.Errorf("Failed to get token info: %w", err)
		}
		decimals := tokenInfo.Value.Decimals

		actualValue := uint64(valueFloat * math.Pow10(int(decimals)))

		transferInstruction := token.NewTransferInstruction(
			actualValue,
			fromTokenAccount,
			toTokenAccount,
			fromPubKey,
			[]solana.PublicKey{},
		).Build()
		//accountInfo, err := c.sdkClient.GetAccountInfo(context.Background(), toTokenAccount)
		accountInfo, err := s.GetAccountInfo(ctx, toTokenAccount)

		if err != nil || accountInfo.Value == nil {
			// Create associated token account if it doesn't exist
			createATAInstruction := associatedtokenaccount.NewCreateInstruction(
				fromPubKey,
				toPubKey,
				mintPubKey,
			).Build()

			tx, err = solana.NewTransaction(
				[]solana.Instruction{createATAInstruction, transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
		} else {
			// Directly create transfer transaction
			tx, err = solana.NewTransaction(
				[]solana.Instruction{transferInstruction},
				solana.MustHashFromBase58(data.Nonce),
				solana.TransactionPayer(fromPubKey),
			)
		}
	}

	// Ensure the Signatures slice is initialized
	if len(tx.Signatures) == 0 {
		tx.Signatures = make([]solana.Signature, 1)
	}

	// Decode the signature from hex
	//  签名处理
	signatureBytes, err := hex.DecodeString(data.Signature)
	if err != nil {
		log.Error("Failed to decode hex signature", "err", err)
	}

	// Verify the signature length
	if len(signatureBytes) != 64 {
		log.Error("Invalid signature length", "length", len(signatureBytes))
	}

	// Convert to Solana Signature
	var solSignature solana.Signature
	copy(solSignature[:], signatureBytes)

	// Set the signature
	tx.Signatures[0] = solSignature

	// 验证签名是否正确（非强制）
	// Dump the transaction for debugging
	spew.Dump(tx)
	if err = tx.VerifySignatures(); err != nil {
		log.Info("Invalid signatures", "err", err)
	}

	// Serialize the transaction
	serializedTx, err := tx.MarshalBinary()
	if err != nil {
		return signedTransaction, fmt.Errorf("dailed to serialize transaction: %w", err)
	}

	// Encode the serialized transaction to base58
	base58Tx := base58.Encode(serializedTx)
	//base64Tx := base64.StdEncoding.EncodeToString(serializedTx)
	signedTransaction.SignedTx = base58Tx
	return signedTransaction, nil
}

func (s *SOLNodeService) DecodeTransaction(ctx context.Context, param domain.DecodeTransactionParam) (string, error) {
	// Decode base58 encoded transaction
	rawTx, err := base58.Decode(param.RawTx)
	if err != nil {
		return "", fmt.Errorf("failed to decode base58 transaction: %w", err)
	}

	// Unmarshal binary transaction
	tx := &solana.Transaction{}
	dec := bin.NewBinDecoder(rawTx)
	err = tx.UnmarshalWithDecoder(dec)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction: %w", err)
	}
	message := tx.Message

	// Prepare result struct
	var result TxStructure
	result.Nonce = message.RecentBlockhash.String()
	result.GasPrice = "0"  // Solana不使用
	result.GasTipCap = "0" // Solana不使用
	result.GasFeeCap = "0" // Solana不使用
	result.Gas = 0         // Solana没有 gas

	if len(message.AccountKeys) > 0 {
		result.FromAddress = message.AccountKeys[0].String()
	}

	// 查找签名（只取第一个）
	if len(tx.Signatures) > 0 {
		result.Signature = hex.EncodeToString(tx.Signatures[0][:])
	}

	// 遍历指令分析交易类型
	for _, instr := range message.Instructions {
		program := message.AccountKeys[instr.ProgramIDIndex]
		switch program.String() {
		case solana.SystemProgramID.String():
			// 系统转账 (SOL transfer)
			result.ContractAddress = "" // native token 没有合约地址
			result.TokenId = ""         // 非 NFT
			result.Value = fmt.Sprintf("%.9f", float64(binary.LittleEndian.Uint64(instr.Data))/1e9)
			if len(instr.Accounts) >= 2 {
				toIdx := instr.Accounts[1]
				result.ToAddress = message.AccountKeys[toIdx].String()
			}
		case solana.TokenProgramID.String():
			// TODO 为解析合约地址 和 token id
			// SPL 转账或 NFT 转移
			if len(instr.Data) > 0 && instr.Data[0] == 3 {
				// SPL Transfer
				result.Value = fmt.Sprintf("%.0f", float64(binary.LittleEndian.Uint64(instr.Data[1:])))
				result.ContractAddress = program.String() // token 合约地址应另行补充（见下方建议）
				if len(instr.Accounts) >= 2 {
					toIdx := instr.Accounts[1]
					result.ToAddress = message.AccountKeys[toIdx].String()
				}
				result.TokenId = "" // 不是 NFT
			} else if len(instr.Data) > 0 && instr.Data[0] == 12 {
				// TransferChecked，可能是 NFT
				if len(instr.Accounts) >= 2 {
					toIdx := instr.Accounts[1]
					result.ToAddress = message.AccountKeys[toIdx].String()
				}
				result.Value = "1"
				result.TokenId = "?" // NFT 的 ID 一般需要额外解析 metadata account
			}
		case solana.SPLAssociatedTokenAccountProgramID.String():
			// 创建 ATA，无需设置 Value
			// 通常前一个 Transfer 指令已处理
			continue
		default:
			// 其他合约调用，可忽略或记录日志
		}
	}

	// JSON 编码返回
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal TxStructure: %w", err)
	}

	return string(jsonBytes), nil
}

func (s *SOLNodeService) VerifySignedTransaction(ctx context.Context, param domain.VerifyTransactionParam) (bool, error) {
	txBytes, err := base58.Decode(param.Signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode transaction: %w", err)
	}

	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return false, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	if err = tx.VerifySignatures(); err != nil {
		log.Info("Invalid signatures", "err", err)
		return false, nil
	}

	return true, nil
}

func (s *SOLNodeService) GetExtraData(ctx context.Context, param domain.ExtraDataParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewSOLNodeService(svmClient svmbase.SVMClient, sdkClient *rpc.Client, solData *svmbase.SolData) service.WalletAccountService {
	return &SOLNodeService{
		sdkClient: sdkClient,
		svmClient: svmClient,
		solData:   solData,
	}
}
