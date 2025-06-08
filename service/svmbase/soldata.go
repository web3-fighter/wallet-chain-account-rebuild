package svmbase

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/web3-fighter/chain-explorer-api/client/solscan"
	"github.com/web3-fighter/chain-explorer-api/types"
	"time"
)

type SolData struct {
	SolDataCli *solscan.ChainExplorerClient
}

func (ss *SolData) GetTxByAddress(page, pagesize uint64, address string, action types.ActionType) (*types.TransactionResponse[types.AccountTxResponse], error) {
	request := &types.AccountTxRequest{
		PageRequest: types.PageRequest{
			Page:  page,
			Limit: pagesize,
		},
		Action:  action,
		Address: address,
	}
	txData, err := ss.SolDataCli.GetTxByAddress(request)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

func NewSolScanClient(baseUrl, apiKey string, timeout time.Duration) (*SolData, error) {
	solCli, err := solscan.NewChainExplorerClient(apiKey, baseUrl, false, time.Duration(timeout))
	if err != nil {
		log.Error("New solscan client fail", "err", err)
		return nil, err
	}
	return &SolData{SolDataCli: solCli}, err
}
