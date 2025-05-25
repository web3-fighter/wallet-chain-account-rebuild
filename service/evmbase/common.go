package evmbase

import (
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

func CreateLegacyUnSignTx(txData *types.LegacyTx, chainId *big.Int) string {
	tx := types.NewTx(txData)
	signer := types.LatestSignerForChainID(chainId)
	txHash := signer.Hash(tx)
	return txHash.String()
}

func BuildErc721Data(fromAddress, toAddress common.Address, tokenId *big.Int) []byte {
	var data []byte

	transferFnSignature := []byte("safeTransferFrom(address,address,uint256)")
	hash := crypto.Keccak256Hash(transferFnSignature)
	methodId := hash[:4]

	dataFromAddress := common.LeftPadBytes(fromAddress.Bytes(), 32)
	dataToAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	dataTokenId := common.LeftPadBytes(tokenId.Bytes(), 32)

	data = append(data, methodId...)
	data = append(data, dataFromAddress...)
	data = append(data, dataToAddress...)
	data = append(data, dataTokenId...)

	return data
}

// BuildErc20Data 构造一个标准的 ERC20 transfer(address,uint256) 方法的调用数据（即交易中的 data 字段）， 用于发起代币转账交易。
func BuildErc20Data(toAddress common.Address, amount *big.Int) []byte {
	// 定义一个 data 字节切片用于存放最终生成的 ABI 编码数据。
	/*
		ABI（Application Binary Interface）编码，是智能合约调用和交互的“底层语言”，即：
			将合约函数调用和参数转换为二进制格式（字节数组），以便发送到链上供 EVM 执行的标准规则。
			ABI 编码是构造交易中的 data 字段的关键，它告诉以太坊：
				要调用哪个合约函数？
				传了哪些参数？参数的顺序、类型、值是什么？
	*/
	var data []byte
	// 表示要调用合约的 transfer(address,uint256) 方法。 这个是标准 ERC20 的转账函数签名。
	transferFnSignature := []byte("transfer(address,uint256)")
	// 将函数签名 transfer(address,uint256) 做 Keccak256 哈希
	hash := crypto.Keccak256Hash(transferFnSignature)
	// 取前 4 个字节，得到 methodId，即 方法选择器（method selector）
	// 对于 ERC20 的 transfer 方法，其结果就是： 0xa9059cbb，这告诉以太坊虚拟机要调用哪个方法
	methodId := hash[:4]
	// 将接收地址编码成 32 字节（右对齐，左补 0）
	// 因为以太坊 ABI 规定地址作为参数是 uint160（20字节），但每个参数在 ABI 编码中都占 32 字节
	dataAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	// 把转账的 token 数额（big.Int）编码为 32 字节格式。
	dataAmount := common.LeftPadBytes(amount.Bytes(), 32)
	// 0xa9059cbb                                           <-- transfer method
	//0000000000000000000000001234567890abcdef1234567890abcdef12345678 <-- address
	//0000000000000000000000000000000000000000000000000de0b6b3a7640000  <-- amount

	// 这个字节数组就是智能合约交易的 tx.Data 部分。
	data = append(data, methodId...)
	data = append(data, dataAddress...)
	data = append(data, dataAmount...)

	return data
}

// CreateEip1559UnSignTx 创建一个 EIP-1559 格式的未签名交易，并返回该未签名交易的哈希值
/*
	构造一笔 EIP-1559 交易（DynamicFeeTx）
	获取它的 待签名哈希（这个函数做的事 ）
	将这个哈希发给硬件钱包、冷钱包、HSM 等去做签名
	再将签名后的完整交易发送上链
*/
func CreateEip1559UnSignTx(txData *types.DynamicFeeTx, chainId *big.Int) (string, error) {
	/*
		创建一个统一的 types.Tx 实例，它是对所有类型交易（LegacyTx、AccessListTx、DynamicFeeTx）的一层封装。
		你传入的是 DynamicFeeTx，所以最终得到的是 EIP-1559 类型的交易。
	*/
	tx := types.NewTx(txData)
	// 创建一个签名器（Signer），LatestSignerForChainID 会根据链 ID 自动选择正确的签名规则（比如 EIP-1559 的 LondonSigner）。
	signer := types.LatestSignerForChainID(chainId)
	// 使用 signer 对交易 tx 计算签名哈希（也叫 待签名哈希），这是在离线签名时要用私钥签名的哈希值。
	txHash := signer.Hash(tx)
	return txHash.String(), nil
}

// CreateEip1559SignedTx 将离线签名后的 EIP-1559 交易组装成可广播的签名交易 并返回：
/*
	这是函数签名，输入：
		txData: 构造好的 EIP-1559 交易结构
		signature: 离线签名后生成的签名（65 字节：r + s + v）
		chainId: 链 ID，用于确定签名规则
	返回：
		Signer: 用于后续签名验证的 signer 对象
		*types.Transaction: 已签名的交易
		string: 已签名交易的 RLP 编码（hex 字符串）
		string: 交易哈希（txHash）
		error: 错误信息
*/
func CreateEip1559SignedTx(txData *types.DynamicFeeTx, signature []byte, chainId *big.Int) (types.Signer, *types.Transaction, string, string, error) {
	// 将 DynamicFeeTx 包装成 *types.Transaction，这是通用的交易结构体，用于签名和广播。
	tx := types.NewTx(txData)
	// 根据链 ID 获取当前使用的签名规则，主网上返回 LondonSigner，用于支持 EIP-1559。
	signer := types.LatestSignerForChainID(chainId)
	// 把已经生成好的签名 signature 应用到交易 tx 上，得到 signedTx。
	// WithSignature 会验证签名格式并将签名字段写入交易对象。
	// 如果失败，说明签名格式有误或和 signer 不兼容。
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, nil, "", "", errors.New("tx with signature fail")
	}
	// 将已签名交易进行 RLP 编码，生成字节流，这就是可以用 eth_sendRawTransaction 广播的原始交易数据。
	signedTxData, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return nil, nil, "", "", errors.New("encode tx to byte fail")
	}
	return signer, signedTx, "0x" + hex.EncodeToString(signedTxData)[4:], signedTx.Hash().String(), nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	return rpc.BlockNumber(number.Int64()).String()
}

func toFilterArg(q ethereum.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{"address": q.Addresses, "topics": q.Topics}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != nil || q.ToBlock != nil {
			return nil, errors.New("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == nil {
			arg["fromBlock"] = "0x0"
		} else {
			arg["fromBlock"] = toBlockNumArg(q.FromBlock)
		}
		arg["toBlock"] = toBlockNumArg(q.ToBlock)
	}
	return arg, nil
}
