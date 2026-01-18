# Changelog

All notable changes to CloudBoot NG will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Documentation
- 创建落地冲刺模式文档 MISSION_CONTROL.md (2026-01-18 23:50)
- 创建自动化同步脚本 scripts/sync.sh (2026-01-18 23:50)
- 识别12个技术债，定义P0高危项 (2026-01-18 23:50)

## [1.0.0-alpha] - 2026-01-16

### Added
- ✅ 完成 Platform 核心 100%
- ✅ 完成 CSPM 机制 92%
- ✅ DRM 完整流程（AES-256, ECDSA, Session Key）
- ✅ .cbp 包解析器
- ✅ 水印审计系统
- ✅ Adaptor 双层架构
- ✅ Provider Schema + User Overlay
- ✅ 配置生成引擎（Kickstart/Preseed/AutoYaST）
- ✅ BootOS Agent（cb-agent/cb-probe/cb-exec）
- ✅ E2E 测试环境
- ✅ embed.FS 单体二进制

### Fixed
- ✅ 修复模板名称冲突
- ✅ 修复 OS Designer 重复定义
- ✅ 修复左侧 Sidebar 布局问题

### Testing
- ✅ 测试覆盖率提升至 65%
- ✅ 151+ 单元测试全部通过
- ✅ CSPM DRM 测试 19 个用例通过

### Documentation
- ✅ 完成架构设计文档
- ✅ 完成 API 规范文档
- ✅ 完成 CSPM 实施报告
- ✅ 重组根目录文档结构

## [0.1.0] - 2026-01-15

### Added
- 项目初始化
- Go mod + Makefile 配置
- UI 组件库（Card, Button, Badge, Terminal）
- 数据层（Gorm Models）
- CSPM 引擎基础框架
- Mock Provider

---

[Unreleased]: https://github.com/huha-yy/cloudboot-ng/compare/v1.0.0-alpha...HEAD
[1.0.0-alpha]: https://github.com/huha-yy/cloudboot-ng/compare/v0.1.0...v1.0.0-alpha
[0.1.0]: https://github.com/huha-yy/cloudboot-ng/releases/tag/v0.1.0

### Added
- 实现Provider原子序列编排器(Orchestrator) - Plan→Probe→Apply闭环 (2026-01-19 00:30)
- 强制Plan步骤预演变更，确保配置合法性
- Probe步骤实现幂等性检查，跳过重复操作
- Apply步骤执行实际变更，Verify步骤验证结果
- 所有步骤支持超时熔断(context.WithTimeout)
