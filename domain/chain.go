package domain

type CommonParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin          string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network       string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
}

type ExtraDataParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Address       string `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
	Coin          string `protobuf:"bytes,5,opt,name=coin,proto3" json:"coin,omitempty"`
}

type VerifyTransactionParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	PublicKey     string `protobuf:"bytes,4,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Signature     string `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
	RawTx         string `protobuf:"bytes,4,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
}

type DecodeTransactionParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	RawTx         string `protobuf:"bytes,4,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
}

type SignedTransactionParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Base64Tx      string `protobuf:"bytes,4,opt,name=base64_tx,json=base64Tx,proto3" json:"base64_tx,omitempty"`
	Signature     string `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
	PublicKey     string `protobuf:"bytes,6,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

type SignedTransaction struct {
	TxHash   string `json:"tx_hash,omitempty"`
	SignedTx string `json:"signed_tx,omitempty"`
}

type UnSignTransactionParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Base64Tx      string `protobuf:"bytes,4,opt,name=base64_tx,json=base64Tx,proto3" json:"base64_tx,omitempty"`
}

type GetTxByHashParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin          string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network       string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	Hash          string `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
}

type TxAddressParam struct {
	ConsumerToken   string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain           string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin            string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network         string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	Address         string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	ContractAddress string `protobuf:"bytes,6,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	Page            uint32 `protobuf:"varint,7,opt,name=page,proto3" json:"page,omitempty"`
	PageSize        uint32 `protobuf:"varint,8,opt,name=pagesize,proto3" json:"pagesize,omitempty"`
	Cursor          string `protobuf:"bytes,9,opt,name=cursor,proto3" json:"cursor,omitempty"`
}

type TxMessage struct {
	Hash            string   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Index           uint32   `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
	Froms           []string `protobuf:"bytes,3,rep,name=froms,proto3" json:"froms,omitempty"`
	Tos             []string `protobuf:"bytes,4,rep,name=tos,proto3" json:"tos,omitempty"`
	Values          []string `protobuf:"bytes,7,rep,name=values,proto3" json:"values,omitempty"`
	Fee             string   `protobuf:"bytes,5,opt,name=fee,proto3" json:"fee,omitempty"`
	Status          TxStatus `protobuf:"varint,6,opt,name=status,proto3,enum=dapplink.account.TxStatus" json:"status,omitempty"`
	Type            int32    `protobuf:"varint,8,opt,name=type,proto3" json:"type,omitempty"`
	Height          string   `protobuf:"bytes,9,opt,name=height,proto3" json:"height,omitempty"`
	ContractAddress string   `protobuf:"bytes,10,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	Datetime        string   `protobuf:"bytes,11,opt,name=datetime,proto3" json:"datetime,omitempty"`
	Data            string   `protobuf:"bytes,12,opt,name=data,proto3" json:"data,omitempty"`
}

type SendTxParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin          string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network       string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	RawTx         string `protobuf:"bytes,5,opt,name=raw_tx,json=rawTx,proto3" json:"raw_tx,omitempty"`
}

type FeeParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin          string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network       string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	RawTx         string `protobuf:"bytes,5,opt,name=rawTx,proto3" json:"rawTx,omitempty"`
	Address       string `protobuf:"bytes,6,opt,name=address,proto3" json:"address,omitempty"`
}

type Fee struct {
	SlowFee   GasFee `protobuf:"bytes,3,opt,name=slow_fee,json=slowFee,proto3" json:"slow_fee,omitempty"`
	NormalFee GasFee `protobuf:"bytes,4,opt,name=normal_fee,json=normalFee,proto3" json:"normal_fee,omitempty"`
	FastFee   GasFee `protobuf:"bytes,5,opt,name=fast_fee,json=fastFee,proto3" json:"fast_fee,omitempty"`
}

type GasFee struct {
	GasPrice  string
	GasTipCap string
	MultiVal  string
}

type AccountParam struct {
	ConsumerToken    string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain            string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Coin             string `protobuf:"bytes,3,opt,name=coin,proto3" json:"coin,omitempty"`
	Network          string `protobuf:"bytes,4,opt,name=network,proto3" json:"network,omitempty"`
	Address          string `protobuf:"bytes,5,opt,name=address,proto3" json:"address,omitempty"`
	ContractAddress  string `protobuf:"bytes,6,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
	ProposerKeyIndex uint64 `protobuf:"varint,7,opt,name=proposer_key_index,json=proposerKeyIndex,proto3" json:"proposer_key_index,omitempty"`
}

type Account struct {
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	AccountNumber string `protobuf:"bytes,4,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Sequence      string `protobuf:"bytes,5,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Balance       string `protobuf:"bytes,6,opt,name=balance,proto3" json:"balance,omitempty"`
}

type BlockHeaderByRangeParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Start         string `protobuf:"bytes,4,opt,name=start,proto3" json:"start,omitempty"`
	End           string `protobuf:"bytes,5,opt,name=end,proto3" json:"end,omitempty"`
}

