package crawler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/admin"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/extractor"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/llm"
	"github.com/trigold786/94-AI-Insurance-Design/shared/errors"
	"github.com/trigold786/94-AI-Insurance-Design/shared/models"
)

// DBStore DB 后端实现的存储
type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) (*DBStore, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &DBStore{db: db}, nil
}

func (s *DBStore) ListEnabledSources() ([]SourceConfig, error) {
	rows, err := s.db.Query(`SELECT source_id, source_name, source_url, source_level, crawl_type, interval_sec, COALESCE(region_code,''), enabled, COALESCE(proxy_url,''), COALESCE(request_delay_ms,1000), COALESCE(max_concurrent,1), COALESCE(respect_robots,true) FROM policy_sources WHERE enabled = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var cfgs []SourceConfig
	for rows.Next() {
		var c SourceConfig
		if err := rows.Scan(&c.SourceID, &c.SourceName, &c.SourceURL, &c.SourceLevel, &c.CrawlType, &c.IntervalSec, &c.RegionCode, &c.Enabled, &c.ProxyURL, &c.RequestDelayMs, &c.MaxConcurrent, &c.RespectRobots); err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, rows.Err()
}

func (s *DBStore) MarkRawTextExtractedByTitle(sourceID, title string) {
	s.db.Exec(`UPDATE policy_raw_texts SET extracted = true WHERE source_id = $1 AND title = $2`, sourceID, title)
}

func (s *DBStore) SaveRawText(sourceID, title, content, sourceURL, versionHash string) error {
	if versionHash != "" {
		var exists int
		s.db.QueryRow(`SELECT 1 FROM policy_raw_texts WHERE version_hash = $1 LIMIT 1`, versionHash).Scan(&exists)
		if exists == 1 {
			return nil
		}
	}
	_, err := s.db.Exec(
		`INSERT INTO policy_raw_texts (source_id, title, content, source_url, version_hash, fetch_time) VALUES ($1,$2,$3,$4,$5,$6)`,
		sourceID, title, content, sourceURL, versionHash, time.Now(),
	)
	return err
}

func (s *DBStore) SaveCrawlLog(sourceID string, success bool, errMsg string) {
	status := "success"
	msg := ""
	if !success {
		status = "failed"
		msg = errMsg
	}
	_, err := s.db.Exec(
		`INSERT INTO crawl_logs (source_id, status, error_message, crawled_at) VALUES ($1,$2,$3,$4)`,
		sourceID, status, msg, time.Now(),
	)
	if err != nil {
		log.Printf("[crawler] failed to save crawl log: %v", err)
	}
}

func (s *DBStore) SaveCrawlLogWithDetails(sourceID string, success bool, errMsg string, claimID string, summary string) {
	status := "success"
	msg := ""
	if !success {
		status = "failed"
		msg = errMsg
	}
	_, err := s.db.Exec(
		`INSERT INTO crawl_logs (source_id, status, error_message, crawled_at, extracted_claim_id, content_summary) VALUES ($1,$2,$3,$4,$5,$6)`,
		sourceID, status, msg, time.Now(), claimID, summary,
	)
	if err != nil {
		log.Printf("[crawler] failed to save crawl log: %v", err)
	}
}

// ClaimDB 实现
func (s *DBStore) Ingest(claim *models.PolicyClaim) error {
	condJSON := ""
	if len(claim.Conditions) > 0 {
		condJSON = string(claim.Conditions)
	}
	docJSON := ""
	if len(claim.RequiredDocuments) > 0 {
		docJSON = string(claim.RequiredDocuments)
	}

	_, err := s.db.Exec(`
		INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags,
			subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
			effective_date, expire_date, publish_date, confidence_score, status, version_number,
			conditions, required_documents, source_id, source_name, source_url, policy_url,
			policy_title, issuing_authority, document_number, application_process,
			contact_info, source_type, extraction_method, raw_text_length, split_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)`,
		claim.ClaimID, claim.PolicyID, claim.RegionCode, claim.PolicyType,
		claim.TargetGroupTags,
		claim.SubsidyCalcMethod, claim.SubsidyAmountMin, claim.SubsidyAmountMax,
		claim.SubsidyDuration, claim.EffectiveDate, claim.ExpireDate, claim.PublishDate,
		claim.ConfidenceScore, claim.Status, claim.VersionNumber,
		condJSON, docJSON,
		claim.SourceID, claim.SourceName, claim.SourceURL, claim.PolicyURL,
		claim.PolicyTitle, claim.IssuingAuthority, claim.DocumentNumber, claim.ApplicationProcess,
		claim.ContactInfo, claim.SourceType, claim.ExtractionMethod, claim.RawTextLength, claim.SplitCount,
	)
	return err
}

func (s *DBStore) ListByRegionAndType(regionCode, policyType string) ([]models.PolicyClaim, error) {
	rows, err := s.db.Query(`
		SELECT claim_id, policy_id, region_code, policy_type, status, confidence_score
		FROM policy_claims WHERE region_code = $1 AND policy_type = $2`,
		regionCode, policyType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		var c models.PolicyClaim
		if err := rows.Scan(&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, &c.Status, &c.ConfidenceScore); err != nil {
			return nil, err
		}
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

// admin.ClaimStore 接口方法
func (s *DBStore) ListByStatus(status string, regionCode string, sourceID string, policyType string, sourceLevel string) ([]models.PolicyClaim, error) {
	query := `SELECT pc.claim_id, pc.policy_id, pc.region_code, pc.policy_type, pc.target_group_tags,
		pc.subsidy_calc_method, pc.subsidy_amount_min, pc.subsidy_amount_max, pc.subsidy_duration,
		pc.effective_date, pc.expire_date, pc.confidence_score, pc.status, pc.version_number,
		COALESCE(pc.conditions::text,''), COALESCE(pc.required_documents::text,''),
		COALESCE(pc.source_id,''), COALESCE(pc.source_name,''), COALESCE(pc.source_url,''), COALESCE(pc.policy_url,''),
		COALESCE(pc.policy_title,''), COALESCE(pc.issuing_authority,''), COALESCE(pc.document_number,''),
		COALESCE(pc.application_process::text,''), COALESCE(pc.contact_info,''), COALESCE(pc.source_type,''),
		COALESCE(pc.extraction_method,'full'), COALESCE(pc.raw_text_length,0), COALESCE(pc.split_count,0)
		FROM policy_claims pc`
	if sourceLevel != "" {
		query += ` JOIN policy_sources ps ON ps.source_id = pc.source_id`
	}
	var args []interface{}
	var conditions []string
	argIdx := 0
	if status != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`pc.status = $%d`, argIdx))
		args = append(args, status)
	}
	if regionCode != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`pc.region_code = $%d`, argIdx))
		args = append(args, regionCode)
	}
	if sourceID != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`pc.source_id = $%d`, argIdx))
		args = append(args, sourceID)
	}
	if policyType != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`pc.policy_type = $%d`, argIdx))
		args = append(args, policyType)
	}
	if sourceLevel != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`ps.source_level = $%d`, argIdx))
		args = append(args, sourceLevel)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	query += ` ORDER BY pc.updated_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to list claims: %v", err))
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		var c models.PolicyClaim
		tags := []string{}
		var condStr, docStr, appProcStr string
		err := rows.Scan(
			&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, pq.Array(&tags),
			&c.SubsidyCalcMethod, &c.SubsidyAmountMin, &c.SubsidyAmountMax,
			&c.SubsidyDuration, &c.EffectiveDate, &c.ExpireDate,
			&c.ConfidenceScore, &c.Status, &c.VersionNumber,
			&condStr, &docStr,
			&c.SourceID, &c.SourceName, &c.SourceURL, &c.PolicyURL,
			&c.PolicyTitle, &c.IssuingAuthority, &c.DocumentNumber,
			&appProcStr, &c.ContactInfo, &c.SourceType,
			&c.ExtractionMethod, &c.RawTextLength, &c.SplitCount,
		)
		if err != nil {
			return nil, errors.NewInternal(fmt.Sprintf("failed to scan claim: %v", err))
		}
		if condStr != "" {
			c.Conditions = []byte(condStr)
		}
		if docStr != "" {
			c.RequiredDocuments = []byte(docStr)
		}
		if appProcStr != "" {
			c.ApplicationProcess = []byte(appProcStr)
		}
		c.TargetGroupTags = tags
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (s *DBStore) GetSourceByID(sourceID string) (*admin.SourceInfo, error) {
	row := s.db.QueryRow(`
		SELECT ps.source_id, ps.source_name, ps.source_url, ps.source_level,
			COALESCE(ps.crawl_type,'govsite'), COALESCE(ps.interval_sec,86400), COALESCE(ps.region_code,''), ps.enabled,
			COALESCE((SELECT MAX(cl.crawled_at)::text FROM crawl_logs cl WHERE cl.source_id = ps.source_id), ''),
			COALESCE((SELECT cl.status FROM crawl_logs cl WHERE cl.source_id = ps.source_id ORDER BY cl.crawled_at DESC LIMIT 1), ''),
			(SELECT COUNT(*) FROM policy_claims pc WHERE pc.source_id = ps.source_id),
			COALESCE(ps.proxy_url,''), COALESCE(ps.request_delay_ms,0), COALESCE(ps.max_concurrent,1), COALESCE(ps.respect_robots,true)
		FROM policy_sources ps WHERE ps.source_id = $1`, sourceID)
	var src admin.SourceInfo
	if err := row.Scan(&src.SourceID, &src.SourceName, &src.SourceURL, &src.SourceLevel,
		&src.CrawlType, &src.IntervalSec, &src.RegionCode, &src.Enabled,
		&src.LastCrawl, &src.LastStatus, &src.ClaimsCount,
		&src.ProxyURL, &src.RequestDelayMs, &src.MaxConcurrent, &src.RespectRobots); err != nil {
		return nil, err
	}
	return &src, nil
}

func (s *DBStore) ListAllSources() ([]admin.SourceInfo, error) {
	rows, err := s.db.Query(`
		SELECT ps.source_id, ps.source_name, ps.source_url, ps.source_level,
			COALESCE(ps.crawl_type,'govsite'), COALESCE(ps.interval_sec,86400), COALESCE(ps.region_code,''), ps.enabled,
			COALESCE((SELECT MAX(cl.crawled_at)::text FROM crawl_logs cl WHERE cl.source_id = ps.source_id), ''),
			COALESCE((SELECT cl.status FROM crawl_logs cl WHERE cl.source_id = ps.source_id ORDER BY cl.crawled_at DESC LIMIT 1), ''),
			(SELECT COUNT(*) FROM policy_claims pc WHERE pc.policy_id LIKE ps.source_id || '%'),
			COALESCE(ps.proxy_url,''), COALESCE(ps.request_delay_ms,0), COALESCE(ps.max_concurrent,1), COALESCE(ps.respect_robots,true)
		FROM policy_sources ps ORDER BY ps.source_level, ps.source_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []admin.SourceInfo
	for rows.Next() {
		var s admin.SourceInfo
		if err := rows.Scan(&s.SourceID, &s.SourceName, &s.SourceURL, &s.SourceLevel,
			&s.CrawlType, &s.IntervalSec, &s.RegionCode, &s.Enabled,
			&s.LastCrawl, &s.LastStatus, &s.ClaimsCount,
			&s.ProxyURL, &s.RequestDelayMs, &s.MaxConcurrent, &s.RespectRobots); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (s *DBStore) UpdateSource(id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	var sets []string
	var args []interface{}
	argIdx := 1
	allowed := map[string]string{
		"enabled": "boolean", "interval_sec": "int",
		"source_name": "text", "source_url": "text",
		"source_level": "text", "crawl_type": "text", "region_code": "text",
		"proxy_url": "text", "request_delay_ms": "int", "max_concurrent": "int", "respect_robots": "boolean",
	}
	for col := range allowed {
		val, ok := updates[col]
		if !ok {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argIdx))
		args = append(args, val)
		argIdx++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE policy_sources SET %s WHERE source_id = $%d",
		strings.Join(sets, ", "), argIdx)
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *DBStore) CreateSource(src *admin.SourceInfo) error {
	_, err := s.db.Exec(`INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code, proxy_url, request_delay_ms, max_concurrent, respect_robots)
		VALUES ($1, $2, $3, $4, 0.7, true, $5, $6, $7, $8, $9, $10, $11)`,
		src.SourceID, src.SourceName, src.SourceURL, src.SourceLevel,
		src.CrawlType, src.IntervalSec, src.RegionCode,
		src.ProxyURL, src.RequestDelayMs, src.MaxConcurrent, src.RespectRobots)
	return err
}

func (s *DBStore) DeleteSource(sourceID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM crawl_logs WHERE source_id = $1`, sourceID)
	tx.Exec(`DELETE FROM policy_raw_texts WHERE source_id = $1`, sourceID)
	tx.Exec(`DELETE FROM extract_logs WHERE source_id = $1`, sourceID)
	tx.Exec(`UPDATE policy_claims SET source_id = '' WHERE source_id = $1`, sourceID)
	_, err = tx.Exec(`DELETE FROM policy_sources WHERE source_id = $1`, sourceID)
	if err != nil {
		return fmt.Errorf("delete source: %w", err)
	}
	return tx.Commit()
}

