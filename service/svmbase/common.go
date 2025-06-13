package svmbase

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/btcutil/base58"
	"github.com/gagliardetto/solana-go"
	"sort"
	"strconv"
)

const (
	ProgramTokenMetadata = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
	ProgramCandyMachine  = "cndyAnrLdpQ5YwhpQdNceFMvx6bM2he7u3U4LVzGzjA"
	ProgramBubblegum     = "BGumetW1zi6dfL4nqJG1oD8T4PZ9FeZr4u8B7u4N1NYy"
	ProgramSPLToken      = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)

// GetSuggestedPriorityFee 从一组交易优先费 PrioritizationFee 中，
// 计算出建议使用的优先费（SuggestedPriorityFee）值，
// 返回第 75 百分位数（也就是排在后 25% 的第一个值）作为建议优先费。
// 这样可以有效排除极端低值，提升交易打包成功率，同时避免选择太高的费率浪费成本。
func GetSuggestedPriorityFee(fees []*PrioritizationFee) uint64 {
	if len(fees) == 0 {
		return 0
	}

	// 创建一个和 fees 等长的 uint64 切片 priorityFees。
	//把每个结构体中的 PrioritizationFee 提取出来，填入 priorityFees 中。
	priorityFees := make([]uint64, len(fees))
	for i, fee := range fees {
		priorityFees[i] = fee.PrioritizationFee
	}

	// 对 priorityFees 切片进行升序排序。
	sort.Slice(priorityFees, func(i, j int) bool {
		return priorityFees[i] < priorityFees[j]
	})

	// 计算 第 75 百分位（P75） 的下标：
	// 例如有 10 个值，那么 index = 10 * 0.75 = 7（返回第 7 个，注意 Go 是从 0 开始的索引）。
	// 返回排序后的 priorityFees 中第 75 百分位的值，作为建议优先费。
	index := int(float64(len(priorityFees)) * 0.75)
	return priorityFees[index]
}

type ParsedTransfer struct {
	From         string
	To           string
	TokenAddress string
	Amount       string
}

type AccountPair struct {
	From string
	To   string
}

// SplTransfer SPL Token 转账解析（优先级高于系统转账）
type SplTransfer struct {
	From, To, TokenAddress, ContractWallet, Amount string
}

/*
NftTransfer
在 Solana 生态中，NFT（非同质化代币）的标准主要有以下几种，了解这些标准对于解析 NFT 铸造、转移、展示等行为非常关键：

✅ 主要 NFT 标准（协议）
	Metaplex Token Metadata (SPL Metadata)	最主流的 Solana NFT 标准，NFT 都会附加 metadata 账户，支持图片、动画、描述等	✅ 广泛采用
	Metaplex Candy Machine v1 / v2	是 NFT 铸造机制，使用者部署 Candy Machine 程序合约，批量铸造 NFT	✅ 已广泛部署（v2为主）
	Compressed NFTs (cNFTs)	通过 Merkle Tree 技术，极大降低 NFT 铸造和存储成本，由 Metaplex 提出	✅ 高效，越来越多项目采用
	Programmable NFTs (pNFTs)	新一代 NFT 规范，支持锁定、授权、自定义规则等复杂控制逻辑	✅ 新兴标准
	Edition NFT (Master Edition, Print Edition)	类似 ERC-1155 的“印刷版 NFT”，一个主 NFT 拥有多个 Edition	✅ 用于限量 NFT
	Creator Standard	支持验证创作者信息（creators 列表 + verified 字段）	✅ 所有 Metaplex NFT 使用

🔎 辅助结构
	Metadata Account
		所有标准 NFT 都依赖一个和 NFT mint 相关联的 Metadata PDA（程序派生地址），用于描述名称、symbol、uri、创作者等信息。
	Master Edition Account
		指定为主 NFT 的数据结构，可生成印刷版（Edition）NFT。
	Use Authority Record / Token Record
		在 Programmable NFT 中用于记录授权者、使用情况、访问控制。

🔥 常见 Program ID（合约地址）
	Token Metadata	metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s	Metaplex Metadata 主合约
	Candy Machine v2	cndyAnrLdpQ5YwhpQdNceFMvx6bM2he7u3U4LVzGzjA	NFT 批量铸造合约
	Bubblegum（Compressed NFT）	BGumetW1zi6dfL4nqJG1oD8T4PZ9FeZr4u8B7u4N1NYy	Bubblegum 合约
	Token Program (SPL)	TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA	SPL Token 转账合约

🧠 小结
Solana NFT 标准可以归类为：
	元数据标准	Metaplex Token Metadata
	铸造机制	Candy Machine v1/v2
	存储优化	Compressed NFT (Bubblegum)
	功能拓展	Programmable NFTs
	创作者控制	Verified Creator / Royalties


Program ID 与 NFT 标准的对应关系
Program ID	名称	功能/标准	类型	是否代表一种 NFT 标准
metaqbxx...bt518x1s	Token Metadata	Metaplex NFT 元数据标准	✅ 标准合约	✅ 是（核心）
cndyAnrL...LVzGzjA	Candy Machine v2	批量铸造 NFT	NFT 铸造工具	✅ 是
BGumetW1...uN1NYy	Bubblegum	Compressed NFT（压缩 NFT）	新型存储机制	✅ 是
TokenkegQ...23VQ5DA	SPL Token Program	SPL Token 转账、持有	通用代币合约	✅ 可支持 NFT（但非唯一）
mplTokenMetadata（同上）	Programmable NFT (pNFT)	支持锁定、授权等逻辑的复杂 NFT	pNFT 标准合约	✅ 是
Sysvar / BPFLoader	系统合约	调度器、加载器等	系统功能	❌ 否

*/
// NftTransfer NFT 转账解析：检查 Metaplex 指令
type NftTransfer struct {
	From, To, Mint, ContractWallet string
}

