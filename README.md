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
go get github.com/yourusername/nmap_sd
```

## 🚀 快速开始

### 基础用法

```go
package main

import (
    "nmap_sd/pkg/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 使用默认配置
    r.Use(middleware.New())

    // 或使用自定义配置
    r.Use(middleware.New(middleware.Config{
        CIDR:         "192.168.2.0/22",  // 扫描的网段
        ScanPath:     "/mgsd",            // API 路径
        ScanInterval: 1,                  // 扫描间隔（分钟）
    }))

    r.Run(":8080")
}
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

## 🔍 扫描的端口

默认扫描以下端口：

- **9182** - Windows Exporter
- **80** - HTTP
- **443** - HTTPS
- **8080** - HTTP Proxy
- **8083, 8089, 8888, 38089** - HTTP 替代端口

## 📝 与 Prometheus 集成

在 Prometheus 配置文件中使用 HTTP SD：

```yaml
scrape_configs:
  - job_name: 'dynamic_discovery'
    http_sd_configs:
      - url: 'http://your-server:8080/mgsd'
        refresh_interval: 60s
```

## 🛠️ 开发

### 前置要求

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
