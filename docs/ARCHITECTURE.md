# 架构说明

## 顶层目录

```text
.
├── app              # 应用装配
├── cmd              # 未定义
├── config           # 配置定义
├── docs             # 文档说明
├── http             # HTTP 传输层
├── model            # 领域模型
├── repository       # 数据访问
├── service          # 业务逻辑
├── session          # Session 抽象与实现
├── store            # MySQL / Redis 初始化
├── templates        # 页面模板和静态资源
├── view             # 模板加载
└── main.go          # 程序入口
```

## 分层关系

- `main.go` 负责启动应用。
- `app` 负责依赖装配。
- `store` 和 `session` 负责基础设施能力。
- `repository` 负责数据库读写。
- `service` 负责业务逻辑。
- `http` 负责路由、中间件和请求处理。
- `view` 负责模板加载。
- `model` 负责承载领域数据结构。

## 当前说明


