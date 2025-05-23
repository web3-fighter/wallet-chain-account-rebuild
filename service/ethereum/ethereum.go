package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/shopspring/decimal"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/web3-fighter/chain-explorer-api/types"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
	"github.com/web3-fighter/wallet-chain-account/service/donoting"
	"github.com/web3-fighter/wallet-chain-account/service/evmbase"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const (
	ChainName        = "Ethereum"
	ContractTransfer = "contract"
)

var _ service.WalletAccountService = (*EthNodeService)(nil)

type EthNodeService struct {
	ethClient     evmbase.EVMClient
	ethDataClient *evmbase.EthScan
	donoting.DoNotingService
}

// ConvertAddress 将传入的十六进制字符串形式的 公钥 转换为 以太坊地址。
func (s *EthNodeService) ConvertAddress(_ context.Context, param domain.ConvertAddressParam) (string, error) {
	// param.PublicKey：是未经压缩的公钥（通常是 130 个字符，0x04 开头）。
	// hex.DecodeString(...)：将公钥字符串转为字节切片。
	// 如果解码失败，就返回一个空地址 0x0000000000000000000000000000000000000000。
	publicKeyBytes, err := hex.DecodeString(param.PublicKey)
	if err != nil {
		return common.Address{}.String(), nil
	}
	/*
				publicKeyBytes[1:]：跳过第一个字节（0x04，表示未压缩公钥）。

				crypto.Keccak256(publicKeyBytes[1:])：对剩下的 64 字节（x、y 坐标）做 Keccak-256 哈希。
				crypto.Keccak256(publicKey[1:]) 结果 是 32 字节
				[12:]：取后 20 字节作为以太坊地址（以太坊地址就是公钥哈希的最后 20 字节）。

			0        11     12                                      31
			|--------|----------------------------------------------|
			[ 前12字节丢弃 ][     后20字节 => address 字节内容     ]

		[12:] 是 以太坊地址生成规范 的要求 —— 取 Keccak256 哈希的后 20 字节（32 - 20 = 12）。
		这是 不是随意取的偏移量，而是协议定义的标准。
	*/
	addressCommon := common.BytesToAddress(crypto.Keccak256(publicKeyBytes[1:])[12:])
	return addressCommon.String(), nil
}

func (s *EthNodeService) ValidAddress(_ context.Context, param domain.ValidAddressParam) (bool, error) {
	//以太坊地址 = 0x + 40位十六进制字符 → 长度必须是 42。
	//必须以 "0x" 开头，否则格式不合法。
	if len(param.Address) != 42 || !strings.HasPrefix(param.Address, "0x") {
		return false, nil
	}
	//用正则校验 "0x" 后面的部分是否为 40 位合法的十六进制字符（不区分大小写）。
	isValid := regexp.MustCompile("^[0-9a-fA-F]{40}$").MatchString(param.Address[2:])
	return isValid, nil
}

func (s *EthNodeService) GetBlockByNumber(ctx context.Context, param domain.BlockNumberParam) (domain.Block, error) {
	block, err := s.ethClient.BlockByNumber(ctx, big.NewInt(param.Height))
	if err != nil {
		log.Error("block by number error", err)
		return domain.Block{}, fmt.Errorf("block by number error: %w", err)
	}
	blockNumber, _ := block.NumberUint64()
	var txListRet []*domain.BlockTransaction
	for _, v := range block.Transactions {
		txItem := &domain.BlockTransaction{
			From: v.From,
			To:   v.To,
			//TokenAddress:   v.To,
			//ContractWallet: v.To,
			Hash:   v.Hash,
			Height: blockNumber,
			Amount: v.Value,
		}
		txListRet = append(txListRet, txItem)
	}
	return domain.Block{
		Height:       int64(blockNumber),
		Hash:         block.Hash.String(),
		BaseFee:      block.BaseFee,
		Transactions: txListRet,
	}, nil
}

