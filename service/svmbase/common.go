package svmbase

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/btcutil/base58"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gagliardetto/solana-go"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"sort"
	"strconv"
)

const (
	ProgramTokenMetadata = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
	ProgramCandyMachine  = "cndyAnrLdpQ5YwhpQdNceFMvx6bM2he7u3U4LVzGzjA"
	ProgramBubblegum     = "BGumetW1zi6dfL4nqJG1oD8T4PZ9FeZr4u8B7u4N1NYy"
	ProgramSPLToken      = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)

// parseTokenTransferAmount 从 base64 data 中解析出 amount
// SPL Token 金额获取方式对比如下：
// 来源	适用场景	是否推荐	说明
// PostTokenBalances - PreTokenBalances	通用方式，尤其适合主指令	✅ 推荐	可靠但不适用于 inner CPI 中没有列出的账户
// 指令 data 解析（如上）	innerInstruction CPI 中的 transfer	✅ 推荐	必须解析 base64 data 字节，指令 ID = 3，amount 为 little-endian uint64
func parseTokenTransferAmount(base64Data string) (uint64, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return 0, err
	}
	if len(dataBytes) < 9 {
		return 0, fmt.Errorf("invalid SPL Token transfer data length: %d", len(dataBytes))
	}

	// 指令类型应该是 3
	if dataBytes[0] != 3 {
		return 0, fmt.Errorf("not a transfer instruction, id = %d", dataBytes[0])
	}

	// 剩余 8 字节是 little-endian 的 uint64
	amount := binary.LittleEndian.Uint64(dataBytes[1:9])
	return amount, nil
}

// 解析 InnerInstructions 中的 CPI（跨程序调用）
/* CPI 是什么？
	在 Solana 中，每个合约是一个「程序（Program）」。
	当一个程序调用另一个程序，就形成了 CPI（跨程序调用），
	这些调用不会出现在主 Instructions 列表中，而是记录在 Meta.InnerInstructions 中。
举个例子：
	比如你调用一个 SPL Token 合约转账，主指令可能是某个合约（如 Candy Machine）调用，
	而转账的动作是由它内部调用 Token Program 触发的，那这个转账就会出现在 InnerInstructions 中。
*/
func processInnerInstructions(txResult *TransactionResult, tx *domain.TxMessage) {
	tx.Type = TypeContractCall.ToInt32()
	for _, item := range txResult.Meta.InnerInstructions {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		instructions, ok := entry["instructions"].([]interface{})
		if !ok {
			continue
		}

		for _, inst := range instructions {
			innerInst, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}

			// programIdIndex → 找到 ProgramId
			programIdIndex := int(innerInst["programIdIndex"].(float64))
			if programIdIndex >= len(txResult.Transaction.Message.AccountKeys) {
				continue
			}
			programId := txResult.Transaction.Message.AccountKeys[programIdIndex]
			if programId != "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA" {
				continue
			}

			// 解析账户（from, to）
			accounts := innerInst["accounts"].([]interface{})
			if len(accounts) < 2 {
				continue
			}
			toIndex := int(accounts[1].(float64))
			if toIndex >= len(txResult.Transaction.Message.AccountKeys) {
				continue
			}
			toAddr := txResult.Transaction.Message.AccountKeys[toIndex]

			// 金额解析
			dataStr, ok := innerInst["data"].(string)
			if !ok {
				continue
			}
			amount, err := parseTokenTransferAmount(dataStr)
			if err != nil {
				log.Error("failed to parse token transfer amount:", err)
				continue
			}

			tx.Tos = append(tx.Tos, toAddr)
			tx.Values = append(tx.Values, strconv.FormatUint(amount, 10))
		}
	}
}

// isLikelyMetaplexNF Metaplex NFT 地址识别
func isLikelyMetaplexNFT(mint string) bool {
	// 根据 Metaplex Metadata PDA 派生规则：
	// PDA = ["metadata", METADATA_PROGRAM_ID, mint]
	// 实际上我们可以通过程序 ID 是否存在 metadata program 来推断

	// ✅ 可以改成调用链上合约/缓存元数据的方式
	// 这里只是一个简单启发式：mint 地址为 32 长度 base58 的字符串
	return len(mint) > 0 && len(mint) <= 44 // base58 字符串长度判断
}