type BlockHeader struct {
	Hash             string `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	ParentHash       string `protobuf:"bytes,2,opt,name=parent_hash,json=parentHash,proto3" json:"parent_hash,omitempty"`
	UncleHash        string `protobuf:"bytes,3,opt,name=uncle_hash,json=uncleHash,proto3" json:"uncle_hash,omitempty"`
	CoinBase         string `protobuf:"bytes,4,opt,name=coin_base,json=coinBase,proto3" json:"coin_base,omitempty"`
	Root             string `protobuf:"bytes,5,opt,name=root,proto3" json:"root,omitempty"`
	TxHash           string `protobuf:"bytes,6,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	ReceiptHash      string `protobuf:"bytes,7,opt,name=receipt_hash,json=receiptHash,proto3" json:"receipt_hash,omitempty"`
	ParentBeaconRoot string `protobuf:"bytes,8,opt,name=parent_beacon_root,json=parentBeaconRoot,proto3" json:"parent_beacon_root,omitempty"`
	Difficulty       string `protobuf:"bytes,9,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
	Number           string `protobuf:"bytes,10,opt,name=number,proto3" json:"number,omitempty"`
	GasLimit         uint64 `protobuf:"varint,11,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	GasUsed          uint64 `protobuf:"varint,12,opt,name=gas_used,json=gasUsed,proto3" json:"gas_used,omitempty"`
	Time             uint64 `protobuf:"varint,13,opt,name=time,proto3" json:"time,omitempty"`
	Extra            string `protobuf:"bytes,14,opt,name=extra,proto3" json:"extra,omitempty"`
	MixDigest        string `protobuf:"bytes,15,opt,name=mix_digest,json=mixDigest,proto3" json:"mix_digest,omitempty"`
	Nonce            string `protobuf:"bytes,16,opt,name=nonce,proto3" json:"nonce,omitempty"`
	BaseFee          string `protobuf:"bytes,17,opt,name=base_fee,json=baseFee,proto3" json:"base_fee,omitempty"`
	WithdrawalsHash  string `protobuf:"bytes,18,opt,name=withdrawals_hash,json=withdrawalsHash,proto3" json:"withdrawals_hash,omitempty"`
	BlobGasUsed      uint64 `protobuf:"varint,19,opt,name=blob_gas_used,json=blobGasUsed,proto3" json:"blob_gas_used,omitempty"`
	ExcessBlobGas    uint64 `protobuf:"varint,20,opt,name=excess_blob_gas,json=excessBlobGas,proto3" json:"excess_blob_gas,omitempty"`
}

type BlockHeaderNumberParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Height        int64  `protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

type BlockHeaderHashParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Hash          string `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
}

type BlockHashParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Hash          string `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	ViewTx        bool   `protobuf:"varint,4,opt,name=view_tx,json=viewTx,proto3" json:"view_tx,omitempty"`
}

type BlockNumberParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Height        int64  `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	ViewTx        bool   `protobuf:"varint,4,opt,name=view_tx,json=viewTx,proto3" json:"view_tx,omitempty"`
}

type Block struct {
	Height       int64               `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Hash         string              `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	BaseFee      string              `protobuf:"bytes,5,opt,name=base_fee,json=baseFee,proto3" json:"base_fee,omitempty"`
	Transactions []*BlockTransaction `protobuf:"bytes,6,rep,name=transactions,proto3" json:"transactions,omitempty"`
}

type BlockTransaction struct {
	From           string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To             string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	TokenAddress   string `protobuf:"bytes,3,opt,name=token_address,json=tokenAddress,proto3" json:"token_address,omitempty"`
	ContractWallet string `protobuf:"bytes,4,opt,name=contract_wallet,json=contractWallet,proto3" json:"contract_wallet,omitempty"`
	Hash           string `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	Height         uint64 `protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
	Amount         string `protobuf:"bytes,7,opt,name=amount,proto3" json:"amount,omitempty"`
}

type ValidAddressParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Address       string `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
}

type SupportChainsParam struct {
	ConsumerToken string `json:"consumer_token,omitempty"`
	Chain         string `json:"chain,omitempty"`
	Network       string `json:"network,omitempty"`
}

type ConvertAddressParam struct {
	ConsumerToken string `protobuf:"bytes,1,opt,name=consumer_token,json=consumerToken,proto3" json:"consumer_token,omitempty"`
	Chain         string `protobuf:"bytes,2,opt,name=chain,proto3" json:"chain,omitempty"`
	Network       string `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Type          string `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	PublicKey     string `protobuf:"bytes,5,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
}

type TxStatus int32

const (
	TxStatus_NotFound              TxStatus = iota
	TxStatus_Pending               TxStatus = 1
	TxStatus_Failed                TxStatus = 2
	TxStatus_Success               TxStatus = 3
	TxStatus_ContractExecuteFailed TxStatus = 4
	TxStatus_Other                 TxStatus = 5
)

const (
	ChainIdRedisKey     = "chainId:%s"
	ChainId             = "chainId"
	ChainName           = "chainName"
	Polygon             = "Polygon"
	GormInfoFmt         = "%s\n[%.3fms] [rows:%v] %s"
	ZeroAddress         = "0x0000000000000000000000000000000000000000"
	WEthAddress         = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	EthAddress          = "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	SepoliaWETH         = ""
	LogTimeFormat       = "2006-01-02"
	LayerTypeOne        = 1
	LayerTypeTwo        = 2
	StakingTypeStake    = 1
	StakingTypeReward   = 2
	StakingTypeWithdraw = 3

	BridgeOperaInitType     = 1
	BridgeOperaFinalizeType = 2

	ScrollChainId          uint64 = 534352
	PolygonChainId         uint64 = 1101
	PolygonSepoliaChainId  uint64 = 1442
	EthereumChainId        uint64 = 1
	EthereumSepoliaChainId uint64 = 11155111
	BaseChainId            uint64 = 8453
	BaseSepoliaChainId     uint64 = 84532
	MantaChainId           uint64 = 169
	MantaSepoliaChainId    uint64 = 3441006
	MantleSepoliaChainId   uint64 = 5003
	MantleChainId          uint64 = 5000
	ZkFairSepoliaChainId   uint64 = 43851
	ZkFairChainId          uint64 = 42766
	OkxSepoliaChainId      uint64 = 195
	OkxChainId             uint64 = 66
	OpChinId               uint64 = 10
	OpTestChinId           uint64 = 11155420
	LineaChainId           uint64 = 59144
	BlocksLimit                   = 10000
)
