# Blog

这是一个基于 `Gin + MySQL + Redis + Vue` 的博客项目。
当前项目已经支持注册、登录、Session、博客增删改查、分页搜索，以及基础后台用户管理。

## 目录结构

```text
.
├── app
├── config
├── docs
├── frontend
├── http
├── migrations
├── model
├── repository
├── service
├── session
├── store
├── .env.example
├── ai.sql
├── main.go
└── README.md
```

## 配置说明

项目启动时会优先读取当前目录下的 `.env` 文件，同时允许系统环境变量覆盖 `.env` 中的值。
以下配置为必填或常用配置：

```env
APP_ADDR=:5345
MYSQL_DSN=blog:123456@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=true&loc=Local
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
SESSION_EXPIRE_MINUTES=60
```

你可以直接从示例文件复制：

```powershell
Copy-Item .\.env.example .\.env
```

如果缺少关键配置，程序启动时会直接报错并退出。

## 数据库初始化

### 方式一：直接用完整结构脚本重建

`ai.sql` 会直接重建整个 `blog` 库，适合本地全新初始化。

```powershell
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\ai.sql
```

### 方式二：按 migrations 逐步执行

如果你更希望按版本执行初始化，可以按顺序运行 `migrations` 目录中的脚本：

```powershell
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\migrations\001_create_database.sql
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\migrations\002_create_users_and_categories.sql
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\migrations\003_create_posts.sql
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\migrations\004_create_engagement.sql
mysql -u blog -p -h 127.0.0.1 -P 3306 --default-character-set=utf8mb4 < .\migrations\005_seed_data.sql
```

`005_seed_data.sql` 会插入一条默认管理员数据：

- 用户名：`admin`
- 密码：`admin123`

## 启动项目

确保本地 MySQL 和 Redis 已启动，然后执行：

```powershell
go run .\main.go
```

前端静态资源如果需要重新构建：

```powershell
Set-Location .\frontend
npm install
npm run build
Set-Location ..
```

## 测试

当前已补的基础测试主要覆盖：

- 注册 / 登录
- 博客创建 / 更新 / 删除
- 权限判断
- 列表分页参数归一化

运行命令：

```powershell
go test ./service ./repository ./http/... ./model
```

如果你运行 `go test ./...` 失败，请先检查项目根目录里是否还存在无效的临时文件，例如 `tempCodeRunnerFile.go`。
