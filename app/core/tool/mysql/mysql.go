package mysql

import (
	cboot "cc-robot/core/boot"
	clog "cc-robot/core/tool/log"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var client *gorm.DB

func Setup() {
	var dbURI string
	var dialector gorm.Dialector
	dbURI = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cboot.GV.Config.Infra.MySQLConfig.User,
		cboot.GV.Config.Infra.MySQLConfig.Password,
		cboot.GV.Config.Infra.MySQLConfig.Host,
		cboot.GV.Config.Infra.MySQLConfig.Port,
		cboot.GV.Config.Infra.MySQLConfig.Name)
	config := mysql.Config{
		DSN:                       dbURI, // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}
	dialector = mysql.New(config)

	logger := clog.EventLog().With(zap.Reflect("mysql config", config))

	conn, err := gorm.Open(dialector, &gorm.Config{})
	logger.Info("start connect")
	if err != nil {
		logger.Error(err.Error())
	}
	sqlDB, err := conn.DB()
	if err != nil {
		logger.Error("connect MySQLClient server failed.")
	}
	sqlDB.SetMaxIdleConns(10)                   // SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxOpenConns(100)                  // SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetConnMaxLifetime(time.Second * 600) // SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	client = conn
}

func MySQLClient() *gorm.DB {
	if client == nil {
		Setup()
	}

	sqlDB, err := client.DB()
	if err != nil {
		clog.EventLog().Error("connect MySQLClient server failed.")
		Setup()
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		Setup()
	}

	return client
}
