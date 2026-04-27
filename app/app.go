/*
该文件负责应用装配，把配置、存储、仓储、服务和路由连接起来。
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

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App 保存应用运行所需的核心依赖。
type App struct {
	server *gin.Engine
	db     *gorm.DB
	redis  *redis.Client
	addr   string
}

// New 创建并装配一个可运行的应用实例。
func New() (*App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	session.SetExpire(cfg.Session.Expire)

	db, err := store.NewMySQL(cfg.MySQL)
	if err != nil {
		return nil, err
	}
	redisClient, err := store.NewRedis(cfg.Redis)
	if err != nil {
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			_ = sqlDB.Close()
		}
		return nil, err
	}

	blogRepo := repository.NewBlogRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	userRepo := repository.NewUserRepository(db)
	sessionStore := session.NewRedisStore(redisClient)

	blogService := service.NewBlogService(blogRepo)
	commentService := service.NewCommentService(commentRepo, blogRepo)
	authService := service.NewAuthService(userRepo, sessionStore)
	webHandler := handler.NewWebHandler(blogService, commentService, authService)
	server := router.New(webHandler)

	return &App{
		server: server,
		db:     db,
		redis:  redisClient,
		addr:   cfg.Server.Address,
	}, nil
}

// Run 启动 HTTP 服务并在退出时释放资源。
func (a *App) Run() error {
	sqlDB, err := a.db.DB()
	if err != nil {
		return err
	}

	defer sqlDB.Close()
	defer a.redis.Close()
	return a.server.Run(a.addr)
}
