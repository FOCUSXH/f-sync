package db

import (
	"database/sql"
	"fmt"
	"fsync/server/global"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB 根据配置初始化数据库连接
func InitDB() error {
	// 构建不包含数据库名的DSN用于连接MySQL服务器
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=%t&loc=Local",
		global.Configs.Database.Username,
		global.Configs.Database.Password,
		global.Configs.Database.Host,
		global.Configs.Database.Port,
		global.Configs.Database.Charset,
		global.Configs.Database.ParseTime,
	)

	// 先尝试连接到MySQL服务器
	sqlDB, err := sql.Open("mysql", baseDSN)
	if err != nil {
		global.Logger.Error("连接MySQL服务器失败", zap.Error(err))
		return fmt.Errorf("连接MySQL服务器失败: %w", err)
	}
	defer sqlDB.Close()

	// 检查数据库是否存在
	dbName := global.Configs.Database.Name
	var exists int
	query := "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?"
	err = sqlDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		global.Logger.Error("查询数据库是否存在失败", zap.Error(err))
		return fmt.Errorf("查询数据库是否存在失败: %w", err)
	}

	// 如果数据库不存在，则创建数据库
	if exists == 0 {
		global.Logger.Info("数据库不存在，正在创建数据库", zap.String("database", dbName))
		createQuery := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET %s COLLATE %s_general_ci",
			dbName, global.Configs.Database.Charset, global.Configs.Database.Charset)
		_, err = sqlDB.Exec(createQuery)
		if err != nil {
			global.Logger.Error("创建数据库失败", zap.Error(err))
			return fmt.Errorf("创建数据库失败: %w", err)
		}
		global.Logger.Info("数据库创建成功", zap.String("database", dbName))
	} else {
		global.Logger.Info("数据库已存在", zap.String("database", dbName))
	}

	// 构建包含数据库名的DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		global.Configs.Database.Username,
		global.Configs.Database.Password,
		global.Configs.Database.Host,
		global.Configs.Database.Port,
		global.Configs.Database.Name,
		global.Configs.Database.Charset,
		global.Configs.Database.ParseTime,
	)

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		global.Logger.Error("连接数据库失败", zap.Error(err))
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err = db.DB()
	if err != nil {
		global.Logger.Error("获取数据库连接池失败", zap.Error(err))
		return fmt.Errorf("获取数据库连接池失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(global.Configs.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(global.Configs.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(global.Configs.Database.ConnMaxLifetime)

	global.Logger.Info("已配置连接池",
		zap.Int("max_open_conns", global.Configs.Database.MaxOpenConns),
		zap.Int("max_idle_conns", global.Configs.Database.MaxIdleConns))

	// 将数据库实例保存到全局变量
	global.DB = db

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		global.Logger.Error("数据库连接测试失败", zap.Error(err))
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	global.Logger.Info("数据库连接成功")
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(models ...interface{}) error {
	if global.DB == nil {
		err := fmt.Errorf("数据库未初始化")
		global.Logger.Error("自动迁移失败", zap.Error(err))
		return err
	}

	global.Logger.Info("开始自动迁移数据库表结构")
	err := global.DB.AutoMigrate(models...)
	if err != nil {
		global.Logger.Error("自动迁移失败", zap.Error(err))
		return fmt.Errorf("自动迁移失败: %w", err)
	}

	global.Logger.Info("自动迁移完成")
	return nil
}