func (s *DBStore) GetCrawlLogsFiltered(startDate, endDate, sourceType, sourceLevel, status string, limit int) ([]admin.CrawlLogEntry, error) {
	query := `SELECT cl.id, cl.source_id, COALESCE(ps.source_name, cl.source_id), cl.status, cl.error_message, cl.crawled_at::text,
		COALESCE(cl.extracted_claim_id,''), COALESCE(cl.content_summary,'')
		FROM crawl_logs cl LEFT JOIN policy_sources ps ON cl.source_id = ps.source_id WHERE 1=1`
	var args []interface{}
	argIdx := 1
	if startDate != "" {
		query += fmt.Sprintf(" AND cl.crawled_at >= $%d::timestamp", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND cl.crawled_at <= $%d::timestamp + interval '1 day'", argIdx)
		args = append(args, endDate)
		argIdx++
	}
	if sourceType != "" {
		query += fmt.Sprintf(" AND ps.crawl_type = $%d", argIdx)
		args = append(args, sourceType)
		argIdx++
	}
	if sourceLevel != "" {
		query += fmt.Sprintf(" AND ps.source_level = $%d", argIdx)
		args = append(args, sourceLevel)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND cl.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY cl.crawled_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []admin.CrawlLogEntry
	for rows.Next() {
		var l admin.CrawlLogEntry
		if err := rows.Scan(&l.ID, &l.SourceID, &l.SourceName, &l.Status, &l.ErrorMessage, &l.CrawledAt,
			&l.ExtractedClaimID, &l.ContentSummary); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *DBStore) GetCrawlLogs(limit int) ([]admin.CrawlLogEntry, error) {
	rows, err := s.db.Query(`
		SELECT cl.id, cl.source_id, COALESCE(ps.source_name, cl.source_id), cl.status, cl.error_message, cl.crawled_at::text,
		COALESCE(cl.extracted_claim_id,''), COALESCE(cl.content_summary,'')
		FROM crawl_logs cl LEFT JOIN policy_sources ps ON cl.source_id = ps.source_id
		ORDER BY cl.crawled_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []admin.CrawlLogEntry
	for rows.Next() {
		var l admin.CrawlLogEntry
		if err := rows.Scan(&l.ID, &l.SourceID, &l.SourceName, &l.Status, &l.ErrorMessage, &l.CrawledAt,
			&l.ExtractedClaimID, &l.ContentSummary); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *DBStore) GetDashboardStats() (*admin.DashboardStats, error) {
	stats := &admin.DashboardStats{
		PolicyTypeDist: make(map[string]int),
		RegionDist:     make(map[string]int),
		CrawlTrend7d:   make([]int, 7),
	}

	s.db.QueryRow(`SELECT COUNT(*) FROM policy_sources`).Scan(&stats.TotalSources)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_sources WHERE enabled = true`).Scan(&stats.ActiveSources)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_sources WHERE source_level = 'HIGH'`).Scan(&stats.HighSources)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_sources WHERE source_level = 'MEDIUM'`).Scan(&stats.MediumSources)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_sources WHERE source_level = 'LOW'`).Scan(&stats.LowSources)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims`).Scan(&stats.TotalClaims)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims WHERE status = 'verified'`).Scan(&stats.VerifiedClaims)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims WHERE status = 'pending_review'`).Scan(&stats.PendingClaims)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims WHERE status = 'unverified'`).Scan(&stats.UnverifiedClaims)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims WHERE embedding IS NOT NULL`).Scan(&stats.WithEmbedding)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_claims WHERE policy_url != ''`).Scan(&stats.WithPolicyURL)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_raw_texts WHERE NOT extracted`).Scan(&stats.PendingExtraction)
	s.db.QueryRow(`SELECT COUNT(*) FROM crawl_logs WHERE crawled_at >= NOW() - INTERVAL '24 hours'`).Scan(&stats.TodayCrawls)
	s.db.QueryRow(`SELECT COUNT(*) FROM crawl_logs WHERE crawled_at >= NOW() - INTERVAL '24 hours' AND status = 'failed'`).Scan(&stats.FailedCrawls)

	// 提取成功率
	var totalExtracts, successExtracts int
	s.db.QueryRow(`SELECT COUNT(*) FROM extract_logs`).Scan(&totalExtracts)
	s.db.QueryRow(`SELECT COUNT(*) FROM extract_logs WHERE status = 'success'`).Scan(&successExtracts)
	if totalExtracts > 0 {
		stats.ExtractSuccessRate = float64(successExtracts) / float64(totalExtracts) * 100
	}

	// 政策类型分布
	rows, err := s.db.Query(`SELECT policy_type, COUNT(*) FROM policy_claims GROUP BY policy_type ORDER BY count DESC`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t string
			var c int
			rows.Scan(&t, &c)
			stats.PolicyTypeDist[t] = c
		}
	}

	// 城市分布
	rows2, err := s.db.Query(`SELECT region_code, COUNT(*) FROM policy_claims WHERE region_code != '' GROUP BY region_code ORDER BY count DESC LIMIT 10`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var r string
			var c int
			rows2.Scan(&r, &c)
			stats.RegionDist[r] = c
		}
	}

	// 近7天爬取趋势
	for i := 0; i < 7; i++ {
		s.db.QueryRow(`SELECT COUNT(*) FROM crawl_logs WHERE crawled_at >= NOW() - INTERVAL '1 day' * $1 AND crawled_at < NOW() - INTERVAL '1 day' * $2`,
			i, i-1).Scan(&stats.CrawlTrend7d[6-i])
	}

	return stats, nil
}

