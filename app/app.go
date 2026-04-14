/*
该文件负责应用装配，把配置、存储、仓储、服务和路由连接起来
*/
package app

import (
	"blog/config"
	"blog/http/handler"
	"blog/http/router"
	"blog/repository"
	"blog/service"
	"blog/session"
	"blog/store"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// App 保存应用运行所需的核心依赖
// 它负责在应用生命周期内持有服务启动需要的基础对象
type App struct {
	server *gin.Engine
	db     *sql.DB
	redis  *redis.Client
	addr   string
}

// New 创建并装配一个可运行的应用实例
func New() (*App, error) {
	cfg := config.New()

	// 先初始化底层基础设施，确保数据库和缓存都可用后再继续向上组装
	db, err := store.NewMySQL(cfg.MySQL)
	if err != nil {
		return nil, err
	}
	redisClient, err := store.NewRedis(cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	// 按照 repository -> service -> handler -> router 的依赖方向逐层组装
	// 这样 main 入口可以保持简洁，模块边界也更清晰
	blogRepo := repository.NewBlogRepository(db)
	userRepo := repository.NewUserRepository(db)
	sessionStore := session.NewRedisStore(redisClient)

	blogService := service.NewBlogService(blogRepo)
	authService := service.NewAuthService(userRepo, sessionStore)
	webHandler := handler.NewWebHandler(blogService, authService)
	server := router.New(webHandler)

	return &App{
		server: server,
		db:     db,
		redis:  redisClient,
		addr:   cfg.Server.Address,
	}, nil
}

// Run 启动 HTTP 服务并在退出时释放资源
func (a *App) Run() error {
	// 连接关闭放在这里统一处理，避免在多个退出分支里重复清理
	defer a.db.Close()
	defer a.redis.Close()
	return a.server.Run(a.addr)
}
