package evmbase

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/web3-fighter/chain-explorer-api/client/etherscan"
	"github.com/web3-fighter/chain-explorer-api/types"
	"time"
)

type EthScan struct {
	EthDataCli *etherscan.ChainExplorerClient
}

func NewEthDataClient(baseUrl, apiKey string, timeout time.Duration) (*EthScan, error) {
	etherscanCli, err := etherscan.NewChainExplorerClient(apiKey, baseUrl, false, timeout)
	if err != nil {
		log.Error("New etherscan client fail", "err", err)
		return nil, err
	}
	return &EthScan{EthDataCli: etherscanCli}, err
}

func (ed *EthScan) GetTxByAddress(page, pagesize uint64, address string, action types.ActionType) (*types.TransactionResponse[types.AccountTxResponse], error) {
	request := &types.AccountTxRequest{
		PageRequest: types.PageRequest{
			Page:  page,
			Limit: pagesize,
		},
		Action:  action,
		Address: address,
	}
	txData, err := ed.EthDataCli.GetTxByAddress(request)
	if err != nil {
		return nil, err
	}
	return txData, nil
}

func (ed *EthScan) GetBalanceByAddress(contractAddr, address string) (*types.AccountBalanceResponse, error) {
	accountItem := []string{address}
	symbol := []string{"ETH"}
	contractAddress := []string{contractAddr}
	protocolType := []string{""}
	page := []string{"1"}
	limit := []string{"10"}
	acbr := &types.AccountBalanceRequest{
		ChainShortName:  "ETH",
		ExplorerName:    "etherescan",
		Account:         accountItem,
		Symbol:          symbol,
		ContractAddress: contractAddress,
		ProtocolType:    protocolType,
		Page:            page,
		Limit:           limit,
	}
	etherscanResp, err := ed.EthDataCli.GetAccountBalance(acbr)
	if err != nil {
		log.Error("get account balance error", "err", err)
		return nil, err
	}
	return etherscanResp, nil
}
