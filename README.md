# NmapSD - Network Service Discovery Middleware for Gin

![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

NmapSD 是一个基于 Gin 的网络服务发现中间件，使用 nmap 自动扫描网络中的活跃主机和开放端口，适用于 Prometheus 服务发现等场景。

## ✨ 特性

- 🔍 自动扫描指定 CIDR 网段的活跃主机
- 🔌 检测常用端口（HTTP, HTTPS, Windows Exporter 等）
- 🔄 定时自动扫描和更新
- 🎯 以 Gin 中间件形式无缝集成
- 📊 支持 Prometheus 服务发现格式
- 🚀 并发扫描，性能优异

## 📦 安装

```bash
go get github.com/Hoverhuang-er/nmap_sd@latest
```

或在你的项目中：

```bash
go mod init your-project
go get github.com/Hoverhuang-er/nmap_sd
```

## 🚀 快速开始

### 1. 在现有 Gin 项目中使用

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Hoverhuang-er/nmap_sd/pkg/middleware"
)

func main() {
    r := gin.Default()

    // 注册 NmapSD 中间件（使用默认端口）
    r.Use(middleware.New(middleware.Config{
        CIDR:         "192.168.2.0/22",  // 扫描的网段
        ScanPath:     "/mgsd",            // API 路径
        ScanInterval: 1,                  // 扫描间隔（分钟）
        // Ports: 留空使用默认端口，或自定义端口列表
    }))

    // 你的其他路由
    r.GET("/api/users", handleUsers)
    r.POST("/api/data", handleData)

    r.Run(":8080")
}
```

### 2. 使用默认配置

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Hoverhuang-er/nmap_sd/pkg/middleware"
)

func main() {
    r := gin.Default()
    
    // 使用默认配置（CIDR: 192.168.2.0/22, Path: /mgsd, Interval: 1分钟）
    r.Use(middleware.New())
    
    r.Run(":8080")
}
```

### 3. 作为独立服务运行

克隆仓库并运行示例：

```bash
git clone https://github.com/Hoverhuang-er/nmap_sd.git
cd nmap_sd/example
go run main.go
```

### API 响应示例

访问 `GET /mgsd` 获取发现的服务：

```json
[
  {
    "targets": [
      "192.168.2.1:9182",
      "192.168.2.2:9182"
    ],
    "labels": {
      "job": "windows_exporter"
    }
  },
  {
    "targets": [
      "192.168.2.10:80",
      "192.168.2.11:8080",
      "192.168.2.12:8083"
    ],
    "labels": {
      "job": "http_services"
    }
  }
]
```

## ⚙️ 配置选项

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `CIDR` | string | `"192.168.2.0/22"` | 要扫描的网络 CIDR |
| `ScanPath` | string | `"/mgsd"` | API 端点路径 |
| `ScanInterval` | int | `1` | 扫描间隔（分钟） |
| `Ports` | []sd.PortService | 见下方 | 要扫描的端口列表 |

## 🔍 扫描的端口

默认扫描以下端口（可通过 `Config.Ports` 自定义）：

- **9182** - Windows Exporter
- **80** - HTTP
- **443** - HTTPS
- **8080** - HTTP Proxy
- **8083, 8089, 8888, 38089** - HTTP 替代端口

## 📝 与 Prometheus 集成

在 Prometheus 配置文件中使用 HTTP SD：

```yaml
scrape_configs:
  - job_name: 'dynamic_d     # Gin 中间件
│   │   └── nmap_sd.go      # 中间件核心实现
│   └── sd/                 # 扫描逻辑
│       └── scanner.go      # Nmap 扫描封装
├── example/                # 完整示例
│   └── main.go            # 示例主程序
├── .github/
│   └── workflows/         # CI/CD 配置
│       ├── ci.yml        # 持续集成
│       └── release-package.yml   # 版本发布
├── go.mod
├── LICENSE
├── README.md             # 项目文档
└── USAGE.md             # 使用指南
```

## 📚 完整示例

查看 [example/main.go](example/main.go) 获取完整的可运行示例。

### 运行示例

```bash
# 克隆仓库
git clone https://github.com/Hoverhuang-er/nmap_sd.git
cd nmap_sd/example

# 运行
go run main.go

# 或编� 进阶配置

### 自定义扫描端口

通过 `Config.Ports` 配置自定义端口：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/Hoverhuang-er/nmap_sd/pkg/middleware"
    "github.com/Hoverhuang-er/nmap_sd/pkg/sd"
)

func main() {
    r := gin.Default()
    
    r.Use(middleware.New(middleware.Config{
        CIDR:         "192.168.1.0/24",
        ScanPath:     "/mgsd",
        ScanInterval: 5,
        Ports: []sd.PortService{
            {Port: 9182, Name: "windows_exporter", Job: "windows_exporter"},
            {Port: 3000, Name: "custom-app", Job: "my_services"},
            {Port: 5000, Name: "api-server", Job: "my_services"},
            {Port: 8080, Name: "web-app", Job: "web_services"},
        },
    }))
    
    r.Run(":8080")
}
```