// ExtractAccounts 提取主要账户信息（假设账户0为发送者，1为接收者）
func ExtractAccounts(tx TransactionDetail) AccountPair {
	msg, ok := tx.Message.(map[string]interface{})
	if !ok {
		return AccountPair{}
	}
	accts, ok := msg["accountKeys"].([]interface{})
	if !ok || len(accts) < 2 {
		return AccountPair{}
	}
	from := fmt.Sprintf("%v", accts[0])
	to := fmt.Sprintf("%v", accts[1])
	return AccountPair{From: from, To: to}
}

func ParseSPLTransfer(tx TransactionDetail) *SplTransfer {
	for _, inner := range tx.Meta.InnerInstructions {
		m, ok := inner.(map[string]interface{})
		if !ok {
			continue
		}
		if insList, ok := m["instructions"].([]interface{}); ok {
			for _, rawIns := range insList {
				ins, ok := rawIns.(map[string]interface{})
				if !ok || ins["programName"] == nil {
					continue
				}
				if ins["programName"] != "spl-token" {
					continue
				}
				parsed, ok := ins["parsed"].(map[string]interface{})
				if !ok || parsed["type"] != "transfer" && parsed["type"] != "transferChecked" {
					continue
				}
				info := parsed["info"].(map[string]interface{})
				return &SplTransfer{
					From:           info["source"].(string),
					To:             info["destination"].(string),
					TokenAddress:   info["mint"].(string),
					ContractWallet: info["authority"].(string),
					Amount:         fmt.Sprintf("%v", info["amount"]),
				}
			}
		}
	}
	return nil
}

func ParseNFTTransfer(tx TransactionDetail) *NftTransfer {
	for _, ins := range tx.Meta.InnerInstructions {
		m, ok := ins.(map[string]interface{})
		if !ok {
			continue
		}
		if insList, ok := m["instructions"].([]interface{}); ok {
			for _, rawIns := range insList {
				instr, ok := rawIns.(map[string]interface{})
				if !ok {
					continue
				}
				pidIndex := int(instr["programIdIndex"].(float64))
				msg, _ := tx.Message.(map[string]interface{})
				if acct, ok := msg["accountKeys"].([]interface{}); ok {
					programID := acct[pidIndex].(string)
					transfer := &NftTransfer{}
					switch programID {
					case ProgramTokenMetadata:
						transfer.ContractWallet = ProgramTokenMetadata
					case ProgramCandyMachine:
						transfer.ContractWallet = ProgramCandyMachine
					case ProgramBubblegum:
						transfer.ContractWallet = ProgramBubblegum
					case ProgramSPLToken:
						transfer.ContractWallet = ProgramSPLToken
					default:
						continue
					}

					instrParsed, ok := instr["parsed"].(map[string]interface{})
					if !ok || instrParsed["type"] != "transfer" {
						continue
					}
					info := instrParsed["info"].(map[string]interface{})
					transfer.From = fmt.Sprintf("%v", info["owner"])
					transfer.To = fmt.Sprintf("%v", info["destination"])
					transfer.Mint = fmt.Sprintf("%v", info["mint"])
					return transfer
				}
			}
		}
	}
	return nil
}

