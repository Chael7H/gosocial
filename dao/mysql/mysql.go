package mysql

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	//_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gosocial/models"
	"gosocial/settings"
)

// 声明一个全局变量，类型为*gorm.DB
var db *gorm.DB

// Init 初始化mysql连接
func Init(cfg *settings.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.PassWord,
		cfg.Host,
		cfg.Port,
		cfg.Dbname,
	)
	// 使用GORM的Open方法连接数据库
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		zap.L().Error("connect mysql failed", zap.Error(err))
		return
	}
	// 自动迁移模型（创建表或更新表结构）
	err = db.AutoMigrate(
		&models.User{},       // 用户模型
		&models.Friendship{}, // 好友模型
		&models.Message{},    // 消息模型
		&models.Post{},       // 动态模型
	)
	if err != nil {
		zap.L().Error("自动迁移失败", zap.Error(err))
		return
	}
	// 设置最大空闲连接数和最大打开连接数
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Error("failed to get underlying sql.DB", zap.Error(err))
		return
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	return
}

func GetDB() *gorm.DB {
	return db
}

func SetDB(dbTest *gorm.DB) {
	db = dbTest
}
