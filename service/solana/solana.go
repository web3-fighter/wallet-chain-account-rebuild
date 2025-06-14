package solana

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/web3-fighter/chain-explorer-api/types"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
	"github.com/web3-fighter/wallet-chain-account/service/donoting"
	"github.com/web3-fighter/wallet-chain-account/service/svmbase"
	"strconv"
)

const ChainName = "Solana"

const (
	MaxBlockRange = 1000
)

var _ service.WalletAccountService = (*SOLNodeService)(nil)

type SOLNodeService struct {
	svmClient svmbase.SVMClient
	sdkClient svmbase.JSONRpc
	solData   *svmbase.SolData
	donoting.DoNotingService
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

func (s *SOLNodeService) CreateUnSignTransaction(ctx context.Context, param domain.UnSignTransactionParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) BuildSignedTransaction(ctx context.Context, param domain.SignedTransactionParam) (domain.SignedTransaction, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) DecodeTransaction(ctx context.Context, param domain.DecodeTransactionParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) VerifySignedTransaction(ctx context.Context, param domain.VerifyTransactionParam) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) GetExtraData(ctx context.Context, param domain.ExtraDataParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func NewSOLNodeService(svmClient svmbase.SVMClient, sdkClient svmbase.JSONRpc, solData *svmbase.SolData) service.WalletAccountService {
	return &SOLNodeService{
		sdkClient: sdkClient,
		svmClient: svmClient,
		solData:   solData,
	}
}
