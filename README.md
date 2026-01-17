# Anywhere Agent

Anywhere Agent 是一个基于 Go 语言开发的代理服务管理工具，主要用于自动化部署、管理和监控 V2Ray 服务，并提供 RESTful API 接口进行远程管理。

## 项目架构

### 目录结构

```
├── cmd/                    # 命令行入口
│   └── agent/              # Agent 主程序入口
│       └── main.go         # 程序启动文件
├── conf/                   # 配置文件目录
│   └── config.yaml.example # 配置文件示例
├── internal/               # 内部代码包
│   ├── agent/              # Agent 核心逻辑
│   ├── api/                # API 服务
│   ├── aws/                # AWS EC2 集成
│   ├── config/             # 配置管理
│   ├── logger/             # 日志系统
│   └── v2ray/              # V2Ray 管理
├── scripts/                # 辅助脚本
│   ├── aw_agent.service    # systemd 服务文件
│   └── install.sh          # 安装脚本
├── build.sh                # 构建脚本
├── go.mod                  # Go 模块依赖
└── go.sum                  # 依赖校验和
```

### 核心组件

1. **Agent 核心** (`internal/agent/`)
   - 负责协调各个子模块
   - 启动和管理 V2Ray 服务
   - 调度定期任务

2. **API 服务** (`internal/api/`)
   - 基于 Gin 框架实现的 RESTful API
   - 提供状态查询和配置获取接口

3. **AWS EC2 集成** (`internal/aws/`)
   - 与 AWS EC2 API 交互
   - 支持实例自动终止
   - 自动获取实例元数据

4. **配置管理** (`internal/config/`)
   - 命令行参数解析
   - 配置文件加载和验证
   - 全局配置管理

5. **日志系统** (`internal/logger/`)
   - 基于 Zap 实现的高性能日志
   - 支持日志轮转和归档
   - 可配置日志级别

6. **V2Ray 管理** (`internal/v2ray/`)
   - V2Ray 自动部署和配置
   - 状态监控和流量统计
   - 日志分析

### 技术栈

- **语言**: Go 1.25.5
- **Web 框架**: Gin 1.11.0
- **日志库**: Zap 1.27.1
- **AWS SDK**: AWS SDK for Go v2
- **配置解析**: yaml.v2
- **代理服务**: V2Ray

## 功能特性

1. **自动化部署**
   - 自动下载和安装 V2Ray
   - 自动配置 V2Ray 服务
   - 支持系统服务自动启动

2. **流量监控**
   - 实时监控 V2Ray 流量
   - 支持空闲超时检测
   - 流量数据持久化

3. **API 管理**
   - RESTful API 接口
   - 状态查询和配置获取

4. **AWS 集成**
   - EC2 实例自动终止
   - 支持实例元数据获取
   - 自动适配 AWS 区域

5. **高可用性**
   - 健康检查机制
   - 优雅关闭
   - 错误处理和恢复

## 工作流程

1. **启动流程**
   - 解析命令行参数
   - 加载配置文件
   - 初始化日志系统
   - 创建 Agent 实例
   - 部署 V2Ray 服务
   - 启动 API 服务器
   - 启动调度器

2. **运行时流程**
   - 定期检查 V2Ray 状态
   - 监控流量使用情况
   - 响应 API 请求
   - 根据配置自动管理实例生命周期

3. **关闭流程**
   - 接收终止信号
   - 停止调度器
   - 等待所有 goroutine 完成
   - 优雅退出

## 配置说明

### 配置文件示例

```yaml
v2ray:
  port: 10086
  uuid: "your-uuid-here"
  access_log: "/var/log/v2ray/access.log"

api:
  address: "0.0.0.0"
  port: 8080

checks:
  traffic_interval: 300  # 流量检查间隔（秒）
  idle_timeout: 3600     # 空闲超时时间（秒）

log:
  level: "info"
  max_size: 100          # 单日志文件最大大小（MB）
  max_backups: 5         # 保留日志文件数量
  max_age: 30            # 日志文件最大保留天数（天）
```

### 配置项说明

