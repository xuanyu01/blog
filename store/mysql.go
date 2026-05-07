/*
负责初始化 MySQL 与 GORM 数据库连接。
*/
package store

import (
	"blog/config"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// NewMySQL 创建并返回可用的 GORM MySQL 连接。
func NewMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 获取底层 sql.DB 以设置连接池参数。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 测试数据库连接。
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	// 统一设置连接池参数。
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)

	fmt.Println("mysql connect success")
	return db, nil
}