func (s *EthNodeService) GetBlockByHash(ctx context.Context, param domain.BlockHashParam) (domain.Block, error) {
	block, err := s.ethClient.BlockByHash(ctx, common.HexToHash(param.Hash))
	if err != nil {
		log.Error("block by hash error", err)
		return domain.Block{}, fmt.Errorf("block by hash error: %w", err)
	}
	blockNumber, _ := block.NumberUint64()
	var txListRet []*domain.BlockTransaction
	for _, v := range block.Transactions {
		txItem := &domain.BlockTransaction{
			From: v.From,
			To:   v.To,
			//TokenAddress:   v.To,
			//ContractWallet: v.To,
			Hash:   v.Hash,
			Amount: v.Value,
			Height: blockNumber,
		}
		txListRet = append(txListRet, txItem)
	}
	return domain.Block{
		Height:       int64(blockNumber),
		Hash:         block.Hash.String(),
		BaseFee:      block.BaseFee,
		Transactions: txListRet,
	}, nil
}

func (s *EthNodeService) GetBlockHeaderByHash(ctx context.Context, param domain.BlockHeaderHashParam) (domain.BlockHeader, error) {
	blockInfo, err := s.ethClient.BlockHeaderByHash(ctx, common.HexToHash(param.Hash))
	if err != nil {
		log.Error("get latest block header fail", "err", err)
		return domain.BlockHeader{}, fmt.Errorf("get latest block header fail: %w", err)
	}
	blockHeader := domain.BlockHeader{
		Hash:             blockInfo.Hash().String(),
		ParentHash:       blockInfo.ParentHash.String(),
		UncleHash:        blockInfo.UncleHash.String(),
		CoinBase:         blockInfo.Coinbase.String(),
		Root:             blockInfo.Root.String(),
		TxHash:           blockInfo.TxHash.String(),
		ReceiptHash:      blockInfo.ReceiptHash.String(),
		ParentBeaconRoot: blockInfo.ParentBeaconRoot.String(),
		Difficulty:       blockInfo.Difficulty.String(),
		Number:           blockInfo.Number.String(),
		GasLimit:         blockInfo.GasLimit,
		GasUsed:          blockInfo.GasUsed,
		Time:             blockInfo.Time,
		Extra:            string(blockInfo.Extra),
		MixDigest:        blockInfo.MixDigest.String(),
		Nonce:            strconv.FormatUint(blockInfo.Nonce.Uint64(), 10),
		BaseFee:          blockInfo.BaseFee.String(),
		WithdrawalsHash:  blockInfo.WithdrawalsHash.String(),
		BlobGasUsed:      *blockInfo.BlobGasUsed,
		ExcessBlobGas:    *blockInfo.ExcessBlobGas,
	}
	return blockHeader, nil
}

func (s *EthNodeService) ListBlockHeaderByRange(ctx context.Context, param domain.BlockHeaderByRangeParam) ([]domain.BlockHeader, error) {
	startBlock := new(big.Int)
	endBlock := new(big.Int)
	startBlock.SetString(param.Start, 10)
	endBlock.SetString(param.End, 10)
	blockRange, err := s.ethClient.BlockHeadersByRange(ctx, startBlock, endBlock, uint(domain.EthereumChainId))
	if err != nil {
		log.Error("list block header range fail", "err", err)
		return nil, fmt.Errorf("list block header range fail: %w", err)
	}
	blockHeaderList := make([]domain.BlockHeader, 0, len(blockRange))
	for _, block := range blockRange {
		blockItem := domain.BlockHeader{
			ParentHash:       block.ParentHash.String(),
			UncleHash:        block.UncleHash.String(),
			CoinBase:         block.Coinbase.String(),
			Root:             block.Root.String(),
			TxHash:           block.TxHash.String(),
			ReceiptHash:      block.ReceiptHash.String(),
			ParentBeaconRoot: block.ParentBeaconRoot.String(),
			Difficulty:       block.Difficulty.String(),
			Number:           block.Number.String(),
			GasLimit:         block.GasLimit,
			GasUsed:          block.GasUsed,
			Time:             block.Time,
			Extra:            string(block.Extra),
			MixDigest:        block.MixDigest.String(),
			Nonce:            strconv.FormatUint(block.Nonce.Uint64(), 10),
			BaseFee:          block.BaseFee.String(),
			WithdrawalsHash:  block.WithdrawalsHash.String(),
			BlobGasUsed:      *block.BlobGasUsed,
			ExcessBlobGas:    *block.ExcessBlobGas,
		}
		blockHeaderList = append(blockHeaderList, blockItem)
	}
	return blockHeaderList, nil
}

