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
	rows, err := s.db.Query(`SELECT source_id, source_name, source_url, source_level, crawl_type, interval_sec, COALESCE(region_code,''), enabled FROM policy_sources WHERE enabled = true`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var cfgs []SourceConfig
	for rows.Next() {
		var c SourceConfig
		if err := rows.Scan(&c.SourceID, &c.SourceName, &c.SourceURL, &c.SourceLevel, &c.CrawlType, &c.IntervalSec, &c.RegionCode, &c.Enabled); err != nil {
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
			effective_date, expire_date, confidence_score, status, version_number,
			conditions, required_documents, source_id, source_name, source_url, policy_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		claim.ClaimID, claim.PolicyID, claim.RegionCode, claim.PolicyType,
		claim.TargetGroupTags,
		claim.SubsidyCalcMethod, claim.SubsidyAmountMin, claim.SubsidyAmountMax,
		claim.SubsidyDuration, claim.EffectiveDate, claim.ExpireDate,
		claim.ConfidenceScore, claim.Status, claim.VersionNumber,
		condJSON, docJSON,
		claim.SourceID, claim.SourceName, claim.SourceURL, claim.PolicyURL,
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
func (s *DBStore) ListByStatus(status string, regionCode string) ([]models.PolicyClaim, error) {
	query := `SELECT claim_id, policy_id, region_code, policy_type, target_group_tags,
		subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
		effective_date, expire_date, confidence_score, status, version_number,
		COALESCE(conditions::text,''), COALESCE(required_documents::text,''),
		COALESCE(source_id,''), COALESCE(source_name,''), COALESCE(source_url,''), COALESCE(policy_url,'')
		FROM policy_claims`
	var args []interface{}
	var conditions []string
	argIdx := 0
	if status != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`status = $%d`, argIdx))
		args = append(args, status)
	}
	if regionCode != "" {
		argIdx++
		conditions = append(conditions, fmt.Sprintf(`region_code = $%d`, argIdx))
		args = append(args, regionCode)
	}
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to list claims: %v", err))
	}
	defer rows.Close()

	var claims []models.PolicyClaim
	for rows.Next() {
		var c models.PolicyClaim
		tags := []string{}
		var condStr, docStr string
		err := rows.Scan(
			&c.ClaimID, &c.PolicyID, &c.RegionCode, &c.PolicyType, pq.Array(&tags),
			&c.SubsidyCalcMethod, &c.SubsidyAmountMin, &c.SubsidyAmountMax,
			&c.SubsidyDuration, &c.EffectiveDate, &c.ExpireDate,
			&c.ConfidenceScore, &c.Status, &c.VersionNumber,
			&condStr, &docStr,
			&c.SourceID, &c.SourceName, &c.SourceURL, &c.PolicyURL,
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
			(SELECT COUNT(*) FROM policy_claims pc WHERE pc.source_id = ps.source_id)
		FROM policy_sources ps WHERE ps.source_id = $1`, sourceID)
	var src admin.SourceInfo
	if err := row.Scan(&src.SourceID, &src.SourceName, &src.SourceURL, &src.SourceLevel,
		&src.CrawlType, &src.IntervalSec, &src.RegionCode, &src.Enabled,
		&src.LastCrawl, &src.LastStatus, &src.ClaimsCount); err != nil {
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
			(SELECT COUNT(*) FROM policy_claims pc WHERE pc.policy_id LIKE ps.source_id || '%')
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
			&s.LastCrawl, &s.LastStatus, &s.ClaimsCount); err != nil {
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
	_, err := s.db.Exec(`INSERT INTO policy_sources (source_id, source_name, source_url, source_level, weight, enabled, crawl_type, interval_sec, region_code)
		VALUES ($1, $2, $3, $4, 0.7, true, $5, $6, $7)`,
		src.SourceID, src.SourceName, src.SourceURL, src.SourceLevel,
		src.CrawlType, src.IntervalSec, src.RegionCode)
	return err
}

func (s *DBStore) DeleteSource(sourceID string) error {
	_, err := s.db.Exec(`DELETE FROM policy_sources WHERE source_id = $1`, sourceID)
	return err
}

func (s *DBStore) GetCrawlLogsFiltered(startDate, endDate string, limit int) ([]admin.CrawlLogEntry, error) {
	query := `SELECT cl.id, cl.source_id, COALESCE(ps.source_name, cl.source_id), cl.status, cl.error_message, cl.crawled_at::text
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
		if err := rows.Scan(&l.ID, &l.SourceID, &l.SourceName, &l.Status, &l.ErrorMessage, &l.CrawledAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func (s *DBStore) GetCrawlLogs(limit int) ([]admin.CrawlLogEntry, error) {
	rows, err := s.db.Query(`
		SELECT cl.id, cl.source_id, COALESCE(ps.source_name, cl.source_id), cl.status, cl.error_message, cl.crawled_at::text
		FROM crawl_logs cl LEFT JOIN policy_sources ps ON cl.source_id = ps.source_id
		ORDER BY cl.crawled_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []admin.CrawlLogEntry
	for rows.Next() {
		var l admin.CrawlLogEntry
		if err := rows.Scan(&l.ID, &l.SourceID, &l.SourceName, &l.Status, &l.ErrorMessage, &l.CrawledAt); err != nil {
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
		COALESCE(embedding_model, 'text-embedding-3-small'), COALESCE(embedding_dimensions, 1536)
		FROM llm_configs ORDER BY id DESC LIMIT 1`).Scan(
		&cfg.Provider, &cfg.APIKey, &cfg.Endpoint, &cfg.ModelName, &cfg.MaxTokens, &cfg.Enabled,
		&cfg.EmbeddingModel, &cfg.EmbeddingDimensions)
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
	_, err = tx.Exec(`INSERT INTO llm_configs (provider, api_key, endpoint, model_name, max_tokens, enabled) VALUES ($1,$2,$3,$4,$5,$6)`,
		cfg.Provider, cfg.APIKey, cfg.Endpoint, cfg.ModelName, cfg.MaxTokens, cfg.Enabled)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// --- LLM 提取 ---
func (s *DBStore) GetUnprocessedRawTexts(limit int) ([]extractor.RawTextEntry, error) {
	rows, err := s.db.Query(`SELECT prt.id, prt.source_id, prt.content, COALESCE(prt.source_url,''), COALESCE(ps.source_name,'')
		FROM policy_raw_texts prt
		LEFT JOIN policy_sources ps ON ps.source_id = prt.source_id
		WHERE NOT prt.extracted ORDER BY prt.id ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []extractor.RawTextEntry
	for rows.Next() {
		var e extractor.RawTextEntry
		if err := rows.Scan(&e.ID, &e.SourceID, &e.Content, &e.SourceURL, &e.SourceName); err != nil {
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
	_, err := s.db.Exec(`
		INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags,
			subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
			effective_date, expire_date, confidence_score, status, version_number,
			conditions, required_documents, source_id, source_name, source_url, policy_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		claim.ClaimID, claim.PolicyID, claim.RegionCode, claim.PolicyType,
		pq.Array(claim.TargetGroupTags),
		claim.SubsidyCalcMethod, claim.SubsidyAmountMin, claim.SubsidyAmountMax,
		claim.SubsidyDuration, claim.EffectiveDate, claim.ExpireDate,
		claim.ConfidenceScore, claim.Status, claim.VersionNumber,
		condJSON, docJSON,
		claim.SourceID, claim.SourceName, claim.SourceURL, claim.PolicyURL)
	return err
}

func (s *DBStore) SaveExtractLog(sourceID string, success bool, msg string) {
	status := "failed"
	if success {
		status = "success"
	}
	s.db.Exec(`INSERT INTO extract_logs (source_id, status, message) VALUES ($1,$2,$3)`, sourceID, status, msg)
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
		COALESCE(prt.title,''), COALESCE(prt.source_url,''), prt.fetch_time::text
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
		if err := rows.Scan(&item.ID, &item.SourceID, &item.SourceName, &item.Title, &item.SourceURL, &item.FetchedAt); err != nil {
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
	return nil
}
