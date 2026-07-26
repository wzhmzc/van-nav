## [v2.4.3] - 2026-07-26

### Fixed
- **中文输入法搜索竞态**：拼音组字期间不驱动列表过滤；汉字上屏后立即 flush 搜索态；确认汉字后立刻回车按汉字（而非拼音）打开首条匹配
- **首页首屏装不满**：按视口与列数估算首屏卡片数，并在内容高度不足时自动补批次，避免默认仅数行需滚动才加载
- **离线/弱网首屏慢、无图**：`FetchList` 增加 4s 硬超时；SWR 优先读 localStorage 缓存；Workbox 对 `/api/img` 使用 CacheFirst，在线成功后预热 logo

### Security
- **Dependabot 依赖修复**：升级 axios 1.16.1 → 1.18.0；pnpm.overrides 锁定 form-data v3/v4 与 launch-editor 至修复版本（CVE-2026-12143、CVE-2026-53632）

### Changed
- **搜索防抖**：300ms → 150ms，兼顾连续键入合并与汉字上屏体感

### ⚠️ 风险提示
本项目代码部分由 AI 自动修改，无法保证完全无误。请在升级前备份重要数据，并自行评估使用风险。

---

## 升级注意事项
1. **纯前端体验优化**：无数据库 schema 变更，直接替换二进制/镜像即可
2. **Service Worker**：首次打开新版本后会更新 SW；若离线图片仍异常，请强制刷新（Ctrl+Shift+R）一次以激活新的 `/api/img` 缓存策略
3. **依赖安全**：前端 lockfile 已随 axios/overrides 更新；自建前端时请使用仓库内 `ui/pnpm-lock.yaml` 安装
