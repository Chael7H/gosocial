
# GoSocial 社交平台项目

## 项目概述
这是一个基于Gin框架开发的轻量级社交平台后端服务，提供用户管理、好友关系、消息和帖子等核心社交功能。

## 技术栈
- **后端框架**: Gin + Gorm
- **数据库**: MySQL + Redis
- **认证**: JWT
- **日志**: 自定义日志系统
- **API文档**: Swagger

## 项目结构
```
gosocial_server/
├── conf/          # 配置文件
├── controllers/   # 控制器层
├── dao/           # 数据访问层
│   ├── mysql/     # MySQL操作
│   ├── redis/     # Redis操作
├── docs/          # API文档
├── logic/         # 业务逻辑层
├── middlewares/   # 中间件
├── models/        # 数据模型
├── pkg/           # 公共组件
│   ├── jwt/       # JWT实现
│   ├── snowflake/ # 分布式ID生成
├── routes/        # 路由定义
├── static/        # 静态资源
├── templates/     # 前端模板
├── go.mod         # Go模块定义
└── main.go        # 程序入口
```

## 环境要求
- Go 1.16+
- MySQL 5.7+
- Redis 5.0+

## 安装运行指南
1. 克隆项目:
```bash
git clone [项目地址]
```

2. 安装依赖:
```bash
cd gosocial_server
go mod download
```

3. 配置数据库:
修改 `conf/config.yaml` 中的数据库连接信息

4. 启动服务:
```bash
go run main.go
```

## API文档
项目已集成Swagger文档，启动服务后访问:
http://localhost:8080/swagger/index.html

详细API测试指南请参考 `docs/postman_test_guide.md`
