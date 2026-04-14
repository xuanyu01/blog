# Blog

这是一个基于 `Gin + MySQL + Redis` 的个人博客示例项目，目前已完成基础的注册、登录、Session 鉴权和页面渲染。

## 目录结构

```text
.
├── app
├── cmd
│   └── blog
├── config
├── docs
├── http
│   ├── handler
│   ├── middleware
│   └── router
├── model
├── repository
├── service
├── session
├── store
├── templates
│   ├── css
│   └── img
├── view
├── go.mod
└── go.sum
```

## 分层说明

- `app`：应用装配与依赖注入。
- `cmd/blog`：程序启动入口。
- `config`：配置定义。
- `http`：路由、中间件和页面/接口处理器。
- `model`：领域模型。
- `repository`：数据库访问层。
- `service`：业务逻辑层。
- `session`：Session 抽象和 Redis 实现。
- `store`：MySQL、Redis 基础设施初始化。
- `view`：模板加载。
- `templates`：HTML 模板和静态资源。
- `docs`：架构和开发说明。

## 运行方式

确保本地 MySQL 和 Redis 已启动，然后执行：

```powershell
go run ./cmd/blog
```

默认监听端口：

```text
:5345
```
