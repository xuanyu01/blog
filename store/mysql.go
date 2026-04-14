/*
该文件负责初始化 MySQL 数据库连接
*/
package store

import (
	"blog/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQL 创建并返回可用的 MySQL 连接池
func NewMySQL(cfg config.MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// 启动阶段先做连通性检查，尽早暴露配置或服务异常
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 连接池参数统一放在这里管理，业务层不需要关心底层连接细节
	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	fmt.Println("mysql connect success")
	return db, nil
}