// --- LLM 配置 ---
func (s *DBStore) GetLLMConfig() (*admin.LLMConfig, error) {
	var cfg admin.LLMConfig
	err := s.db.QueryRow(`SELECT provider, api_key, endpoint, model_name, max_tokens, enabled,
		COALESCE(NULLIF(embedding_model,''), 'doubao-embedding-vision'), COALESCE(NULLIF(embedding_dimensions::text,''), '1024')::int,
		COALESCE(embedding_api_key, ''), COALESCE(embedding_endpoint, ''),
		COALESCE(backup_provider, ''), COALESCE(backup_api_key, ''), COALESCE(backup_endpoint, ''), COALESCE(backup_model_name, '')
		FROM llm_configs ORDER BY id DESC LIMIT 1`).Scan(
		&cfg.Provider, &cfg.APIKey, &cfg.Endpoint, &cfg.ModelName, &cfg.MaxTokens, &cfg.Enabled,
		&cfg.EmbeddingModel, &cfg.EmbeddingDimensions,
		&cfg.EmbeddingAPIKey, &cfg.EmbeddingEndpoint,
		&cfg.BackupProvider, &cfg.BackupAPIKey, &cfg.BackupEndpoint, &cfg.BackupModelName)
	if err != nil {
		return nil, fmt.Errorf("query llm config: %w", err)
	}
	return &cfg, nil
}

