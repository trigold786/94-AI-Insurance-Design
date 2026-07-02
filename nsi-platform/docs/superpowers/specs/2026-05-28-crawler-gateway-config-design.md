# Crawler Config Integration with llm-gateway

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**Date:** 2026-05-28
**Status:** Approved
**Approach:** Config-only (Approach A - Replace entirely)

## Problem

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

The policy-crawler service maintains its own `llm_configs` and `asr_configs` tables in the `nsi_crawler` database, separate from the unified `model_configs` table in `nsi_llm` (managed by llm-gateway). This creates:
- Dual config management (admin must configure in two places)
- Config drift risk (crawler and gateway may disagree)
- No single source of truth for model configuration

## Decision

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Replace all crawler config reads with calls to llm-gateway's unified `model_configs` API. The crawler will continue to call LLM/ASR/Embedding provider APIs directly 鈥?only the config source changes.

## Architecture

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

```
llm-gateway (model_configs table, :39404)
    |
    | GET /internal/model-configs/{function_key}
    | (unauthenticated, Docker-internal only)
    |
    v
crawler [GatewayConfigClient]
    |
    | reads config on demand (60s TTL cache)
    |
    v
llm.Client / ASRProvider / EmbeddingProvider
(calls provider APIs directly - unchanged)
```

## Components

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### 1. New internal endpoint in llm-gateway

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**File:** `services/llm-gateway/cmd/main.go`

Add unauthenticated route `GET /internal/model-configs/{function_key}` that returns the full `ModelConfig` JSON (including unmasked `api_key`). This route is only accessible within the Docker network (no external exposure).

### 2. GatewayConfigClient in policy-crawler

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**File:** `services/policy-crawler/internal/config/gateway_client.go` (new)

HTTP client that:
- Calls `GET {LLM_GATEWAY_URL}/internal/model-configs/{function_key}`
- Caches responses with 60s TTL
- Provides typed accessors:
  - `GetLLMConfig(ctx) -> llm.Config` (maps `llm_extract` -> llm.Config with backup)
  - `GetEmbeddingConfig(ctx) -> EmbeddingConfig` (maps `embedding` -> api_key, endpoint, model, dimensions)
  - `GetASRConfig(ctx) -> ASRConfig` (maps `asr` -> full ASRConfig with extra_params)

### Config mapping

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

**llm_extract -> llm.Config:**
| ModelConfig field | llm.Config field |
|---|---|
| provider | Provider (via llm.ParseProvider) |
| api_key | APIKey |
| api_endpoint | Endpoint |
| model_id | ModelName |
| max_tokens | MaxTokens |
| enabled | Enabled |
| backup_provider | backup.Provider |
| backup_api_key | backup.APIKey |
| backup_api_endpoint | backup.Endpoint |
| backup_model_id | backup.ModelName |

**embedding -> Embedding params:**
| ModelConfig field | Embedding field |
|---|---|
| api_key | apiKey |
| api_endpoint | baseURL |
| model_id | model |
| extra_params.dimensions | dimensions |

**asr -> ASRConfig:**
| ModelConfig field | ASRConfig field |
|---|---|
| api_key | APIKey |
| api_endpoint | Endpoint |
| provider | Provider |
| model_id | (used for provider selection) |
| extra_params.app_id | AppID |
| extra_params.resource_id | ResourceID |
| extra_params.language | Language |

## Changes by file

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

### llm-gateway

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Change |
|---|---|
| `cmd/main.go` | Add `GET /internal/model-configs/{function_key}` route |
| `internal/admin/admin.go` | Add `GetModelConfigInternal` handler (no auth, full API key) |

### policy-crawler

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

| File | Change |
|---|---|
| `internal/config/gateway_client.go` | New file: GatewayConfigClient with cache |
| `cmd/main.go` | Create GatewayConfigClient at startup, pass to managers/handlers. Remove local DB config reads. |
| `internal/admin/admin_llm.go` | LLMExtractRunHandler: use gatewayClient.GetLLMConfig() instead of llmStore.GetLLMConfig() |
| `internal/crawler/store.go` | RunExtraction(): use gatewayClient.GetLLMConfig() instead of local DB read |
| `internal/admin/admin_page.go` | Replace LLM/ASR config forms with link to llm-gateway admin UI |

## What stays unchanged

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- `llm.Client`, `ASRProvider`, `EmbeddingProvider` - call provider APIs directly
- LLM extraction logic, ASR transcription logic, embedding logic
- `llm_configs` / `asr_configs` tables - kept but no longer read by code
- Docker compose, networking

## Auth approach

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

Internal endpoint `/internal/model-configs/{function_key}` is unauthenticated. Security relies on Docker network isolation 鈥?only containers within the same Docker network can reach llm-gateway's port 39404. No admin credentials needed in crawler config.

## Risks

| **版本号** | V1.0.0 |
| **状态** | 已生效 |
| **发布日期** | 2026-06-15 |

- **llm-gateway down:** Crawler cannot read config, extraction/ASR fails. Acceptable because crawler already depends on llm-gateway for LLM calls via api-server.
- **Cache staleness:** 60s TTL means config changes take up to 60s to propagate. Acceptable for config changes that are infrequent.
