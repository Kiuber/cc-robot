package dao

import "time"

type ExchangeSymbolPair struct {
	ExchangeName  string
	SymbolPair    string
	Symbol1       string
	Symbol2       string
	OpenTimestamp int
	CreatedAt     time.Time `gorm:"column:ctime"`
	UpdatedAt     time.Time `gorm:"column:mtime"`
}

func (ExchangeSymbolPair) TableName() string {
	return "s_cc_exchange_symbol_pair"
}

type ExchangePrimeConfig struct {
	ExchangeName string
	SymbolPair   string
	Status       string
	CreatedAt    time.Time `gorm:"column:ctime"`
	UpdatedAt    time.Time `gorm:"column:mtime"`
}

func (ExchangePrimeConfig) TableName() string {
	return "s_cc_exchange_prime_config"
}
