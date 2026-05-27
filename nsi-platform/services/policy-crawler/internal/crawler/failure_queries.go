package crawler

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/admin"
)

func (s *DBStore) GetFailureSummary() (*admin.FailureSummary, error) {
	summary := &admin.FailureSummary{}
	s.db.QueryRow(`SELECT COUNT(*) FROM crawl_logs WHERE status = 'failed'`).Scan(&summary.CrawlFailures)
	s.db.QueryRow(`SELECT COUNT(*) FROM extract_logs WHERE status = 'failed'`).Scan(&summary.ExtractFailures)
	s.db.QueryRow(`SELECT COUNT(*) FROM policy_raw_texts WHERE video_extract_status = 'failed'`).Scan(&summary.VideoFailures)
	return summary, nil
}

func (s *DBStore) GetFailureTrend(days int) ([]admin.FailureTrendPoint, error) {
	rows, err := s.db.Query(`
		WITH date_series AS (
			SELECT generate_series(
				(CURRENT_DATE - ($1::int || ' days')::interval)::date,
				CURRENT_DATE,
				'1 day'::interval
			)::date AS d
		)
		SELECT
			ds.d::text AS date,
			COALESCE(cf.cnt, 0) AS crawl_failures,
			COALESCE(ef.cnt, 0) AS extract_failures,
			COALESCE(vf.cnt, 0) AS video_failures
		FROM date_series ds
		LEFT JOIN (
			SELECT crawled_at::date AS d, COUNT(*) AS cnt
			FROM crawl_logs
			WHERE status = 'failed' AND crawled_at >= CURRENT_DATE - ($1::int || ' days')::interval
			GROUP BY d
		) cf ON cf.d = ds.d
		LEFT JOIN (
			SELECT created_at::date AS d, COUNT(*) AS cnt
			FROM extract_logs
			WHERE status = 'failed' AND created_at >= CURRENT_DATE - ($1::int || ' days')::interval
			GROUP BY d
		) ef ON ef.d = ds.d
		LEFT JOIN (
			SELECT fetch_time::date AS d, COUNT(*) AS cnt
			FROM policy_raw_texts
			WHERE video_extract_status = 'failed' AND fetch_time >= CURRENT_DATE - ($1::int || ' days')::interval
			GROUP BY d
		) vf ON vf.d = ds.d
		ORDER BY ds.d`, days)
	if err != nil {
		return nil, fmt.Errorf("failure trend query: %w", err)
	}
	defer rows.Close()

	var points []admin.FailureTrendPoint
	for rows.Next() {
		var p admin.FailureTrendPoint
		if err := rows.Scan(&p.Date, &p.CrawlFailures, &p.ExtractFailures, &p.VideoFailures); err != nil {
			return nil, fmt.Errorf("failure trend scan: %w", err)
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *DBStore) GetFailureBySource() ([]admin.FailureBySourceEntry, error) {
	rows, err := s.db.Query(`
		SELECT
			combined.source_id,
			COALESCE(ps.source_name, combined.source_id),
			COALESCE(cf.cnt, 0) AS crawl_failures,
			COALESCE(ef.cnt, 0) AS extract_failures,
			COALESCE(vf.cnt, 0) AS video_failures
		FROM (
			SELECT source_id FROM crawl_logs WHERE status = 'failed'
			UNION
			SELECT source_id FROM extract_logs WHERE status = 'failed'
			UNION
			SELECT source_id FROM policy_raw_texts WHERE video_extract_status = 'failed'
		) combined
		LEFT JOIN policy_sources ps ON ps.source_id = combined.source_id
		LEFT JOIN (
			SELECT source_id, COUNT(*) AS cnt FROM crawl_logs WHERE status = 'failed' GROUP BY source_id
		) cf ON cf.source_id = combined.source_id
		LEFT JOIN (
			SELECT source_id, COUNT(*) AS cnt FROM extract_logs WHERE status = 'failed' GROUP BY source_id
		) ef ON ef.source_id = combined.source_id
		LEFT JOIN (
			SELECT source_id, COUNT(*) AS cnt FROM policy_raw_texts WHERE video_extract_status = 'failed' GROUP BY source_id
		) vf ON vf.source_id = combined.source_id
		ORDER BY (COALESCE(cf.cnt,0) + COALESCE(ef.cnt,0) + COALESCE(vf.cnt,0)) DESC`)
	if err != nil {
		return nil, fmt.Errorf("failure by source query: %w", err)
	}
	defer rows.Close()

	var entries []admin.FailureBySourceEntry
	for rows.Next() {
		var e admin.FailureBySourceEntry
		if err := rows.Scan(&e.SourceID, &e.SourceName, &e.CrawlFailures, &e.ExtractFailures, &e.VideoFailures); err != nil {
			return nil, fmt.Errorf("failure by source scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *DBStore) GetTopFailureReasons(limit int) ([]admin.TopFailureReason, error) {
	rows, err := s.db.Query(`
		SELECT reason, SUM(cnt) AS total FROM (
			SELECT COALESCE(NULLIF(error_message,''), 'unknown error') AS reason, COUNT(*) AS cnt
			FROM crawl_logs WHERE status = 'failed' GROUP BY error_message
			UNION ALL
			SELECT COALESCE(NULLIF(message,''), 'unknown error') AS reason, COUNT(*) AS cnt
			FROM extract_logs WHERE status = 'failed' GROUP BY message
		) combined
		GROUP BY reason
		ORDER BY total DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("top failure reasons query: %w", err)
	}
	defer rows.Close()

	var reasons []admin.TopFailureReason
	for rows.Next() {
		var r admin.TopFailureReason
		if err := rows.Scan(&r.Reason, &r.Count); err != nil {
			return nil, fmt.Errorf("top failure reasons scan: %w", err)
		}
		reasons = append(reasons, r)
	}
	return reasons, rows.Err()
}

func (s *DBStore) GetFailedRawTexts(sourceID string, failureType string, limit int) ([]admin.FailedRawTextEntry, error) {
	var queries []string
	var args []interface{}
	argIdx := 1

	if failureType == "" || failureType == "extract" {
		sIdx := argIdx
		lIdx1 := argIdx + 1
		argIdx += 2
		args = append(args, sourceID, limit)
		queries = append(queries, fmt.Sprintf(`
			SELECT el.raw_text_id::bigint,
				el.source_id,
				COALESCE(ps.source_name, el.source_id),
				COALESCE(el.title, ''),
				COALESCE(el.message, ''),
				el.created_at::text,
				'extract'
			FROM extract_logs el
			LEFT JOIN policy_sources ps ON ps.source_id = el.source_id
			WHERE el.status = 'failed'
				AND el.raw_text_id IS NOT NULL
				AND ($%d = '' OR el.source_id = $%d)
			ORDER BY el.created_at DESC
			LIMIT $%d`, sIdx, sIdx, lIdx1))
	}

	if failureType == "" || failureType == "video" {
		sIdx := argIdx
		lIdx2 := argIdx + 1
		argIdx += 2
		args = append(args, sourceID, limit)
		queries = append(queries, fmt.Sprintf(`
			SELECT prt.id,
				prt.source_id,
				COALESCE(ps.source_name, prt.source_id),
				COALESCE(prt.title, ''),
				'',
				COALESCE(prt.fetch_time::text, ''),
				'video'
			FROM policy_raw_texts prt
			LEFT JOIN policy_sources ps ON ps.source_id = prt.source_id
			WHERE prt.video_extract_status = 'failed'
				AND ($%d = '' OR prt.source_id = $%d)
			ORDER BY prt.fetch_time DESC
			LIMIT $%d`, sIdx, sIdx, lIdx2))
	}

	if len(queries) == 0 {
		return nil, nil
	}

	query := strings.Join(queries, " UNION ALL ") + " ORDER BY failed_at DESC"
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed raw texts query: %w", err)
	}
	defer rows.Close()

	var entries []admin.FailedRawTextEntry
	for rows.Next() {
		var e admin.FailedRawTextEntry
		if err := rows.Scan(&e.ID, &e.SourceID, &e.SourceName, &e.Title, &e.ErrorReason, &e.FailedAt, &e.FailureType); err != nil {
			return nil, fmt.Errorf("failed raw texts scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *DBStore) RetryRawText(id int64) error {
	var videoStatus string
	s.db.QueryRow(`SELECT COALESCE(video_extract_status,'') FROM policy_raw_texts WHERE id = $1`, id).Scan(&videoStatus)
	if videoStatus == "failed" {
		_, err := s.db.Exec(`UPDATE policy_raw_texts SET video_extract_status = 'pending' WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("retry video for %d: %w", id, err)
		}
	} else {
		_, err := s.db.Exec(`UPDATE policy_raw_texts SET extracted = false WHERE id = $1`, id)
		if err != nil {
			return fmt.Errorf("retry extract for %d: %w", id, err)
		}
	}
	return nil
}

func (s *DBStore) RetryAllFailed(sourceID string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	videoResult, err := tx.Exec(
		`UPDATE policy_raw_texts SET video_extract_status = 'pending' WHERE source_id = $1 AND video_extract_status = 'failed'`,
		sourceID)
	if err != nil {
		return 0, fmt.Errorf("retry video failed: %w", err)
	}
	videoCount, _ := videoResult.RowsAffected()

	rows, err := tx.Query(
		`SELECT DISTINCT raw_text_id FROM extract_logs WHERE source_id = $1 AND status = 'failed' AND raw_text_id IS NOT NULL`,
		sourceID)
	if err != nil {
		return int(videoCount), nil
	}

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()

	var extractCount int64
	if len(ids) > 0 {
		extractResult, err := tx.Exec(
			`UPDATE policy_raw_texts SET extracted = false WHERE id = ANY($1)`,
			pq.Array(ids))
		if err != nil {
			return int(videoCount), fmt.Errorf("retry extract failed: %w", err)
		}
		extractCount, _ = extractResult.RowsAffected()
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return int(videoCount + extractCount), nil
}