func (s *DBStore) SaveLLMConfig(cfg *admin.LLMConfig) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`DELETE FROM llm_configs`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO llm_configs (provider, api_key, endpoint, model_name, max_tokens, enabled, embedding_model, embedding_dimensions, embedding_api_key, embedding_endpoint, backup_provider, backup_api_key, backup_endpoint, backup_model_name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		cfg.Provider, cfg.APIKey, cfg.Endpoint, cfg.ModelName, cfg.MaxTokens, cfg.Enabled,
		cfg.EmbeddingModel, cfg.EmbeddingDimensions, cfg.EmbeddingAPIKey, cfg.EmbeddingEndpoint,
		cfg.BackupProvider, cfg.BackupAPIKey, cfg.BackupEndpoint, cfg.BackupModelName)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DBStore) GetASRConfig() (ASRConfig, error) {
	var cfg ASRConfig
	err := s.db.QueryRow(`SELECT id, provider, api_key, app_id, endpoint, resource_id, language, sample_rate, max_wait_seconds, poll_interval_seconds, enabled FROM asr_configs ORDER BY id LIMIT 1`).
		Scan(&cfg.ID, &cfg.Provider, &cfg.APIKey, &cfg.AppID, &cfg.Endpoint, &cfg.ResourceID, &cfg.Language, &cfg.SampleRate, &cfg.MaxWaitSeconds, &cfg.PollIntervalSeconds, &cfg.Enabled)
	if err != nil {
		return ASRConfig{}, err
	}
	return cfg, nil
}

