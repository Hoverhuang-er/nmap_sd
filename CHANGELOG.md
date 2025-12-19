# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-12-19

### Added
- 🎉 首次发布 - 基于 Gin 的网络服务发现中间件
- 🔍 支持 CIDR 网段自动扫描
- 🔌 多端口扫描支持（HTTP, HTTPS, Windows Exporter 等）
- 🔄 定时自动扫描和更新机制
- 📊 Prometheus 服务发现格式支持
- ✅ 完整的单元测试覆盖
- 📚 详细的使用文档和示例

### Changed
- 🔄 从 GoFiber 迁移到 Gin 框架
- 📦 从独立二进制应用改为可重用的 Go 库

### Technical Details
- Go 1.18+ 支持，兼容 Go 1.23
- 跨平台支持（Linux, macOS, Windows）
- GitHub Actions CI/CD 自动化
- 完整的测试套件

### Breaking Changes
- 不再发布预编译的二进制文件
- 使用 `go get` 安装而非下载可执行文件
- API 端点从 `/dip` 改为 `/mgsd`

[Unreleased]: https://github.com/Hoverhuang-er/nmap_sd/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Hoverhuang-er/nmap_sd/releases/tag/v1.0.0