func (s *EthNodeService) GetAccount(ctx context.Context, param domain.AccountParam) (domain.Account, error) {
	nonceResult, err := s.ethClient.TxCountByAddress(ctx, common.HexToAddress(param.Address))
	if err != nil {
		log.Error("get nonce by address fail", "err", err)
		return domain.Account{}, fmt.Errorf("get nonce by address fail: %w", err)
	}
	balanceResult, err := s.ethDataClient.GetBalanceByAddress(param.ContractAddress, param.Address)
	if err != nil {
		return domain.Account{}, fmt.Errorf("get balance by address fail: %w", err)
	}
	log.Info("balance result", "balance=", balanceResult.Balance, "balanceStr=", balanceResult.BalanceStr)
	balanceStr := "0"
	if balanceResult.Balance != nil && balanceResult.Balance.Int() != nil {
		balanceStr = balanceResult.Balance.Int().String()
	}
	sequence := strconv.FormatUint(uint64(nonceResult), 10)
	return domain.Account{
		Sequence: sequence,
		Balance:  balanceStr,
	}, nil
}

func (s *EthNodeService) GetFee(ctx context.Context, _ domain.FeeParam) (domain.Fee, error) {
	// 网络推荐的 gasPrice（适用于非 EIP-1559 的旧交易，单位为 wei）
	gasPrice, err := s.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		log.Error("get gas price failed", "err", err)
		return domain.Fee{}, fmt.Errorf("get gas price failed: %w", err)
	}
	// maxPriorityFeePerGas，即交易者愿意给矿工的小费（EIP-1559 的优先费用部分）
	gasTipCap, err := s.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		log.Error("get gas tip cap failed", "err", err)
		return domain.Fee{}, fmt.Errorf("get gas tip cap failed: %w", err)
	}
	/*
		按 | 分割，提取出 gasPrice 和 tipCap。
		*2 / *3 是标记「倍数提升」，用于快速交易加速。
	*/
	//return domain.Fee{
	//	SlowFee:   gasPrice.String() + "|" + gasTipCap.String(),
	//	NormalFee: gasPrice.String() + "|" + gasTipCap.String() + "|" + "*2",
	//	FastFee:   gasPrice.String() + "|" + gasTipCap.String() + "|" + "*3",
	//}, nil
	return domain.Fee{
		SlowFee: domain.GasFee{
			GasPrice:  gasPrice.String(),
			GasTipCap: gasTipCap.String(),
		},
		NormalFee: domain.GasFee{
			GasPrice:  gasPrice.String(),
			GasTipCap: gasTipCap.String(),
			MultiVal:  "2",
		},
		FastFee: domain.GasFee{
			GasPrice:  gasPrice.String(),
			GasTipCap: gasTipCap.String(),
			MultiVal:  "3",
		},
	}, nil
}

func (s *EthNodeService) SendTx(ctx context.Context, param domain.SendTxParam) (string, error) {
	transaction, err := s.ethClient.SendRawTransaction(ctx, param.RawTx)
	if err != nil {
		return "", fmt.Errorf("send transaction error: %w", err)
	}
	return transaction.String(), nil
}