| 配置项 | 类型 | 描述 |
|--------|------|------|
| v2ray.port | int | V2Ray 服务监听端口 |
| v2ray.uuid | string | V2Ray 客户端连接 UUID |
| v2ray.access_log | string | V2Ray 访问日志路径 |
| api.address | string | API 服务监听地址 |
| api.port | int | API 服务监听端口 |
| checks.traffic_interval | int | 流量检查间隔（秒） |
| checks.idle_timeout | int | 空闲超时时间（秒） |
| log.level | string | 日志级别（debug, info, warn, error） |
| log.max_size | int | 单日志文件最大大小（MB） |
| log.max_backups | int | 保留日志文件数量 |
| log.max_age | int | 日志文件最大保留天数（天） |

## API 接口

### 健康检查

```
GET /health
```

**响应示例**:
```json
{
  "status": "ok",
  "service": "anywhere-agent"
}
```

### 获取状态和配置

```
GET /api/status
```

**响应示例**:
```json
{
  "status": {
    "installed": true,
    "running": true,
    "version": "v4.45.2"
  },
  "config": {
    "port": 10086,
    "uuid": "your-uuid-here",
    "access_log": "/var/log/v2ray/access.log"
  }
}
```

## 部署方式

### 手动部署

1. **编译项目**
   ```bash
   ./build.sh
   ```

2. **准备配置文件**
   ```bash
   cp conf/config.yaml.example conf/config.yaml
   # 编辑配置文件
   vim conf/config.yaml
   ```

3. **启动服务**
   ```bash
   ./bin/agent --config conf/config.yaml
   ```

### 系统服务部署

1. **使用安装脚本**
   ```bash
   sudo ./scripts/install.sh
   ```

2. **服务管理命令**
   ```bash
   # 启动服务
   sudo systemctl start aw_agent
   
   # 停止服务
   sudo systemctl stop aw_agent
   
   # 重启服务
   sudo systemctl restart aw_agent
   
   # 查看服务状态
   sudo systemctl status aw_agent
   
   # 设置开机自启
   sudo systemctl enable aw_agent
   
   # 禁用开机自启
   sudo systemctl disable aw_agent
   
   # 查看服务日志
   sudo journalctl -u aw_agent
   sudo journalctl -u aw_agent -f
   ```

## 开发指南

### 环境要求

- Go 1.25.5 或更高版本
- Git

### 依赖管理

```bash
go mod tidy
```

### 编译命令

```bash
./build.sh
```

### 运行命令

```bash
./bin/agent --config conf/config.yaml
```

## 监控与维护

### 日志查看

```bash
# 查看 Agent 日志
tail -f /var/log/aw_agent/agent.log

# 查看 V2Ray 日志
tail -f /var/log/v2ray/access.log
tail -f /var/log/v2ray/error.log
```

### 服务状态检查

```bash
sudo systemctl status aw_agent
```

### 重启服务

```bash
sudo systemctl restart aw_agent
```

## 安全最佳实践

1. **访问控制**
   - 限制 API 服务监听地址
   - 定期检查 API 访问日志

2. **V2Ray 安全**
   - 使用强 UUID
   - 定期更新 V2Ray 版本
   - 配置适当的防火墙规则

3. **AWS 安全**
   - 使用最小权限原则配置 IAM 角色
   - 定期检查 EC2 实例安全组配置
   - 启用 AWS CloudTrail 监控

## 故障排除

### 常见问题

1. **V2Ray 部署失败**
   - 检查网络连接
   - 查看 V2Ray 安装日志
   - 检查系统权限

2. **API 访问拒绝**
   - 检查 API 服务是否运行

3. **流量监控异常**
   - 检查 V2Ray 访问日志路径是否正确
   - 检查日志文件权限
   - 重启 Agent 服务

## 版本更新

### 升级步骤

1. **停止服务**
   ```bash
   sudo systemctl stop aw_agent
   ```

2. **更新代码**
   ```bash
   git pull
   ```

3. **重新编译**
   ```bash
   ./build.sh
   ```

4. **启动服务**
   ```bash
   sudo systemctl start aw_agent
   ```

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进 Anywhere Agent。

1. Fork 本项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 联系方式

如有问题或建议，请通过以下方式联系：

- GitHub Issues: https://github.com/yuhai94/anywhere_agent/issues

---

Anywhere Agent - 让代理服务管理更简单！