### 多网段扫描

```go
// 扫描多个网段，需要启动多个中间件实例
r.Use(middleware.New(middleware.Config{
    CIDR:     "192.168.1.0/24",
    ScanPath: "/mgsd/network1",
}))

r.Use(middleware.New(middleware.Config{
    CIDR:     "10.0.0.0/24",
    ScanPath: "/mgsd/network2",
}))
```

## 📦 版本发布

本项目使用 GitHub Actions 自动化 CI/CD：

- **持续集成**: 每次 push 或 PR 时运行测试和 lint
- **版本发布**: 推送 tag 时自动创建 GitHub Release

### 发布新版本

```bash
# 打标签
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# GitHub Actions 会自动创建 Release
```

### 使用特定版本

```bash
# 使用最新版本
go get github.com/Hoverhuang-er/nmap_sd@latest

# 使用特定版本
go get github.com/Hoverhuang-er/nmap_sd@v1.0.0

# 使用特定 commit
go get github.com/Hoverhuang-er/nmap_sd@commit-hash
```

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 开发指南

```bash
# 克隆仓库
git clone https://github.com/Hoverhuang-er/nmap_sd.git
cd nmap_sd

# 安装依赖
go mod download

# 运行测试
go test -v ./...

# 运行 lint
golangci-lint run

# 构建示例
cd example && go build
```

## ⚠️ 注意事项

1. **权限要求**: nmap 需要 root/sudo 权限进行完整扫描
   ```bash
   # Linux: 为 nmap 添加 capabilities
   sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip $(which nmap)
   
   # 或使用 sudo 运行程序
   sudo ./your-program
   ```

2. **性能考虑**: 
   - 大网段（/22, /16）扫描可能需要 5-10 分钟
   - 建议生产环境扫描间隔设置为 5-10 分钟
   - 首次扫描会阻塞，建议在后台初始化

3. **网络影响**:
   - nmap 扫描会产生网络流量
   - 可能触发某些网络安全设备的告警
   - 建议在内网环境使用

4. **依赖要求**:
   - 必须安装 nmap 命令行工具
   - Go 1.18 或更高版本

## 🐛 故障排查

### nmap: command not found

```bash
# 安装 nmap
brew install nmap      # macOS
sudo apt install nmap  # Ubuntu/Debian
sudo yum install nmap  # CentOS/RHEL
```

### 扫描无结果

1. 检查 nmap 是否有足够权限
2. 验证 CIDR 配置是否正确
3. 查看日志输出排查错误

### 扫描速度慢

1. 减小扫描范围（使用更大的子网掩码）
2. 增加扫描间隔
3. 减少扫描的端口数量

## 📞 联系方式

- 提交 Issue: [GitHub Issues](https://github.com/Hoverhuang-er/nmap_sd/issues)
- 讨论: [GitHub Discussions](https://github.com/Hoverhuang-er/nmap_sd/discussions)

## 🌟 Star History

如果这个项目对你有帮助，请给它一个 ⭐️！

## 🔗 相关项目

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [nmap](https://nmap.org/) - Network exploration tool
- [Prometheus](https://prometheus.io/) - Monitoring system

- Go 1.18+
- nmap 安装在系统中

```bash
# macOS
brew install nmap

# Ubuntu/Debian
sudo apt-get install nmap

# CentOS/RHEL
sudo yum install nmap
```

### 运行示例

```bash
cd example
go run main.go
```

访问 `http://localhost:8080/mgsd` 查看扫描结果。

## 📂 项目结构

```
nmap_sd/
├── pkg/
│   ├── middleware/     # Gin 中间件
│   │   └── nmap_sd.go
│   └── sd/            # 扫描逻辑
│       └── scanner.go
├── example/           # 示例代码
│   └── main.go
├── go.mod
└── README.md
```

## 🔄 从 v1 迁移

**v1 (GoFiber)** → **v2 (Gin)**

```go
// v1 - GoFiber
app := fiber.New()
gc := gocron.NewScheduler(time.Local)
gc.Every(1).Minute().Do(sd.ScanNetworkWNmap)
app.Get("/dip", route.DisplayMgSDInfo)

// v2 - Gin Middleware
r := gin.Default()
r.Use(middleware.New(middleware.Config{
    CIDR:         "192.168.2.0/22",
    ScanPath:     "/mgsd",
    ScanInterval: 1,
}))
```

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## ⚠️ 注意事项

1. 需要 root/sudo 权限运行 nmap（或配置 nmap 的 capabilities）
2. 大网段扫描可能需要较长时间
3. 建议在生产环境中合理设置扫描间隔

## 📞 联系方式

如有问题或建议，请提交 Issue。