// 解码 SPL 转账金额（Data 是 base64 编码）
func decodeSPLTransferAmount(data string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("decode base64 failed: %w", err)
	}

	// spl-token transfer 的第一个字节是指令编号（0x03），后面8字节是金额（小端）
	if len(raw) < 9 || raw[0] != 3 {
		return "", fmt.Errorf("not a valid SPL transfer")
	}
	amount := binary.LittleEndian.Uint64(raw[1:9])
	return strconv.FormatUint(amount, 10), nil
}

func generateKeyPair() (*ed25519.PrivateKey, *ed25519.PublicKey, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, "", err
	}
	address := base58.Encode(publicKey)
	return &privateKey, &publicKey, address, nil
}

func PrivateKeyHexToPrivateKey(privateKeyHex string) (*ed25519.PrivateKey, error) {
	privKeyByteList, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}
	privKey := ed25519.PrivateKey(privKeyByteList)
	return &privKey, nil
}

func PrivateKeyToPubKey(privateKey *ed25519.PrivateKey) (*ed25519.PublicKey, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	pubKey := (*privateKey).Public().(ed25519.PublicKey)
	return &pubKey, nil
}

func PrivateKeyHexToPubKey(privateKeyHex string) (*ed25519.PublicKey, error) {
	privKey, err := PrivateKeyHexToPrivateKey(privateKeyHex)
	if err != nil {
		return nil, err
	}
	return PrivateKeyToPubKey(privKey)
}

func PubKeyHexToPubKey(publicKeyHex string) (*ed25519.PublicKey, error) {
	pubKeyByteList, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key hex: %w", err)
	}
	pubKey := ed25519.PublicKey(pubKeyByteList)
	return &pubKey, nil
}

func PubKeyToPubKeyHex(publicKey *ed25519.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key is nil")
	}
	return hex.EncodeToString(*publicKey), nil
}

func PubKeyToAddress(publicKey *ed25519.PublicKey) (string, error) {
	if publicKey == nil {
		return "", fmt.Errorf("public key is nil")
	}
	return base58.Encode(*publicKey), nil
}

func PubKeyHexToAddress(publicKeyHex string) (string, error) {
	pubKey, err := PubKeyHexToPubKey(publicKeyHex)
	if err != nil {
		return "", err
	}
	return PubKeyToAddress(pubKey)
}

func GenerateNewKeypair() (*solana.PrivateKey, solana.PublicKey) {
	account := solana.NewWallet()
	return &account.PrivateKey, account.PublicKey()
}

func PrivateKeyFromByteList(privateKeyByteList []byte) (*solana.PrivateKey, error) {
	if len(privateKeyByteList) != 64 {
		return nil, fmt.Errorf("invalid private key length")
	}
	privateKey := solana.PrivateKey(privateKeyByteList)
	return &privateKey, nil
}

func PrivateKeyFromHex(privateKeyHex string) (*solana.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode hex error: %w", err)
	}
	return PrivateKeyFromByteList(privateKeyBytes)
}

func PrivateKeyFromBase58(privateKeyBase58 string) (*solana.PrivateKey, error) {
	privateKey, err := solana.PrivateKeyFromBase58(privateKeyBase58)
	if err != nil {
		return nil, fmt.Errorf("create private key from base58 error: %w", err)
	}
	return &privateKey, nil
}

func PrivateKeyToBase58(privateKey *solana.PrivateKey) string {
	return privateKey.String()
}

func PublicKeyFromPrivateKey(privateKey *solana.PrivateKey) solana.PublicKey {
	return privateKey.PublicKey()
}

func PublicKeyFromBase58(publicKeyBase58 string) (solana.PublicKey, error) {
	publicKey, err := solana.PublicKeyFromBase58(publicKeyBase58)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("create public key error: %w", err)
	}
	return publicKey, nil
}

func PublicKeyToBase58(publicKey solana.PublicKey) string {
	return publicKey.String()
}

func AddressFromPubKey(publicKey solana.PublicKey) string {
	return publicKey.String()
}
