package ethereum

import "github.com/web3-fighter/wallet-chain-account/service/evmbase"

const ChainName = "Ethereum"

type EthClient struct {
	ethClient     evmbase.EVMClient
	ethDataClient *evmbase.EthData
}
