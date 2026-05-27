package crawler

import (
	"github.com/trigold786/94-AI-Insurance-Design/policy-crawler/internal/admin"
)

func (s *DBStore) GetFailureAnalysis() (*admin.FailureAnalysis, error) {
	result := &admin.FailureAnalysis{
		CrawlFailures:   []admin.FailureGroup{},
		ExtractFailures: []admin.FailureGroup{},
	}

	crawlRows, err := s.db.Query(`
		SELECT cl.source_id,
		       COALESCE(MAX(ps.source_name), cl.source_id),
		       COALESCE(NULLIF(cl.error_message,''), 'unknown error') AS reason,
		       COUNT(*) AS cnt
		FROM crawl_logs cl
		LEFT JOIN policy_sources ps ON ps.source_id = cl.source_id
		WHERE cl.status != 'success'
		GROUP BY cl.source_id, reason
		ORDER BY cnt DESC, cl.source_id
		LIMIT 30`)
	if err == nil {
		defer crawlRows.Close()
		for crawlRows.Next() {
			var g admin.FailureGroup
			if err := crawlRows.Scan(&g.SourceID, &g.SourceName, &g.ErrorMessage, &g.Count); err == nil {
				result.CrawlFailures = append(result.CrawlFailures, g)
			}
		}
	}

	extractRows, err := s.db.Query(`
		SELECT el.source_id,
		       COALESCE(MAX(ps.source_name), el.source_id),
		       COALESCE(NULLIF(el.message,''), 'unknown error') AS reason,
		       COUNT(*) AS cnt
		FROM extract_logs el
		LEFT JOIN policy_sources ps ON ps.source_id = el.source_id
		WHERE el.status != 'success'
		GROUP BY el.source_id, reason
		ORDER BY cnt DESC, el.source_id
		LIMIT 30`)
	if err == nil {
		defer extractRows.Close()
		for extractRows.Next() {
			var g admin.FailureGroup
			if err := extractRows.Scan(&g.SourceID, &g.SourceName, &g.ErrorMessage, &g.Count); err == nil {
				result.ExtractFailures = append(result.ExtractFailures, g)
			}
		}
	}

	return result, nil
}
