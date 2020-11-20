package test

import (
	"time"
)

type MysqlTestConf struct {
	DataBase        string
	Addr            string
	User            string
	Password        string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifeTime time.Duration
	ConnTimeOut     time.Duration
	WriteTimeOut    time.Duration
	ReadTimeOut     time.Duration
	LogMode         bool
}


