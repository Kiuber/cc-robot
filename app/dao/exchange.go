package dao

import "time"

type ExchangeSymbolPair struct {
	ExchangeName string
	SymbolPair   string
	Symbol1      string
	Symbol2      string
	CreatedAt    time.Time `gorm:"column:ctime"`
	UpdatedAt    time.Time `gorm:"column:mtime"`
}

func (ExchangeSymbolPair) TableName() string {
	return "s_cc_exchange_symbol_pair"
}
