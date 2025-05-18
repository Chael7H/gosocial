package main

import (
	"fmt"
	"go.uber.org/zap"
	"gosocial/controllers"
	"gosocial/dao/mysql"
	"gosocial/dao/redis"
	"gosocial/logger"
	"gosocial/pkg/snowflake"
	"gosocial/routes"
	"gosocial/settings"
)

// @title 社交网络平台API
// @version 1.0
// @description	基于 Gin 和 GORM 构建的高性能社交网络平台接口文档
// @termsOfService http://swagger.io/terms/

// @contact.name API 支持团队
// @contact.url http://www.swagger.io/support
// @contact.email caribdis@163.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1

func main() {
	//1.加载配置
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//1.1初始化雪花算法
	if err := snowflake.Init(settings.Conf.StartTime, settings.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}
	//1.2初始化gin框架内置校验器使用的翻译器
	if err := controllers.InitTrans("zh"); err != nil {
		fmt.Printf("init translator failed, err:%v\n", err)
		return
	}

	//2.初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer zap.L().Sync()
	zap.L().Debug("logger init success!")

	//3.初始化mysql
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	//4.初始化redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()
	//5.注册路由
	r := routes.Init()
	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		return
	}
}
