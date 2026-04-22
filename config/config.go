/*
该文件定义应用运行所需的配置结构和默认配置
*/
package config

// Config 组织服务、数据库和缓存的配置
// 它作为应用启动时的统一配置入口
type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
}

// ServerConfig 表示 HTTP 服务监听配置
type ServerConfig struct {
	Address string
}

// MySQLConfig 表示 MySQL 连接配置
type MySQLConfig struct {
	DSN string
}

// RedisConfig 表示 Redis 连接配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// New 返回当前项目使用的默认配置
func New() Config {
	// 这里集中维护默认值，便于本地开发直接启动项目
	return Config{
		Server: ServerConfig{
			Address: ":5345",
		},
		MySQL: MySQLConfig{
			DSN: "blog:123456@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=true&loc=Local",
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}
