package model

type GlobalVariable struct {
	Env    string
	IsDev  bool
	Config Config
}

type Config struct {
	Infra Infra
	Api   Api
}

type Infra struct {
	MySQLConfig MySQLConfig `yaml:"mysql"`
	RedisConfig RedisConfig `yaml:"redis"`
	EventNU     EventNU     `yaml:"event_notify"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

type EventNU struct {
	URL string `yaml:"url"`
}

const (
	EventNotifyGroupDev  string = "internal"
	EventNotifyGroupProd string = "all"
)

type Api struct {
	Mexc Mexc `yaml:"mexc"`
}

type Mexc struct {
	BaseURL    string `yaml:"base_url"`
	AK         string `yaml:"access_key"`
	AS         string `yaml:"access_secret"`
	AllowTrade bool   `yaml:"allow_trade"`
}
