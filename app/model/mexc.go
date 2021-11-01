package model

type MexcResp struct {
	Code	int
	Data	interface{}
}

type MexcAPIData struct {
	OK	bool
	Msg string
	Payload interface{}
	RawPayload interface{}
}

type SupportSymbolPair struct {
	SymbolPairList []string `mapstructure:"symbol"`
	SymbolPairMap map[string][]string
	Exchange string
}

type AppearSymbolPair struct {
	SymbolPair string
	Symbol1And2 []string
	Exchange string
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
	Symbol string `json:"symbol"`
	Volume string `json:"volume"`
	High string `json:"high"`
	Low string `json:"low"`
	Bid string `json:"bid"`
	Ask string `json:"ask"`
	Open string `json:"open"`
	Last string `json:"last"`
	Time int64 `json:"time"`
	ChangeRate string `json:"change_rate"`
}

type DepthInfo struct {
	Asks []struct {
		Price string `json:"price"`
		Quantity string `json:"quantity"`
	} `json:"asks"`
	Bids []struct {
		Price string `json:"price"`
		Quantity string `json:"quantity"`
	} `json:"bids"`
}

type Order struct {
	SymbolPair string `json:"symbol"`
	Price string `json:"price"`
	Quantity string `json:"quantity"`
	TradeType string `json:"trade_type"`
	OrderType string `json:"order_type"`
	ClientOrderId string `json:"client_order_id"`
}

type AccountInfo map[string]struct {
	Frozen    string `json:"frozen"`
	Available string `json:"available"`
}