func (s *DBStore) SaveASRConfig(cfg *ASRConfig) error {
	_, err := s.db.Exec(`UPDATE asr_configs SET provider=$1, api_key=$2, app_id=$3, endpoint=$4, resource_id=$5, language=$6, sample_rate=$7, max_wait_seconds=$8, poll_interval_seconds=$9, enabled=$10, updated_at=now() WHERE id=$11`,
		cfg.Provider, cfg.APIKey, cfg.AppID, cfg.Endpoint, cfg.ResourceID, cfg.Language, cfg.SampleRate, cfg.MaxWaitSeconds, cfg.PollIntervalSeconds, cfg.Enabled, cfg.ID)
	return err
}

func (s *DBStore) SaveRawTextReturningID(sourceID, title, content, sourceURL, versionHash string) (int64, error) {
	if versionHash != "" {
		var existingID int64
		err := s.db.QueryRow(`SELECT id FROM policy_raw_texts WHERE version_hash = $1 LIMIT 1`, versionHash).Scan(&existingID)
		if err == nil {
			return existingID, nil
		}
	}
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO policy_raw_texts (source_id, title, content, source_url, version_hash, fetch_time) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		sourceID, title, content, sourceURL, versionHash, time.Now(),
	).Scan(&id)
	return id, err
}

func (s *DBStore) SetVideoExtractStatus(id int64, status string) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET video_extract_status = $1 WHERE id = $2`, status, id)
	return err
}

func (s *DBStore) UpdateRawTextContent(id int64, content string) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET content = $1 WHERE id = $2`, content, id)
	return err
}

func (s *DBStore) MarkExtractedByID(id int64) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET extracted = true WHERE id = $1`, id)
	return err
}

type PendingVideoExtract struct {
	ID       int64
	SourceID string
	VideoURL string
	Title    string
	Content  string
}

func (s *DBStore) GetPendingVideoExtracts() ([]PendingVideoExtract, error) {
	rows, err := s.db.Query(`SELECT id, source_id, source_url, COALESCE(title,''), COALESCE(content,'') FROM policy_raw_texts WHERE video_extract_status IN ('pending','processing')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PendingVideoExtract
	for rows.Next() {
		var p PendingVideoExtract
		if err := rows.Scan(&p.ID, &p.SourceID, &p.VideoURL, &p.Title, &p.Content); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// --- LLM 提取 ---
func (s *DBStore) GetUnprocessedRawTexts(limit int) ([]extractor.RawTextEntry, error) {
	rows, err := s.db.Query(`SELECT prt.id, prt.source_id, prt.content, COALESCE(prt.source_url,''), COALESCE(ps.source_name,''),
		COALESCE(prt.title,'')
		FROM policy_raw_texts prt
		LEFT JOIN policy_sources ps ON ps.source_id = prt.source_id
		LEFT JOIN (
			SELECT raw_text_id, COUNT(*) AS fail_count
			FROM extract_logs
			WHERE status = 'failed' AND raw_text_id > 0
			GROUP BY raw_text_id
		) ef ON ef.raw_text_id = prt.id
		WHERE NOT prt.extracted AND LENGTH(prt.content) >= 500
		  AND (prt.video_extract_status IS NULL OR prt.video_extract_status = 'done')
		  AND (ef.fail_count IS NULL OR ef.fail_count < 3)
		ORDER BY prt.id ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []extractor.RawTextEntry
	for rows.Next() {
		var e extractor.RawTextEntry
		if err := rows.Scan(&e.ID, &e.SourceID, &e.Content, &e.SourceURL, &e.SourceName, &e.Title); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *DBStore) MarkExtracted(id int64, claimID string) error {
	_, err := s.db.Exec(`UPDATE policy_raw_texts SET extracted=TRUE, extracted_claim_id=$1 WHERE id=$2`, claimID, id)
	return err
}

func (s *DBStore) InsertClaim(claim *models.PolicyClaim) error {
	condJSON := ""
	if len(claim.Conditions) > 0 {
		condJSON = string(claim.Conditions)
	}
	docJSON := ""
	if len(claim.RequiredDocuments) > 0 {
		docJSON = string(claim.RequiredDocuments)
	}
	appProcJSON := ""
	if len(claim.ApplicationProcess) > 0 {
		appProcJSON = string(claim.ApplicationProcess)
	}
	_, err := s.db.Exec(`
		INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags,
			subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
			effective_date, expire_date, publish_date, confidence_score, status, version_number,
			conditions, required_documents, source_id, source_name, source_url, policy_url,
			policy_title, issuing_authority, document_number, application_process,
			contact_info, source_type, extraction_method, raw_text_length, split_count)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)`,
		claim.ClaimID, claim.PolicyID, claim.RegionCode, claim.PolicyType,
		pq.Array(claim.TargetGroupTags),
		claim.SubsidyCalcMethod, claim.SubsidyAmountMin, claim.SubsidyAmountMax,
		claim.SubsidyDuration, claim.EffectiveDate, claim.ExpireDate, claim.PublishDate,
		claim.ConfidenceScore, claim.Status, claim.VersionNumber,
		condJSON, docJSON,
		claim.SourceID, claim.SourceName, claim.SourceURL, claim.PolicyURL,
		claim.PolicyTitle, claim.IssuingAuthority, claim.DocumentNumber, appProcJSON,
		claim.ContactInfo, claim.SourceType, claim.ExtractionMethod, claim.RawTextLength, claim.SplitCount)
	return err
}

func (s *DBStore) SaveExtractLog(sourceID string, success bool, msg string) {
	status := "failed"
	if success {
		status = "success"
	}
	s.db.Exec(`INSERT INTO extract_logs (source_id, status, message) VALUES ($1,$2,$3)`, sourceID, status, msg)
}

func (s *DBStore) SaveExtractLogDetailed(rawTextID int64, sourceID string, success bool, msg string, claimID string, title string, modelName string, summary string) {
	status := "failed"
	if success {
		status = "success"
	}
	s.db.Exec(`INSERT INTO extract_logs (source_id, status, message, raw_text_id, claim_id, title, model_name, content_summary) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		sourceID, status, msg, rawTextID, claimID, title, modelName, summary)
}