func (s *EthNodeService) ListTxByAddress(_ context.Context, param domain.TxAddressParam) ([]domain.TxMessage, error) {
	var resp *types.TransactionResponse[types.AccountTxResponse]
	var err error
	if param.ContractAddress != "0x00" && param.ContractAddress != "" {
		resp, err = s.ethDataClient.GetTxByAddress(uint64(param.Page), uint64(param.PageSize), param.Address, "tokentx")
	} else {
		resp, err = s.ethDataClient.GetTxByAddress(uint64(param.Page), uint64(param.PageSize), param.Address, "txlist")
	}
	if err != nil {
		log.Error("get GetTxByAddress error", "err", err)
		return nil, fmt.Errorf("get GetTxByAddress error: %w", err)
	}
	txs := resp.TransactionList
	list := make([]domain.TxMessage, 0, len(txs))
	for i := 0; i < len(txs); i++ {
		var txStatus domain.TxStatus
		if txs[i].State == "1" {
			txStatus = domain.TxStatus_Success
		} else {
			txStatus = domain.TxStatus_Failed
		}
		list = append(list, domain.TxMessage{
			Hash:   txs[i].TxId,
			Tos:    []string{txs[i].To},
			Froms:  []string{txs[i].From},
			Fee:    txs[i].TxFee,
			Status: txStatus,
			Values: []string{txs[i].Amount},
			Type:   1,
			Height: txs[i].Height,
		})
	}
	return list, nil
}

