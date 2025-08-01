package main

import (
	"fmt"
	"syscall"

	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/db"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/dao/rdis"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/routes"
	"github.com/mangohow/cloud-ide/pkg/httpserver"
	"github.com/mangohow/cloud-ide/pkg/logger"
	"github.com/mangohow/cloud-ide/pkg/router"
)

func main() {
	// 初始化配置
	if err := conf.LoadConf(); err != nil {
		panic(fmt.Errorf("load conf failed, reason:%s", err.Error()))
	}

	// 初始化日志
	if err := logger.InitLogger(conf.LoggerConfig); err != nil {
		panic(fmt.Errorf("init logger error, reason:%v", err))
	}

	// 初始化数据库
	if err := db.InitMysql(); err != nil {
		panic(fmt.Errorf("init mysql failed, reason:%s", err.Error()))
	}

	// 创建gin路由
	engine := router.NewGinRouter(conf.ServerConfig.Mode)
	// 注册路由
	routes.Register(engine)

	// 创建http server
	server := httpserver.NewServer(conf.ServerConfig.Host, conf.ServerConfig.Port, engine)

	// 启动server
	httpserver.ListenAndServe(server)

	fmt.Println("pid:", syscall.Getpid())

	// 等待服务退出
	httpserver.WaitForShutdown(server, func() {
		db.CloseMysql()
		if conf.EmailConfig.Enabled {
			rdis.CloseRedisConn()
		}
	})
}
