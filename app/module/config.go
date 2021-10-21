package module


type ApiConfig struct {
	Mexc Mexc `yaml:"mexc"`
}

type Mexc struct {
	AK string `yaml:"access_key"`
	AS string `yaml:"access_secret"`
}


type Context struct {
	ApiConfig ApiConfig
	IsDev bool
}