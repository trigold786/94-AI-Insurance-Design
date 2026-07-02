# 澶氬眰娆℃斂绛栨簮瀹屾暣瀹炵幇璁捐鏂规

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

## 鐩爣

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

瀹炵幇 PRD 瀹氫箟鐨?HIGH/MEDIUM/LOW 涓夌骇鏀跨瓥婧愪綋绯伙紝鏂板 RSS 鐖櫕鍜屾墜鍔ㄥ鍏ョ被鍨嬶紝琛ュ厖 MEDIUM/LOW 绉嶅瓙鏁版嵁銆?
## 鐜扮姸

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- 鏋舵瀯宸插氨缁細DB schema 鏀寔 HIGH/MEDIUM/LOW锛岃皟搴﹀櫒宸紓鍖栭鐜囷紝缃俊搴﹀姞鏉冭瘎鍒?- 瀹為檯鍙湁 7 涓?HIGH 绾у埆鏀垮簻瀹樼綉/鏈湴鏂囦欢婧?- `crawl_type='rss'` 鍦ㄤ唬鐮佹敞閲婁腑鍒楀嚭浣嗘湭瀹炵幇
- 缂哄皯 MEDIUM锛堜富娴佸獟浣擄級鍜?LOW锛堢ぞ濯?鍙嶉锛夌骇鍒簮

## 鏂规

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### 1. RSS 鐖櫕 (`rss_crawler.go`)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

鏂板 `crawl_type='rss'` 瀹炵幇锛屽鐢ㄧ幇鏈?`Source` 鎺ュ彛銆?
- 瑙ｆ瀽 RSS 2.0 / Atom feed XML
- 鎻愬彇 `<item>` / `<entry>` 鐨?title銆乴ink銆乨escription/content
- 瀵规瘡涓潯鐩敓鎴?`CrawlResult`锛圚TML 鍐呭瀛?raw_text锛?- 鏀寔 `max_items` 閰嶇疆锛堥粯璁?20 鏉★級
- 鐗堟湰鍘婚噸锛氱敤 URL 鍋?`VersionHash`
- 瓒呮椂 30s锛岄噸璇?1 娆?
瀹炵幇鏂瑰紡锛欸o 鏍囧噯搴?`encoding/xml` 瑙ｆ瀽 RSS/Atom锛屾棤闇�澶栭儴渚濊禆銆?
### 2. 鎵嬪姩瀵煎叆绫诲瀷 (`crawl_type='manual'`)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

鏂板 `manual_crawler.go`锛屽疄鐜?`Source` 鎺ュ彛浣?`Fetch()` 杩斿洖绌猴紙涓嶈嚜鍔ㄧ埇鍙栵級銆?鏁版嵁閫氳繃 Admin API `/admin/ingest` 鎴栨柊鐨?`/admin/sources/import` 鎵嬪姩瀵煎叆銆?
### 3. Manager 娉ㄥ唽

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

`manager.go` Init() 澧炲姞 `case "rss"` 鍜?`case "manual"` 鍒嗘敮銆?
### 4. 绉嶅瓙鏁版嵁 (Migration 013)

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

MEDIUM 绾у埆锛圧SS 濯掍綋婧愶級锛?| source_id | 鍚嶇О | URL | region_code |
|---|---|---|---|
| MEDIA-RM-SS | 浜烘皯鏃ユ姤绀句繚棰戦亾 | http://www.people.com.cn/rss/social_0.xml | |
| MEDIA-XH-SS | 鏂板崕缃戠ぞ淇濋閬?| http://www.news.cn/rss/social.xml | |
| MEDIA-CE-SS | 涓浗缁忔祹缃戠ぞ淇?| http://www.ce.cn/rss/rss_social.xml | |
| MEDIA-21JJ | 21涓栫邯缁忔祹鎶ラ亾 | https://www.21jingji.com/rss/ | |
| MEDIA-YL-ZX | 鍏昏�佽祫璁綉 | https://www.yanglaocn.com/rss/ | |

LOW 绾у埆锛堟墜鍔ㄥ鍏ワ級锛?| source_id | 鍚嶇О | region_code |
|---|---|---|
| MANUAL-WX | 寰俊鍏紬鍙锋斂绛栨眹鎬?| |
| MANUAL-USER | 鐢ㄦ埛鍙嶉/绀惧尯鎻愪氦 | |
| MANUAL-EXPERT | 涓撳瀹℃牳褰曞叆 | |

### 5. Admin 鎵嬪姩瀵煎叆鎺ュ彛

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

鏂板 `POST /admin/sources/import`锛氭帴鏀?JSON `{source_id, title, content, source_url}` 鐩存帴瀵煎叆鏀跨瓥鍘熸枃銆?
## 鏂囦欢娓呭崟

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### 鏂板
- `internal/crawler/rss_crawler.go` 鈥?RSS/Atom 鐖櫕
- `internal/crawler/rss_crawler_test.go` 鈥?RSS 瑙ｆ瀽娴嬭瘯
- `internal/crawler/manual_crawler.go` 鈥?鎵嬪姩瀵煎叆鐖櫕锛堢┖瀹炵幇锛?- `migrations/013_multi_level_sources.sql` 鈥?绉嶅瓙鏁版嵁

### 淇敼
- `internal/crawler/manager.go` 鈥?娉ㄥ唽 rss/manual 绫诲瀷
- `internal/crawler/crawler.go` 鈥?SourceConfig 娉ㄩ噴鏇存柊
- `internal/admin/admin_llm.go` 鎴?`admin.go` 鈥?鏂板 import handler
- `internal/crawler/store.go` 鈥?鏂板 Store 鏂规硶锛堝彲閫夛級
