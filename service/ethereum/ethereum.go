package ethereum

import (
	"context"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
	"github.com/web3-fighter/wallet-chain-account/service/evmbase"
)

const ChainName = "Ethereum"

var _ service.WalletAccountService = (*EthNodeService)(nil)

type EthNodeService struct {
	ethClient     evmbase.EVMClient
	ethDataClient *evmbase.EthScan
}

func (s *EthNodeService) GetSupportChains(ctx context.Context, param domain.SupportChainsParam) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) ConvertAddress(ctx context.Context, param domain.ConvertAddressParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) ValidAddress(ctx context.Context, param domain.ValidAddressParam) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetBlockByNumber(ctx context.Context, param domain.BlockNumberParam) (domain.BlockResult, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetBlockByHash(ctx context.Context, param domain.BlockHashParam) (domain.BlockResult, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetBlockHeaderByHash(ctx context.Context, param domain.BlockHeaderHashParam) (domain.BlockHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) ListBlockHeaderByRange(ctx context.Context, param domain.BlockHeaderByRangeParam) ([]domain.BlockHeader, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetAccount(ctx context.Context, param domain.AccountParam) (domain.Account, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetFee(ctx context.Context, param domain.FeeParam) (domain.Fee, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) SendTx(ctx context.Context, param domain.SendTxParam) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) ListTxByAddress(ctx context.Context, param domain.TxAddressParam) ([]domain.TxMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (s *EthNodeService) GetTxByHash(ctx context.Context, param domain.GetTxByHashParam) (domain.TxMessage, error) {
	//TODO implement me
	panic("implement me")
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
