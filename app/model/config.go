package model


type Context struct {
	IsDev bool
	Config Config
}

type Config struct {
	Infra Infra
	Api   Api
}

type Infra struct {
	MySQLConfig MySQLConfig `yaml:"mysql"`
	RedisConfig RedisConfig `yaml:"redis"`
}

type MySQLConfig struct {
	Host	string `yaml:"host"`
	Port	string `yaml:"port"`
	User	string `yaml:"user"`
	Password	string `yaml:"password"`
	Name	string `yaml:"name"`
}

type RedisConfig struct {
	Host	string `yaml:"host"`
	Port	string `yaml:"port"`
	Password	string `yaml:"password"`
}


type Api struct {
	Mexc Mexc `yaml:"mexc"`
}

type Mexc struct {
	AK string `yaml:"access_key"`
	AS string `yaml:"access_secret"`
}