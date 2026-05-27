package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RelevanceScorer interface {
	Score(text, sourceID, crawlType string) (int, []string)
}

func RelevanceRulesListHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		category := r.URL.Query().Get("category")
		scope := r.URL.Query().Get("scope")

		query := "SELECT id, category, keyword, weight, scope, enabled FROM relevance_rules WHERE enabled"
		var args []interface{}
		if category != "" {
			query += " AND category = ?"
			args = append(args, category)
		}
		if scope != "" {
			query += " AND scope = ?"
			args = append(args, scope)
		}
		query += " ORDER BY id"

		rows, err := db.Query(query, args...)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("query error: %v", err))
			return
		}
		defer rows.Close()

		var rules []map[string]interface{}
		for rows.Next() {
			var id int64
			var category, keyword, scope string
			var weight int
			var enabled bool
			if err := rows.Scan(&id, &category, &keyword, &weight, &scope, &enabled); err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("scan error: %v", err))
				return
			}
			rules = append(rules, map[string]interface{}{
				"id":       id,
				"category": category,
				"keyword":  keyword,
				"weight":   weight,
				"scope":    scope,
				"enabled":  enabled,
			})
		}
		if rules == nil {
			rules = []map[string]interface{}{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": rules})
	})
}

func RelevanceRulesCreateHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Category string `json:"category"`
			Keyword  string `json:"keyword"`
			Weight   int    `json:"weight"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Keyword == "" {
			respondError(w, http.StatusBadRequest, "keyword required")
			return
		}
		if req.Scope == "" {
			req.Scope = "all"
		}
		if req.Weight == 0 {
			req.Weight = 1
		}

		result, err := db.Exec(
			"INSERT INTO relevance_rules (category, keyword, weight, scope, enabled) VALUES (?, ?, ?, ?, true)",
			req.Category, req.Keyword, req.Weight, req.Scope,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert error: %v", err))
			return
		}
		id, _ := result.LastInsertId()

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"id":       id,
				"category": req.Category,
				"keyword":  req.Keyword,
				"weight":   req.Weight,
				"scope":    req.Scope,
				"enabled":  true,
			},
		})
	})
}

func RelevanceRulesUpdateHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			ID      int64  `json:"id"`
			Weight  *int   `json:"weight"`
			Enabled *bool  `json:"enabled"`
			Scope   string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.ID == 0 {
			respondError(w, http.StatusBadRequest, "id required")
			return
		}

		parts := []string{}
		args := []interface{}{}
		if req.Weight != nil {
			parts = append(parts, "weight = ?")
			args = append(args, *req.Weight)
		}
		if req.Enabled != nil {
			parts = append(parts, "enabled = ?")
			args = append(args, *req.Enabled)
		}
		if req.Scope != "" {
			parts = append(parts, "scope = ?")
			args = append(args, req.Scope)
		}
		if len(parts) == 0 {
			respondError(w, http.StatusBadRequest, "no fields to update")
			return
		}
		args = append(args, req.ID)
		query := "UPDATE relevance_rules SET " + strings.Join(parts, ", ") + " WHERE id = ?"

		res, err := db.Exec(query, args...)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("update error: %v", err))
			return
		}
		rows, _ := res.RowsAffected()

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": fmt.Sprintf("updated %d row(s)", rows),
		})
	})
}

func RelevanceRulesDeleteHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.ID == 0 {
			respondError(w, http.StatusBadRequest, "id required")
			return
		}

		res, err := db.Exec("DELETE FROM relevance_rules WHERE id = ?", req.ID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("delete error: %v", err))
			return
		}
		rows, _ := res.RowsAffected()

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": fmt.Sprintf("deleted %d row(s)", rows),
		})
	})
}

func RelevanceThresholdHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := "/admin/relevance/thresholds/"
		sourceID := strings.TrimPrefix(r.URL.Path, prefix)
		if sourceID == "" {
			respondError(w, http.StatusBadRequest, "source_id required in path")
			return
		}

		switch r.Method {
		case http.MethodGet:
			var l1, l2 int
			var ek string
			err := db.QueryRow(
				"SELECT level1_min_score, level2_min_score, extra_keywords FROM relevance_thresholds WHERE source_id = ?",
				sourceID,
			).Scan(&l1, &l2, &ek)
			if err == sql.ErrNoRows {
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"code": 0,
					"data": map[string]interface{}{
						"source_id":        sourceID,
						"level1_min_score": 0,
						"level2_min_score": 0,
						"extra_keywords":   "",
					},
				})
				return
			}
			if err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("query error: %v", err))
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{
					"source_id":        sourceID,
					"level1_min_score": l1,
					"level2_min_score": l2,
					"extra_keywords":   ek,
				},
			})

		case http.MethodPut:
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			var req struct {
				Level1MinScore int    `json:"level1_min_score"`
				Level2MinScore int    `json:"level2_min_score"`
				ExtraKeywords  string `json:"extra_keywords"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				respondError(w, http.StatusBadRequest, "invalid JSON")
				return
			}

			_, err := db.Exec(
				`INSERT INTO relevance_thresholds (source_id, level1_min_score, level2_min_score, extra_keywords)
				 VALUES (?, ?, ?, ?)
				 ON CONFLICT(source_id) DO UPDATE SET level1_min_score=?, level2_min_score=?, extra_keywords=?`,
				sourceID, req.Level1MinScore, req.Level2MinScore, req.ExtraKeywords,
				req.Level1MinScore, req.Level2MinScore, req.ExtraKeywords,
			)
			if err != nil {
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("upsert error: %v", err))
				return
			}
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"code":    0,
				"message": "threshold saved",
			})

		default:
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}

func RelevanceTestHandler(filter RelevanceScorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req struct {
			Text      string `json:"text"`
			SourceID  string `json:"source_id"`
			CrawlType string `json:"crawl_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if req.Text == "" {
			respondError(w, http.StatusBadRequest, "text required")
			return
		}

		score, matched := filter.Score(req.Text, req.SourceID, req.CrawlType)
		if matched == nil {
			matched = []string{}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"score":   score,
				"matched": matched,
			},
		})
	})
}

func RelevanceBulkImportHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		var req []struct {
			Category string `json:"category"`
			Keyword  string `json:"keyword"`
			Weight   int    `json:"weight"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if len(req) == 0 {
			respondError(w, http.StatusBadRequest, "empty array")
			return
		}

		tx, err := db.Begin()
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("begin tx: %v", err))
			return
		}

		stmt, err := tx.Prepare("INSERT INTO relevance_rules (category, keyword, weight, scope, enabled) VALUES (?, ?, ?, ?, true)")
		if err != nil {
			tx.Rollback()
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("prepare: %v", err))
			return
		}
		defer stmt.Close()

		var imported int
		for _, item := range req {
			if item.Keyword == "" {
				continue
			}
			if item.Scope == "" {
				item.Scope = "all"
			}
			if item.Weight == 0 {
				item.Weight = 1
			}
			if _, err := stmt.Exec(item.Category, item.Keyword, item.Weight, item.Scope); err != nil {
				tx.Rollback()
				respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert error at keyword %q: %v", item.Keyword, err))
				return
			}
			imported++
		}

		if err := tx.Commit(); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("commit: %v", err))
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"code":    0,
			"message": fmt.Sprintf("imported %d rules", imported),
		})
	})
}
