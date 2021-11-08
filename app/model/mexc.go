package model

import (
	"math/big"
)

type MexcResp struct {
	Code int
	Data interface{}
	Msg  string
}

type MexcAPIData struct {
	OK         bool
	Msg        string
	Payload    interface{}
	RawPayload interface{}
}

type MockAPIData map[string]struct {
	APIPath string
	Data    string
}

type SupportSymbolPair struct {
	SymbolPairList []string `mapstructure:"symbol"`
	SymbolPairMap  map[string][]string
	Exchange       string
}

type SymbolPairInfo struct {
	SymbolPair string
	WebLink    string
	OpenTime   string `mapstructure:"openTime"`
}

type AppearSymbolPair struct {
	SymbolPair   string
	Symbol1And2  []string
	ExchangeName string
}

type SymbolPairBetterPrice struct {
	AppearSymbolPair AppearSymbolPair
	LowestOfAskPrice *big.Float
}

type SymbolPairInfoList []struct {
	Symbol          string `mapstructure:"symbol"`
	State           string `mapstructure:"state"`
	PriceScale      int    `mapstructure:"price_scale"`
	QuantityScale   int    `mapstructure:"quantity_scale"`
	MinAmount       string `mapstructure:"min_amount"`
	MaxAmount       string `mapstructure:"max_amount"`
	MakerFeeRate    string `mapstructure:"maker_fee_rate"`
	TakerFeeRate    string `mapstructure:"taker_fee_rate"`
	Limited         bool   `mapstructure:"limited"`
	EtfMark         int    `mapstructure:"etf_mark"`
	SymbolPartition string `mapstructure:"symbol_partition"`
}

type SymbolPairTickerInfo []struct {
	SymbolPair string `mapstructure:"symbol"`
	Volume     string `mapstructure:"volume"`
	High       string `mapstructure:"high"`
	Low        string `mapstructure:"low"`
	Bid        string `mapstructure:"bid"`
	Ask        string `mapstructure:"ask"`
	Open       string `mapstructure:"open"`
	Last       string `mapstructure:"last"`
	Time       int64  `mapstructure:"time"`
	ChangeRate string `mapstructure:"change_rate"`
}

type DepthInfo struct {
	Asks []struct {
		Price    string `mapstructure:"price"`
		Quantity string `mapstructure:"quantity"`
	} `json:"asks"`
	Bids []struct {
		Price    string `mapstructure:"price"`
		Quantity string `mapstructure:"quantity"`
	} `json:"bids"`
}

type Order struct {
	SymbolPair    string `json:"symbol"`
	Price         string `json:"price"`
	Quantity      string `json:"quantity"`
	TradeType     string `json:"trade_type"`
	OrderType     string `json:"order_type"`
	ClientOrderId string `json:"client_order_id"`
}

type OrderList []struct {
	ID           string `mapstructure:"id"`
	SymbolPair   string `mapstructure:"symbol"`
	Price        string `mapstructure:"price"`
	Quantity     string `mapstructure:"quantity"`
	State        string `mapstructure:"state"`
	Type         string `mapstructure:"type"`
	DealQuantity string `mapstructure:"deal_quantity"`
	DealCost     string `mapstructure:"deal_amount"`
	CreateTime   int64  `mapstructure:"create_time"`
}

type AccountInfo map[string]struct {
	Frozen    string `mapstructure:"frozen"`
	Available string `mapstructure:"available"`
}

type SymbolPairConf struct {
	BidCost            *big.Float
	ExpectedProfitRate *big.Float
}
