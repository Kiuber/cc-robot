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

type SupportSymbols struct {
	Symbols []string `mapstructure:"symbol"`
}

