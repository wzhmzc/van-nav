## [v2.4.5] - 2026-11-23

本版基于上游 v2.4.4 基线，只做自用维护，不与上游功能同步。

### Fixed
- **Iconify 图标全挂**：`api.iconify.design` 已把路由从 `<prefix>:<name>.svg` 改成 `<prefix>/<name>.svg`，旧写法一律 404，表现为「图标位置空着、无任何报错」。修复后台图标搜索的 URL 拼接（`LogoSearchSelect`），并同步 `nav_img` 缓存 key —— 只改 `nav_table.logo` 会导致首页图标全变灰圈，因为 `GetImgFromDB` 查不到缓存时不抓取、直接回灰色占位 SVG（而且占位图也是 HTTP 200 + 1236 字节，用状态码判断会误判成正常）
- **后台工具列表图标全部裂开**：原来直接拼 `/api/img?url=${record.logo}` 不做 URL 编码，Go 1.17+ 把 `;` 视为非法 query 分隔符，`data:image/svg+xml;base64,…` 会让整段 query 解析失败被丢弃 → handler 拿到空 url → 400。前端改用 `getLogoUrl()`，后端 `/api/img` 增加 data: URI 直出并从 `RawQuery` 兜底解析
- **抓图存进 HTML 错误页**：Cloudflare / WAF 返回 200 + 挑战页时，原代码会原样 base64 存进 `nav_img`，前端按图片渲染 → 破图。`GetImgBase64FromUrl` 现在有三道校验：非 2xx 放弃、Content-Type 非 image/* 放弃、内容魔数（PNG/JPEG/GIF/ICO/BMP/WEBP/SVG）不符放弃
- **空串切片 panic**：`getIcon` 里 `s.Preview.Link[len-1:]` / `s.Preview.Icon[0:1]` 在页面 meta 抓到空值时会越界 panic，改用 `HasSuffix`/`HasPrefix` 并前置判空；抓取失败时不再用空值覆盖已有图标
- **`GetSuffixFromUrl` 越界**：无点 URL（如 `localhost`）会 `url[-1:]` panic，改为返回空串
- **`backup.go` vet 报错**：`fmt.Errorf(变量)` 在 Go 1.24+ 被判为错误，改为 `%s` 格式化

### Changed
- **构建链 react-scripts(CRA) → Vite**：构建耗时 90s → 11s，二进制 23M → 16M，主 bundle 525KB（gzip 179KB）。同步更新 Dockerfile、build-check、release 的产物路径（`ui/build` → `ui/dist`）
- **`handler/handlers.go` 按业务拆分**：1648 行 → 150 行，拆出 logo / tool / setting / backup / search / auth / catelog 共 7 个文件（同 package 机械拆分，无行为变化）
- **PWA 缓存策略**：`/api/img` 从 `CacheFirst` + 30 天改为 `NetworkFirst`（3s 超时回落）+ 7 天，避免坏图被客户端缓存一个月
- **导入工具时的图片转存**：串行改 5 并发，跳过 `data:` 内嵌图（原来给 base64 发 HTTP 请求只会白等一次超时）
- **依赖归位**：`@testing-library/*`、`@types/*`、`typescript` 从 dependencies 移到 devDependencies（版本未变）

### Added
- **测试**：`utils/image_test.go`（looksLikeImage 魔数校验 16 个用例 + GetMIME + GetSuffixFromUrl）、`ui/src/utils/check.test.ts`（getLogoUrl 7 个用例）；CI 增加 `go test ./...` 和前端测试步骤
- 删除 CRA 模板测试 `App.test.tsx`（测的是 "learn react"，App 中不存在）；`src/__tests__/benchmark.ts` 移至 `scripts/`（ts-node 压测脚本，被 jest 误判为测试套件）

### ⚠️ 风险提示
本项目代码部分由 AI 自动修改，无法保证完全无误。请在升级前备份重要数据，并自行评估使用风险。

---

## 升级注意事项
1. **无数据库 schema 变更**：直接替换二进制即可；`data/` 目录与 `jwt_secret` 无需迁移
2. **前端产物路径变了**：自建编译请用 `ui/dist`（不再是 `ui/build`），需要 Node 22 + pnpm 11
3. **浏览器端**：service worker 换了缓存策略，首次访问后建议 Ctrl+Shift+R 强制刷新一次
4. **图标修复提示**：若从 v2.4.4 升级且库里存的是旧格式 Iconify URL，需把 `nav_table.logo` 与 `nav_img.url` 一起把冒号改斜杠，否则图标仍是灰圈

---

## [v2.4.4] - 2026-07-29

### Security
- **Go 运行时与密码学依赖**：`go` 升至 1.25.0；`golang.org/x/crypto` → 0.52.0、`golang.org/x/net` → 0.55.0、`golang.org/x/text` → 0.37.0，关闭多条 critical/high（含 SSH/crypto 相关）Dependabot 告警
- **CRA 传递依赖 Plan A/B**：`pnpm-workspace.yaml` overrides 钉定 form-data、launch-editor、brace-expansion（含 1.x→1.1.16）、js-yaml、fast-uri、shell-quote、websocket-driver、http-proxy-middleware@2、body-parser@1、serialize-javascript（统一 ≥7.0.5，移除不安全的 @6 钉死）、svgo（1.x/2.x→2.8.3）等，并刷新 `ui/pnpm-lock.yaml`
- **前端直接依赖**：`react-router-dom` ^7.18.0、`postcss` ^8.5.18 等随 lock 对齐

### Changed
- **构建矩阵**：CI（build-check / release）与 Dockerfile 的 Go 版本同步为 1.25
- **pnpm 11**：overrides / allowBuilds 迁入 `ui/pnpm-workspace.yaml`（package.json 的 `pnpm` 字段在 pnpm 11 下会被忽略）
