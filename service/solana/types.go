package solana

import (
	"encoding/hex"
	"fmt"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service/svmbase"
	"sort"
	"strconv"
	"strings"
)

func organizeTransactionsByBlock(txResults []*svmbase.TransactionResult) ([]domain.BlockHeader, error) {
	if len(txResults) == 0 {
		return nil, nil
	}

	blockMap := make(map[uint64]domain.BlockHeader)

	for _, txResult := range txResults {
		if txResult == nil {
			continue
		}

		slot := txResult.Slot

		block, exists := blockMap[slot]
		if !exists {
			block = domain.BlockHeader{
				Number: strconv.FormatUint(slot, 10),
			}

			if txResult.BlockTime != nil {
				block.Time = uint64(*txResult.BlockTime)
			}

			if len(txResult.Transaction.Signatures) > 0 {
				block.Hash = txResult.Transaction.Signatures[0]
			}

			txHashes := make([]string, 0)
			for _, sig := range txResult.Transaction.Signatures {
				txHashes = append(txHashes, sig)
			}
			block.TxHash = strings.Join(txHashes, ",")

			block.GasUsed = txResult.Meta.ComputeUnitsConsumed

			blockMap[slot] = block
		} else {
			if len(txResult.Transaction.Signatures) > 0 {
				if block.TxHash != "" {
					block.TxHash += "," + txResult.Transaction.Signatures[0]
				} else {
					block.TxHash = txResult.Transaction.Signatures[0]
				}
			}

			block.GasUsed += txResult.Meta.ComputeUnitsConsumed
		}
	}

	blocks := make([]domain.BlockHeader, 0, len(blockMap))
	for _, block := range blockMap {
		blocks = append(blocks, block)
	}

	sort.Slice(blocks, func(i, j int) bool {
		heightI, _ := strconv.ParseUint(blocks[i].Number, 10, 64)
		heightJ, _ := strconv.ParseUint(blocks[j].Number, 10, 64)
		return heightI < heightJ
	})

	return blocks, nil
}

func validateBlockRangeRequest(param domain.BlockHeaderByRangeParam) error {
	startSlot, err := strconv.ParseUint(param.Start, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid start height format: %s", err)
	}
	endSlot, err := strconv.ParseUint(param.End, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid end height format: %s", err)
	}

	if startSlot > endSlot {
		return fmt.Errorf("invalid height range: start height greater than end height")
	}

	if startSlot-endSlot > MaxBlockRange {
		return fmt.Errorf("invalid range: exceeds maximum allowed range of %d", MaxBlockRange)
	}

	if ok, msg := validateChainAndNetwork(param.Chain, param.Network); !ok {
		return fmt.Errorf("invalid chain or network: %s", msg)
	}

	return nil
}

func validatePublicKey(pubKey string) (bool, string) {
	if pubKey == "" {
		return false, "public key cannot be empty"
	}
	pubKeyWithoutPrefix := strings.TrimPrefix(pubKey, "0x")

	if len(pubKeyWithoutPrefix) != 64 {
		return false, "invalid public key length"
	}
	if _, err := hex.DecodeString(pubKeyWithoutPrefix); err != nil {
		return false, "invalid public key format: must be hex string"
	}

	return true, ""
}

func validateChainAndNetwork(chain, network string) (bool, string) {
	if chain != ChainName {
		return false, "invalid chain"
	}
	//if network != NetworkMainnet && network != NetworkTestnet {
	//	return false, "invalid network"
	//}
	return true, ""
}

func parseBlockTransaction(tx svmbase.TransactionDetail) *domain.BlockTransaction {
	bt := &domain.BlockTransaction{Hash: tx.Signature, Height: tx.Slot}

	// 跳过失败或无 Meta 的交易
	if tx.Meta == nil || tx.Meta.Err != nil {
		return bt
	}

	accounts := svmbase.ExtractAccounts(tx)

	// 默认为系统转账
	bt.From, bt.To = accounts.From, accounts.To
	bt.ContractWallet = "" // 默认为系统
	bt.Amount = "0"

	// 检查是否是 SPL Token 转账
	if parsed := svmbase.ParseSPLTransfer(tx); parsed != nil {
		bt.From = parsed.From
		bt.To = parsed.To
		bt.TokenAddress = parsed.TokenAddress
		bt.ContractWallet = parsed.ContractWallet
		bt.Amount = parsed.Amount
	}

	// 再检查是否是 NFT（Metaplex 标准）
	if nft := svmbase.ParseNFTTransfer(tx); nft != nil {
		/*
			contract_wallet 填写为 Token Metadata Program ID，代表该 NFT 操作属于元数据合约。
			TokenAddress 即 NFT 的 mint 地址；
			amount 固定为 "1"；
			如果交易同时包含多个类型（例如系统 & token），目前优先按 SPL Token/NFT 识别。你也可按业务需求调整顺序。
		*/
		bt.From = nft.From
		bt.To = nft.To
		bt.TokenAddress = nft.Mint
		bt.ContractWallet = nft.ContractWallet // 常量：Token Metadata Program ID
		bt.Amount = "1"
	}

	return bt
}