func (s *DBStore) GetExtractLogsFiltered(startDate, endDate, sourceType, sourceLevel, regionCode, status string, limit int) ([]admin.ExtractLogEntry, error) {
	query := `SELECT el.id, el.raw_text_id, el.source_id, COALESCE(ps.source_name, el.source_id),
		COALESCE(el.claim_id,''), COALESCE(el.title,''), el.status, el.message,
		COALESCE(el.model_name,''), el.created_at::text, COALESCE(el.content_summary,'')
		FROM extract_logs el LEFT JOIN policy_sources ps ON el.source_id = ps.source_id WHERE 1=1`
	var args []interface{}
	argIdx := 1
	if startDate != "" {
		query += fmt.Sprintf(" AND el.created_at >= $%d::timestamp", argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		query += fmt.Sprintf(" AND el.created_at <= $%d::timestamp + interval '1 day'", argIdx)
		args = append(args, endDate)
		argIdx++
	}
	if sourceType != "" {
		query += fmt.Sprintf(" AND ps.crawl_type = $%d", argIdx)
		args = append(args, sourceType)
		argIdx++
	}
	if sourceLevel != "" {
		query += fmt.Sprintf(" AND ps.source_level = $%d", argIdx)
		args = append(args, sourceLevel)
		argIdx++
	}
	if regionCode != "" {
		query += fmt.Sprintf(" AND ps.region_code = $%d", argIdx)
		args = append(args, regionCode)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND el.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY el.created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []admin.ExtractLogEntry
	for rows.Next() {
		var l admin.ExtractLogEntry
		if err := rows.Scan(&l.ID, &l.RawTextID, &l.SourceID, &l.SourceName,
			&l.ClaimID, &l.Title, &l.Status, &l.ErrorMessage,
			&l.ModelName, &l.CreatedAt, &l.ContentSummary); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *DBStore) SaveEmbedding(claimID string, embedding []float64) error {
	if embedding == nil {
		return nil
	}
	f32 := make([]float32, len(embedding))
	for i, v := range embedding {
		f32[i] = float32(v)
	}
	vec := pgvector.NewVector(f32)
	_, err := s.db.Exec(`UPDATE policy_claims SET embedding = $1 WHERE claim_id = $2`, vec, claimID)
	return err
}

// SaveSnapshot 保存政策版本快照
func (s *DBStore) SaveSnapshot(claim *models.PolicyClaim) error {
	snapshot, _ := json.Marshal(claim)
	_, err := s.db.Exec(`
		INSERT INTO policy_snapshots (claim_id, policy_id, version_number, snapshot_data, superseded_by)
		VALUES ($1, $2, $3, $4, '')`,
		claim.ClaimID, claim.PolicyID, claim.VersionNumber, snapshot)
	return err
}

// MarkSuperseded 标记旧版本已被新版本替代
func (s *DBStore) MarkSuperseded(oldClaimID, newClaimID string) error {
	_, err := s.db.Exec(`UPDATE policy_snapshots SET superseded_by = $1 WHERE claim_id = $2`,
		newClaimID, oldClaimID)
	return err
}

// GetMaxVersionNumber 获取指定 policy_id 的最大版本号
func (s *DBStore) GetMaxVersionNumber(policyID string) (int, error) {
	var maxVer int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(version_number), 0) FROM policy_snapshots WHERE policy_id = $1`, policyID).Scan(&maxVer)
	return maxVer, err
}

// GetLatestClaimByPolicyID 获取指定 policy_id 最新版本的 claim_id
func (s *DBStore) GetLatestClaimByPolicyID(policyID string) (string, error) {
	var claimID string
	err := s.db.QueryRow(`SELECT claim_id FROM policy_snapshots WHERE policy_id = $1 ORDER BY version_number DESC LIMIT 1`, policyID).Scan(&claimID)
	return claimID, err
}

// GetVersionAtTime 获取指定 policy_id 在某时间点之前的最新版本
func (s *DBStore) GetVersionAtTime(policyID string, timestamp string) (*models.VersionSnapshot, error) {
	var vs models.VersionSnapshot
	err := s.db.QueryRow(`
		SELECT id, claim_id, policy_id, version_number, snapshot_data,
		       COALESCE(superseded_by, ''), created_at::text
		FROM policy_snapshots
		WHERE policy_id = $1 AND created_at <= $2::timestamptz
		ORDER BY version_number DESC LIMIT 1`, policyID, timestamp).
		Scan(&vs.ID, &vs.ClaimID, &vs.PolicyID, &vs.VersionNumber,
			&vs.SnapshotData, &vs.SupersededBy, &vs.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}

// ListVersions 查询指定 policy_id 的所有版本历史
func (s *DBStore) ListVersions(policyID string) ([]models.VersionSnapshot, error) {
	rows, err := s.db.Query(`
		SELECT id, claim_id, policy_id, version_number, snapshot_data,
		       COALESCE(superseded_by, ''), created_at::text
		FROM policy_snapshots WHERE policy_id = $1
		ORDER BY version_number DESC`, policyID)
	if err != nil {
		return nil, fmt.Errorf("query versions: %w", err)
	}
	defer rows.Close()

	var result []models.VersionSnapshot
	for rows.Next() {
		var vs models.VersionSnapshot
		if err := rows.Scan(&vs.ID, &vs.ClaimID, &vs.PolicyID, &vs.VersionNumber,
			&vs.SnapshotData, &vs.SupersededBy, &vs.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan version: %w", err)
		}
		result = append(result, vs)
	}
	return result, rows.Err()
}

func (s *DBStore) GetUnprocessedCount() (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM policy_raw_texts WHERE NOT extracted`).Scan(&n)
	return n, err
}

func (s *DBStore) GetPendingRawTexts(limit int) ([]admin.PendingRawText, error) {
	rows, err := s.db.Query(`SELECT prt.id, prt.source_id, COALESCE(ps.source_name,''),
		COALESCE(prt.title,''), COALESCE(prt.source_url,''), prt.fetch_time::text,
		COALESCE(ps.crawl_type,''), COALESCE(ps.source_level,''), COALESCE(ps.region_code,'')
		FROM policy_raw_texts prt
		LEFT JOIN policy_sources ps ON ps.source_id = prt.source_id
		WHERE NOT prt.extracted ORDER BY prt.id ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []admin.PendingRawText
	for rows.Next() {
		var item admin.PendingRawText
		if err := rows.Scan(&item.ID, &item.SourceID, &item.SourceName, &item.Title, &item.SourceURL, &item.FetchedAt,
			&item.CrawlType, &item.SourceLevel, &item.RegionCode); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *DBStore) RunExtraction(limit int) (int, int, error) {
	cfg, err := s.GetLLMConfig()
	if err != nil {
		return 0, 0, fmt.Errorf("llm config: %w", err)
	}

	client := llm.NewClient(llm.Config{
		Provider:  llm.ParseProvider(cfg.Provider),
		APIKey:    cfg.APIKey,
		Endpoint:  cfg.Endpoint,
		ModelName: cfg.ModelName,
		MaxTokens: cfg.MaxTokens,
		Enabled:   cfg.Enabled,
	})

	ext := extractor.NewExtractor(s, client)
	return ext.ProcessUnprocessed(limit)
}

func (s *DBStore) UpdateStatus(claimID, status string, confidence float64) error {
	result, err := s.db.Exec(`UPDATE policy_claims SET status = $1, confidence_score = $2 WHERE claim_id = $3`, status, confidence, claimID)
	if err != nil {
		return fmt.Errorf("failed to update claim status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return errors.NewNotFound("claim", claimID)
	}

	var sourceID string
	s.db.QueryRow(`SELECT source_id FROM policy_claims WHERE claim_id = $1`, claimID).Scan(&sourceID)
	if sourceID != "" {
		var currentWeight float64
		s.db.QueryRow(`SELECT weight FROM policy_sources WHERE source_id = $1`, sourceID).Scan(&currentWeight)
		if currentWeight > 0 {
			var newWeight float64
			switch status {
			case "verified":
				newWeight = currentWeight + 0.02
				if newWeight > 1.0 {
					newWeight = 1.0
				}
			case "unverified":
				newWeight = currentWeight - 0.05
				if newWeight < 0.1 {
					newWeight = 0.1
				}
			default:
				return nil
			}
			if newWeight != currentWeight {
				s.db.Exec(`UPDATE policy_sources SET weight = $1 WHERE source_id = $2`, newWeight, sourceID)
				log.Printf("[store] source %s weight adjusted: %.2f -> %.2f (claim %s -> %s)", sourceID, currentWeight, newWeight, claimID, status)
			}
		}
	}
	return nil
}

func (s *DBStore) GetPipeline() ([]admin.PipelineEntry, error) {
	query := `
	SELECT
		ps.source_id,
		COALESCE(ps.source_name, ps.source_id),
		ps.source_level,
		ps.crawl_type,
		ps.enabled,
		COALESCE(cl.crawled_at::text, ''),
		CASE WHEN cl.status IS NULL THEN 'never'
			 WHEN cl.status = 'true' THEN 'success'
			 WHEN cl.status = 'false' THEN 'failed'
			 ELSE cl.status::text END,
		COALESCE(cl.error_message, ''),
		CASE WHEN rt.id IS NOT NULL THEN true ELSE false END,
		COALESCE(el.created_at::text, ''),
		COALESCE(el.status, ''),
		COALESCE(el.message, ''),
		COALESCE(pc.claim_id, ''),
		COALESCE(pc.status, ''),
		COALESCE(pc.confidence_score, 0)
	FROM policy_sources ps
	LEFT JOIN LATERAL (
		SELECT cl.crawled_at, cl.status, cl.error_message
		FROM crawl_logs cl
		WHERE cl.source_id = ps.source_id
		ORDER BY cl.crawled_at DESC LIMIT 1
	) cl ON true
	LEFT JOIN LATERAL (
		SELECT id FROM policy_raw_texts
		WHERE source_id = ps.source_id
		LIMIT 1
	) rt ON true
	LEFT JOIN LATERAL (
		SELECT el.created_at, el.status, el.message
		FROM extract_logs el
		WHERE el.source_id = ps.source_id
		ORDER BY el.created_at DESC LIMIT 1
	) el ON true
	LEFT JOIN LATERAL (
		SELECT pc.claim_id, pc.status, pc.confidence_score
		FROM policy_claims pc
		WHERE pc.source_id = ps.source_id
		ORDER BY pc.version_number DESC LIMIT 1
	) pc ON true
	ORDER BY ps.source_id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []admin.PipelineEntry
	for rows.Next() {
		var e admin.PipelineEntry
		if err := rows.Scan(&e.SourceID, &e.SourceName, &e.SourceLevel, &e.CrawlType,
			&e.Enabled, &e.LastCrawlAt, &e.CrawlStatus, &e.CrawlError,
			&e.HasRawText, &e.LastExtractAt, &e.ExtractStatus, &e.ExtractError,
			&e.ClaimID, &e.ClaimStatus, &e.Confidence); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *DBStore) GetClaimByID(claimID string) (*models.PolicyClaim, error) {
	query := `SELECT claim_id, policy_id, region_code, policy_type, target_group_tags,
		subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
		effective_date, expire_date, COALESCE(publish_date,''), confidence_score, status, version_number,
		conditions, required_documents, source_id, source_name, source_url, policy_url,
		policy_title, issuing_authority, document_number, application_process,
		contact_info, source_type, extraction_method, raw_text_length, split_count
		FROM policy_claims WHERE claim_id = $1`

	var c models.PolicyClaim
	err := s.db.QueryRow(query, claimID).Scan(
		&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, pq.Array(&c.TargetGroupTags),
		&c.SubsidyCalcMethod, &c.SubsidyAmountMin, &c.SubsidyAmountMax, &c.SubsidyDuration,
		&c.EffectiveDate, &c.ExpireDate, &c.PublishDate, &c.ConfidenceScore, &c.Status, &c.VersionNumber,
		&c.Conditions, &c.RequiredDocuments, &c.SourceID, &c.SourceName, &c.SourceURL, &c.PolicyURL,
		&c.PolicyTitle, &c.IssuingAuthority, &c.DocumentNumber, &c.ApplicationProcess,
		&c.ContactInfo, &c.SourceType, &c.ExtractionMethod, &c.RawTextLength, &c.SplitCount,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *DBStore) SearchSimilarClaims(claimID string, limit int) ([]models.PolicyClaim, error) {
	original, err := s.GetClaimByID(claimID)
	if err != nil {
		return nil, err
	}
	query := `SELECT claim_id, policy_id, region_code, policy_type, target_group_tags,
		subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
		effective_date, expire_date, COALESCE(publish_date,''), confidence_score, status, version_number,
		conditions, required_documents, source_id, source_name, source_url, policy_url,
		policy_title, issuing_authority, document_number, application_process,
		contact_info, source_type, extraction_method, raw_text_length, split_count
		FROM policy_claims
		WHERE region_code = $1 AND policy_type = $2 AND claim_id != $3
		ORDER BY confidence_score DESC LIMIT $4`
	rows, err := s.db.Query(query, original.RegionCode, original.PolicyType, claimID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var claims []models.PolicyClaim
	for rows.Next() {
		var c models.PolicyClaim
		if err := rows.Scan(
			&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, pq.Array(&c.TargetGroupTags),
			&c.SubsidyCalcMethod, &c.SubsidyAmountMin, &c.SubsidyAmountMax, &c.SubsidyDuration,
			&c.EffectiveDate, &c.ExpireDate, &c.PublishDate, &c.ConfidenceScore, &c.Status, &c.VersionNumber,
			&c.Conditions, &c.RequiredDocuments, &c.SourceID, &c.SourceName, &c.SourceURL, &c.PolicyURL,
			&c.PolicyTitle, &c.IssuingAuthority, &c.DocumentNumber, &c.ApplicationProcess,
			&c.ContactInfo, &c.SourceType, &c.ExtractionMethod, &c.RawTextLength, &c.SplitCount,
		); err != nil {
			return nil, err
		}
		claims = append(claims, c)
	}
	return claims, rows.Err()
}

func (s *DBStore) UpdateClaimFields(claimID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	allowed := map[string]bool{
		"subsidy_calc_method": true, "effective_date": true, "expire_date": true,
		"policy_type": true, "region_code": true, "subsidy_amount_min": true,
		"subsidy_amount_max": true, "subsidy_duration": true, "policy_title": true,
		"issuing_authority": true, "document_number": true, "contact_info": true,
	}
	var setClauses []string
	var args []interface{}
	argIdx := 1
	for field, val := range fields {
		if !allowed[field] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, argIdx))
		args = append(args, val)
		argIdx++
	}
	if len(setClauses) == 0 {
		return nil
	}
	args = append(args, claimID)
	query := fmt.Sprintf("UPDATE policy_claims SET %s WHERE claim_id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	_, err := s.db.Exec(query, args...)
	return err
}
