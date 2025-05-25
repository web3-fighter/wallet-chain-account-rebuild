package donoting

import (
	"context"
	"github.com/web3-fighter/wallet-chain-account/domain"
	"github.com/web3-fighter/wallet-chain-account/service"
)

var _ service.WalletAccountService = (*DoNotingService)(nil)

type DoNotingService struct{}

func (s *DoNotingService) GetSupportChains(ctx context.Context, param domain.SupportChainsParam) (bool, error) {
	return true, nil
}

func (s *DoNotingService) ConvertAddress(ctx context.Context, param domain.ConvertAddressParam) (string, error) {
	return "", nil
}

func (s *DoNotingService) ValidAddress(ctx context.Context, param domain.ValidAddressParam) (bool, error) {
	return true, nil
}

func (s *DoNotingService) GetBlockByNumber(ctx context.Context, param domain.BlockNumberParam) (domain.Block, error) {
	return domain.Block{}, nil
}

func (s *DoNotingService) GetBlockByHash(ctx context.Context, param domain.BlockHashParam) (domain.Block, error) {
	return domain.Block{}, nil
}

func (s *DoNotingService) GetBlockHeaderByHash(ctx context.Context, param domain.BlockHeaderHashParam) (domain.BlockHeader, error) {
	return domain.BlockHeader{}, nil
}

func (s *DoNotingService) ListBlockHeaderByRange(ctx context.Context, param domain.BlockHeaderByRangeParam) ([]domain.BlockHeader, error) {
	return nil, nil
}

func (s *DoNotingService) GetAccount(ctx context.Context, param domain.AccountParam) (domain.Account, error) {
	return domain.Account{}, nil
}

func (s *DoNotingService) GetFee(ctx context.Context, param domain.FeeParam) (domain.Fee, error) {
	return domain.Fee{}, nil
}

func (s *DoNotingService) SendTx(ctx context.Context, param domain.SendTxParam) (string, error) {
	return "", nil
}

func (s *DoNotingService) ListTxByAddress(ctx context.Context, param domain.TxAddressParam) ([]domain.TxMessage, error) {
	return nil, nil
}

func (s *DoNotingService) GetTxByHash(ctx context.Context, param domain.GetTxByHashParam) (domain.TxMessage, error) {
	return domain.TxMessage{}, nil
}

func (s *DoNotingService) CreateUnSignTransaction(ctx context.Context, param domain.UnSignTransactionParam) (string, error) {
	return "", nil
}

func (s *DoNotingService) BuildSignedTransaction(ctx context.Context, param domain.SignedTransactionParam) (domain.SignedTransaction, error) {
	return domain.SignedTransaction{}, nil
}

func (s *DoNotingService) DecodeTransaction(ctx context.Context, param domain.DecodeTransactionParam) (string, error) {
	return "", nil
}

func (s *DoNotingService) VerifySignedTransaction(ctx context.Context, param domain.VerifyTransactionParam) (bool, error) {
	return true, nil
}

func (s *DoNotingService) GetExtraData(ctx context.Context, param domain.ExtraDataParam) (string, error) {
	return "", nil
}
