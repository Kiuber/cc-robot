package dao

import "time"

type ExchangeSymbol struct {
	ExchangeName string
	Symbol       string
	Symbol1 string
	Symbol2 string
	CreatedAt time.Time `gorm:"column:ctime"`
	UpdatedAt time.Time `gorm:"column:mtime"`
}

func (ExchangeSymbol) TableName() string {
	return "s_cc_exchange_symbol"
}