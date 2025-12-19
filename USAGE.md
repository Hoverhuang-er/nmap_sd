# 使用指南

## 快速集成到你的 Gin 应用

### 1. 导入包

```go
import (
    "nmap_sd/pkg/middleware"
    "github.com/gin-gonic/gin"
)
```

### 2. 注册中间件

```go
func main() {
    r := gin.Default()
    
    // 注册 NmapSD 中间件
    r.Use(middleware.New(middleware.Config{
        CIDR:         "192.168.2.0/22",  // 扫描的网段
        ScanPath:     "/mgsd",            // 服务发现 API 路径
        ScanInterval: 1,                  // 每 1 分钟扫描一次
    }))
    
    // 你的其他路由
    r.GET("/api/users", handleUsers)
    r.POST("/api/data", handleData)
    
    r.Run(":8080")
}
```

### 3. 访问服务发现 API

```bash
curl http://localhost:8080/mgsd
```

返回示例：

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
      "192.168.2.11:8080"
    ],
    "labels": {
      "job": "http_services"
    }
  }
]
```

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `CIDR` | 要扫描的网络范围，支持 CIDR 格式 | `192.168.2.0/22` |
| `ScanPath` | 服务发现 API 的访问路径 | `/mgsd` |
| `ScanInterval` | 扫描间隔（分钟） | `1` |

## Prometheus 配置示例

在 Prometheus 的 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'network_discovery'
    http_sd_configs:
      - url: 'http://your-server:8080/mgsd'
        refresh_interval: 60s
```

## 运行要求

1. **安装 nmap**
   ```bash
   # macOS
   brew install nmap
   
   # Linux
   sudo apt-get install nmap  # Debian/Ubuntu
   sudo yum install nmap      # CentOS/RHEL
   ```

2. **权限要求**
   - nmap 需要 root 权限进行完整扫描
   - 或者给 nmap 配置 capabilities（Linux）
   ```bash
   sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip $(which nmap)
   ```

## 完整示例

查看 `example/main.go` 获取完整的可运行示例。

```bash
cd example
go run main.go
```
