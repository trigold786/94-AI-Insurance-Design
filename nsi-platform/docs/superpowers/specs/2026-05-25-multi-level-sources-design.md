# 多层次政策源完整实现设计方案

## 目标

实现 PRD 定义的 HIGH/MEDIUM/LOW 三级政策源体系，新增 RSS 爬虫和手动导入类型，补充 MEDIUM/LOW 种子数据。

## 现状

- 架构已就绪：DB schema 支持 HIGH/MEDIUM/LOW，调度器差异化频率，置信度加权评分
- 实际只有 7 个 HIGH 级别政府官网/本地文件源
- `crawl_type='rss'` 在代码注释中列出但未实现
- 缺少 MEDIUM（主流媒体）和 LOW（社媒/反馈）级别源

## 方案

### 1. RSS 爬虫 (`rss_crawler.go`)

新增 `crawl_type='rss'` 实现，复用现有 `Source` 接口。

- 解析 RSS 2.0 / Atom feed XML
- 提取 `<item>` / `<entry>` 的 title、link、description/content
- 对每个条目生成 `CrawlResult`（HTML 内容存 raw_text）
- 支持 `max_items` 配置（默认 20 条）
- 版本去重：用 URL 做 `VersionHash`
- 超时 30s，重试 1 次

实现方式：Go 标准库 `encoding/xml` 解析 RSS/Atom，无需外部依赖。

### 2. 手动导入类型 (`crawl_type='manual'`)

新增 `manual_crawler.go`，实现 `Source` 接口但 `Fetch()` 返回空（不自动爬取）。
数据通过 Admin API `/admin/ingest` 或新的 `/admin/sources/import` 手动导入。

### 3. Manager 注册

`manager.go` Init() 增加 `case "rss"` 和 `case "manual"` 分支。

### 4. 种子数据 (Migration 013)

MEDIUM 级别（RSS 媒体源）：
| source_id | 名称 | URL | region_code |
|---|---|---|---|
| MEDIA-RM-SS | 人民日报社保频道 | http://www.people.com.cn/rss/social_0.xml | |
| MEDIA-XH-SS | 新华网社保频道 | http://www.news.cn/rss/social.xml | |
| MEDIA-CE-SS | 中国经济网社保 | http://www.ce.cn/rss/rss_social.xml | |
| MEDIA-21JJ | 21世纪经济报道 | https://www.21jingji.com/rss/ | |
| MEDIA-YL-ZX | 养老资讯网 | https://www.yanglaocn.com/rss/ | |

LOW 级别（手动导入）：
| source_id | 名称 | region_code |
|---|---|---|
| MANUAL-WX | 微信公众号政策汇总 | |
| MANUAL-USER | 用户反馈/社区提交 | |
| MANUAL-EXPERT | 专家审核录入 | |

### 5. Admin 手动导入接口

新增 `POST /admin/sources/import`：接收 JSON `{source_id, title, content, source_url}` 直接导入政策原文。

## 文件清单

### 新增
- `internal/crawler/rss_crawler.go` — RSS/Atom 爬虫
- `internal/crawler/rss_crawler_test.go` — RSS 解析测试
- `internal/crawler/manual_crawler.go` — 手动导入爬虫（空实现）
- `migrations/013_multi_level_sources.sql` — 种子数据

### 修改
- `internal/crawler/manager.go` — 注册 rss/manual 类型
- `internal/crawler/crawler.go` — SourceConfig 注释更新
- `internal/admin/admin_llm.go` 或 `admin.go` — 新增 import handler
- `internal/crawler/store.go` — 新增 Store 方法（可选）
