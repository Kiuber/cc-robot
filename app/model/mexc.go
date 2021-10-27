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

