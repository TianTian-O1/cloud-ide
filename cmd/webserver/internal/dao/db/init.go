package db

import (
	"context"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
)

var sqlDb *sqlx.DB

func InitMysql() error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	var err error
	db, err := sqlx.ConnectContext(timeoutCtx, "mysql", conf.MysqlConfig.DataSourceName)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(int(conf.MysqlConfig.MaxOpenConns))
	db.SetMaxIdleConns(int(conf.MysqlConfig.MaxIdleConns))

	sqlDb = db

	return nil
}

func CloseMysql() {
	sqlDb.Close()
}

func DB() *sqlx.DB {
	return sqlDb
}
