## [v2.4.4] - 2026-07-29

### Security
- **Go 运行时与密码学依赖**：`go` 升至 1.25.0；`golang.org/x/crypto` → 0.52.0、`golang.org/x/net` → 0.55.0、`golang.org/x/text` → 0.37.0，关闭多条 critical/high（含 SSH/crypto 相关）Dependabot 告警
- **CRA 传递依赖 Plan A/B**：`pnpm-workspace.yaml` overrides 钉定 form-data、launch-editor、brace-expansion（含 1.x→1.1.16）、js-yaml、fast-uri、shell-quote、websocket-driver、http-proxy-middleware@2、body-parser@1、serialize-javascript（统一 ≥7.0.5，移除不安全的 @6 钉死）、svgo（1.x/2.x→2.8.3）等，并刷新 `ui/pnpm-lock.yaml`
- **前端直接依赖**：`react-router-dom` ^7.18.0、`postcss` ^8.5.18 等随 lock 对齐

### Changed
- **构建矩阵**：CI（build-check / release）与 Dockerfile 的 Go 版本同步为 1.25
- **pnpm 11**：overrides / allowBuilds 迁入 `ui/pnpm-workspace.yaml`（package.json 的 `pnpm` 字段在 pnpm 11 下会被忽略）

### ⚠️ 风险提示
本项目代码部分由 AI 自动修改，无法保证完全无误。请在升级前备份重要数据，并自行评估使用风险。

---

## 升级注意事项
1. **无数据库 schema 变更**：直接替换二进制/镜像即可；JWT / 数据目录无需迁移
2. **自建前端**：请使用仓库内 `ui/pnpm-lock.yaml` + `ui/pnpm-workspace.yaml` 安装（需 pnpm 11+）
3. **仍开放的依赖告警**：`react-router` 7.x 区间内 Dependabot 仍可能报 RSC 相关项；本项目为 CRA 客户端路由、未使用 RSC，升 8.x 留作后续 Plan C
4. **Service Worker**：若静态资源异常，强制刷新（Ctrl+Shift+R）一次以拉取新 hash 资源