// GetTxByHash 识别并解析 ERC20 标准的转账交易，提取出实际收款地址和金额，为后续统一交易结构封装打好基础
func (s *EthNodeService) GetTxByHash(ctx context.Context, param domain.GetTxByHashParam) (domain.TxMessage, error) {
	tx, err := s.ethClient.TxByHash(ctx, common.HexToHash(param.Hash))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return domain.TxMessage{}, errors.New("ethereum Tx not found")
		}
		log.Error("get transaction error", "err", err)
		return domain.TxMessage{}, fmt.Errorf("get transaction error: %w", err)
	}
	receipt, err := s.ethClient.TxReceiptByHash(ctx, common.HexToHash(param.Hash))
	if err != nil {
		log.Error("get transaction receipt error", "err", err)
		return domain.TxMessage{}, fmt.Errorf("get transaction receipt error: %w", err)
	}

	var beforeToAddress string
	var beforeTokenAddress string
	var beforeValue *big.Int

	code, err := s.ethClient.EthGetCode(ctx, common.HexToAddress(tx.To().String()))
	if err != nil {
		log.Info("Get account code fail", "err", err)
		return domain.TxMessage{}, fmt.Errorf("get account code fail: %w", err)
	}
	// 判断是否是代币转账
	/*
		第一步 判断目标地址是否是合约
		想判断 tx.To() 地址是否是合约地址。原因是：
		如果是合约地址，说明这笔交易是调用合约方法，即合约交互；
		ERC20 代币转账就是一种调用合约 transfer 方法的行为；
		如果不是合约地址，那一定是普通转账（ETH 直接转账）。
	*/
	if code == ContractTransfer {
		// 第二步 获取交易的 data 字段并 hex 编码（如：ERC20 的转账函数和参数）
		/*
				这一步是为了获取交易的“调用数据”，也就是 tx.Data()。
			    如果是合约调用，这里包含了：
			    函数选择器（前 4 字节）；
			    编码后的参数（通常是 32 字节对齐的 ABI 编码）
		*/
		inputData := hexutil.Encode(tx.Data()[:])
		//  第三步 判断是否是 transfer(address,uint256) 方法（ERC20 的标准转账方法）：
		/*
				inputData[:10] == "0xa9059cbb"
					判断方法 ID 是否为 transfer(address,uint256)；
					方法 ID 是函数签名 transfer(address,uint256) 经过 Keccak-256 哈希后的前 4 字节。
				len(inputData) >= 138
				因为：
					1 byte = 2 hex chars；
					函数选择器 + 两个参数共 4 + 32 + 32 = 68 字节；
					68 * 2 = 136 + 0x 前缀 = 138。
			    只有长度足够且前缀匹配 0xa9059cbb 才认为是标准的 ERC20 转账调用。
		*/
		if len(inputData) >= 138 && inputData[:10] == "0xa9059cbb" {
			// 第四步：提取转账地址（To）
			/*
					inputData[34:74] 提取的是第一个参数（address 类型）的位置：
						34 是从 0x 之后第 34 位（也就是 byte offset: 4+0 -> 第5字节开始）；
						地址总是 20 字节，也就是 40 hex 字符；
						因此 [34:74] 刚好取到 20 字节的地址字段（32 字节中的后 20 字节）；
				    address 在 ABI 编码中是一个 32 字节字段，左侧补 0。
						address: 0x5A1...123
						编码后为：
						0000000000000000000000005a1123...
			*/
			beforeToAddress = "0x" + inputData[34:74]
			// 第五步：提取转账金额（value）
			/*
				inputData[74:138] 是 value 字段（第二个参数）：
					从地址字段之后开始；
					也是 32 字节（64 hex）；
				使用 strings.TrimLeft(..., "0")：
					把左侧的 0 去掉，得到最简形式的 hex 值；
					比如 00000000000000000000000000000000000000000000000000000000000003e8（1000）；
					去掉前导 0 后得到 3e8。
				hexutil.DecodeBig 将十六进制字符串转换成 *big.Int。
			*/
			trimHex := strings.TrimLeft(inputData[74:138], "0")
			rawValue, _ := hexutil.DecodeBig("0x" + trimHex)
			//  第六步：设置最终值
			/*
				    beforeTokenAddress = tx.To().String()：
				        合约地址（即代币合约）是这笔交易的目标地址 tx.To()。
					beforeValue = decimal.NewFromBigInt(rawValue, 0).BigInt()：
						你先用 decimal 做中转，最终还是转成 *big.Int。
						实际上这里直接 beforeValue = rawValue 就足够了。
			*/
			beforeTokenAddress = tx.To().String()
			beforeValue = decimal.NewFromBigInt(rawValue, 0).BigInt()
		}
	} else {
		// 如果是普通转账（非合约）
		beforeToAddress = tx.To().String()
		beforeTokenAddress = common.Address{}.String()
		beforeValue = tx.Value()
	}
	var fromAddrs []string
	var toAddrs []string
	var valueList []string
	fromAddrs = append(fromAddrs, "")
	toAddrs = append(toAddrs, beforeToAddress)
	valueList = append(valueList, beforeValue.String())
	var txStatus domain.TxStatus
	if receipt.Status == 1 {
		txStatus = domain.TxStatus_Success
	} else {
		txStatus = domain.TxStatus_Failed
	}
	return domain.TxMessage{
		Hash:            tx.Hash().Hex(),
		Index:           uint32(receipt.TransactionIndex),
		Froms:           fromAddrs,
		Tos:             toAddrs,
		Values:          valueList,
		Fee:             tx.GasFeeCap().String(),
		Status:          txStatus,
		Type:            int32(tx.Type()),
		Height:          receipt.BlockNumber.String(),
		ContractAddress: beforeTokenAddress,
		Data:            hexutils.BytesToHex(tx.Data()),
	}, nil
}

func (s *EthNodeService) CreateUnSignTransaction(ctx context.Context, param domain.UnSignTransactionParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) BuildSignedTransaction(ctx context.Context, param domain.SignedTransactionParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) DecodeTransaction(ctx context.Context, param domain.DecodeTransactionParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) VerifySignedTransaction(ctx context.Context, param domain.VerifyTransactionParam) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetExtraData(ctx context.Context, param domain.ExtraDataParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewEthNodeService(ethClient evmbase.EVMClient, ethDataClient *evmbase.EthScan) service.WalletAccountService {
	return &EthNodeService{
		ethClient:     ethClient,
		ethDataClient: ethDataClient,
	}
}
