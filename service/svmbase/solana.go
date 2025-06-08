package svmbase

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
	"github.com/web3-fighter/wallet-chain-account/service/donoting"
	"strings"
)

const ChainName = "Solana"

const (
	MaxBlockRange = 1000
)

var _ service.WalletAccountService = (*SOLNodeService)(nil)

type SOLNodeService struct {
	svmClient SVMClient
	sdkClient JSONRpc
	solData   *SolData
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
	accountAddress, err := PubKeyHexToAddress(pubKeyHex)
	if err != nil {
		err = fmt.Errorf("ConvertAddress PubKeyHexToAddress failed: %w", err)
		log.Error("err", err)
		return "", err
	}
	return accountAddress, nil

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
		latestSlot, err := s.svmClient.GetSlot(ctx, Finalized)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetSlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		resultSlot = latestSlot
	}

	blockResult := &BlockResult{}
	if param.ViewTx {
		tempBlockBySlot, err := s.svmClient.GetBlockBySlot(ctx, resultSlot, Signatures)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetBlockBySlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		blockResult = tempBlockBySlot
	} else {
		tempBlockBySlot, err := s.svmClient.GetBlockBySlot(ctx, resultSlot, None)
		if err != nil {
			err = fmt.Errorf("GetBlockByNumber GetBlockBySlot failed: %w", err)
			log.Error("err", err)
			return response, err
		}
		blockResult = tempBlockBySlot
	}

	response.Hash = blockResult.BlockHash
	response.Height = int64(resultSlot)
	if param.ViewTx {
		response.Transactions = make([]*domain.BlockTransaction, 0, len(blockResult.Signatures))
		for _, signature := range blockResult.Signatures {
			txInfo := &domain.BlockTransaction{
				Hash: signature,
			}
			response.Transactions = append(response.Transactions, txInfo)
		}
	}
	return response, nil
}

func (s *SOLNodeService) GetBlockByHash(ctx context.Context, param domain.BlockHashParam) (domain.Block, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) GetBlockHeaderByHash(ctx context.Context, param domain.BlockHeaderHashParam) (domain.BlockHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) ListBlockHeaderByRange(ctx context.Context, param domain.BlockHeaderByRangeParam) ([]domain.BlockHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) GetAccount(ctx context.Context, param domain.AccountParam) (domain.Account, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) GetFee(ctx context.Context, param domain.FeeParam) (domain.Fee, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) SendTx(ctx context.Context, param domain.SendTxParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) ListTxByAddress(ctx context.Context, param domain.TxAddressParam) ([]domain.TxMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (s *SOLNodeService) GetTxByHash(ctx context.Context, param domain.GetTxByHashParam) (domain.TxMessage, error) {
	//TODO implement me
	panic("implement me")
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

func NewSOLNodeService(svmClient SVMClient, sdkClient JSONRpc, solData *SolData) service.WalletAccountService {
	return &SOLNodeService{
		sdkClient: sdkClient,
		svmClient: svmClient,
		solData:   solData,
	}
}