// ProcessTokenTransfers 处理 SPL Token 转账（含 NFT）
func ProcessTokenTransfers(txResult *TransactionResult, tx *domain.TxMessage) {
	for _, post := range txResult.Meta.PostTokenBalances {
		postToken, ok := post.(map[string]interface{})
		if !ok {
			continue
		}

		accountIndex := int(postToken["accountIndex"].(float64))
		if accountIndex >= len(txResult.Meta.PreTokenBalances) {
			continue
		}

		preToken := txResult.Meta.PreTokenBalances[accountIndex].(map[string]interface{})

		// amount difference
		postAmount := postToken["uiTokenAmount"].(map[string]interface{})["amount"].(string)
		preAmount := preToken["uiTokenAmount"].(map[string]interface{})["amount"].(string)

		postAmt, _ := strconv.ParseUint(postAmount, 10, 64)
		preAmt, _ := strconv.ParseUint(preAmount, 10, 64)

		if postAmt > preAmt {
			// 是接收方
			owner := postToken["owner"].(string)
			mint := postToken["mint"].(string)

			// 识别 NFT（decimals == 0，amount == 1）
			decimals := int(postToken["uiTokenAmount"].(map[string]interface{})["decimals"].(float64))
			isNFT := decimals == 0 && postAmt-preAmt == 1
			if isNFT {
				tx.Tos = append(tx.Tos, owner)
				tx.ContractAddress = mint
				tx.Type = TypeNftTransfer.ToInt32()
			} else {
				tx.Tos = append(tx.Tos, owner)
				tx.ContractAddress = mint
				tx.Type = TypeSplTransfer.ToInt32()
				tx.Values = append(tx.Values, strconv.FormatUint(postAmt-preAmt, 10))
			}
		}
	}
}

// ProcessInstructions 遍历一笔 Solana 交易的所有指令（Instructions），
// 从中识别系统原生转账指令（Program ID 为 "111111..."），
// 提取接收方地址和转账金额，并保存到 tx 对象中。
/*
	func ProcessInstructions(txResult *TransactionResult, tx *domain.TxMessage) error {
		// 原始 system transfer 解析
		for i, inst := range txResult.Transaction.Message.Instructions {
			if inst.ProgramIdIndex >= len(txResult.Transaction.Message.AccountKeys) {
				log.Warn("Invalid program ID index", "instruction", i)
				continue
			}

			programId := txResult.Transaction.Message.AccountKeys[inst.ProgramIdIndex]
			switch programId {
			case "11111111111111111111111111111111":
				// 系统转账
				if len(inst.Accounts) < 2 {
					log.Warn("Invalid accounts length", "instruction", i)
					continue
				}
				toIndex := inst.Accounts[1]
				if toIndex >= len(txResult.Transaction.Message.AccountKeys) {
					log.Warn("Invalid to account index", "instruction", i)
					continue
				}
				toAddr := txResult.Transaction.Message.AccountKeys[toIndex]
				tx.Tos = append(tx.Tos, toAddr)

				if err := calculateAmount(txResult, toIndex, tx); err != nil {
					log.Warn("Failed to calculate amount", "error", err)
					continue
				}
			case "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA":
				// SPL Token 处理（可扩展具体类型解析）
				continue
			}
		}

		// 扩展处理逻辑
		processTokenTransfers(txResult, tx)
		processInnerInstructions(txResult, tx)

		return nil
	}
*/
func ProcessInstructions(txResult *TransactionResult, tx *domain.TxMessage) error {
	return processSOLTransfers(txResult, tx)
}

func processSOLTransfers(txResult *TransactionResult, tx *domain.TxMessage) error {
	tx.Type = TypeSolTransfer.ToInt32()
	for i, inst := range txResult.Transaction.Message.Instructions {
		if inst.ProgramIdIndex >= len(txResult.Transaction.Message.AccountKeys) {
			log.Warn("Invalid program ID index", "instruction", i)
			continue
		}

		// 筛选系统转账指令（SystemProgram）
		// 仅处理 ProgramID 为 1111... 的系统程序（System Program）指令，即 SOL 原生转账（不处理 SPL Token 或合约调用）。
		if txResult.Transaction.Message.AccountKeys[inst.ProgramIdIndex] != "11111111111111111111111111111111" {
			continue
		}

		// 检查参数是否合法（系统转账指令至少应包含 2 个账户地址，from 和 to）。
		if len(inst.Accounts) < 2 {
			log.Warn("Invalid accounts length", "instruction", i)
			continue
		}
		toIndex := inst.Accounts[1]
		if toIndex >= len(txResult.Transaction.Message.AccountKeys) {
			log.Warn("Invalid to account index", "instruction", i)
			continue
		}
		// 通过索引获取接收方地址。
		// 将接收地址加入 tx.Tos 切片。
		toAddr := txResult.Transaction.Message.AccountKeys[toIndex]
		tx.Tos = append(tx.Tos, toAddr)

		// 调用辅助函数计算转账金额
		if err := calculateAmount(txResult, toIndex, tx); err != nil {
			log.Warn("Failed to calculate amount", "error", err)
			continue
		}
	}
	return nil
}

// calculateAmount
// 根据 preBalance 和 postBalance 差值，计算转账金额并写入 tx.Values。
func calculateAmount(txResult *TransactionResult, toIndex int, tx *domain.TxMessage) error {
	// 避免越界：校验该账户在 pre/post balance 中存在。
	if toIndex >= len(txResult.Meta.PostBalances) || toIndex >= len(txResult.Meta.PreBalances) {
		return fmt.Errorf("invalid balance index: %d", toIndex)
	}

	amount := txResult.Meta.PostBalances[toIndex] - txResult.Meta.PreBalances[toIndex]
	tx.Values = append(tx.Values, strconv.FormatUint(amount, 10))

	return nil
}

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
