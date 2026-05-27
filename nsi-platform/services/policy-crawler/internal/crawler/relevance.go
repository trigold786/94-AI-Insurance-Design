package crawler

import (
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"
)

type Rule struct {
	Keyword string
	Weight  int
	Scope   string
}

type RelevanceFilter struct {
	mu       sync.RWMutex
	rules    []Rule
	extra    map[string][]Rule
	l1Min    map[string]int
	l2Min    map[string]int
	db       *sql.DB
	lastLoad time.Time
}

func NewRelevanceFilter(rules []Rule) *RelevanceFilter {
	return &RelevanceFilter{
		rules: rules,
		extra: make(map[string][]Rule),
		l1Min: make(map[string]int),
		l2Min: make(map[string]int),
	}
}

func (f *RelevanceFilter) SetExtraKeywords(sourceID string, keywords []string) {
	var rules []Rule
	for _, kw := range keywords {
		rules = append(rules, Rule{Keyword: kw, Weight: 1, Scope: "all"})
	}
	f.mu.Lock()
	f.extra[sourceID] = rules
	f.mu.Unlock()
}

func (f *RelevanceFilter) SetThresholds(sourceID string, l1, l2 int) {
	f.mu.Lock()
	f.l1Min[sourceID] = l1
	f.l2Min[sourceID] = l2
	f.mu.Unlock()
}

func (f *RelevanceFilter) MinScore(sourceID, level string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	m := f.l1Min
	if level == "level2" {
		m = f.l2Min
	}
	if v, ok := m[sourceID]; ok {
		return v
	}
	if level == "level1" {
		return 1
	}
	return 2
}

func (f *RelevanceFilter) Score(text, sourceID, crawlType string) (int, []string) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	lower := strings.ToLower(text)
	var total int
	var matched []string
	allRules := f.collectRules(sourceID, crawlType)
	for _, r := range allRules {
		if strings.Contains(lower, strings.ToLower(r.Keyword)) {
			total += r.Weight
			matched = append(matched, r.Keyword)
		}
	}
	return total, matched
}

func (f *RelevanceFilter) collectRules(sourceID, crawlType string) []Rule {
	var result []Rule
	for _, r := range f.rules {
		if r.Scope == "all" || r.Scope == crawlType {
			result = append(result, r)
		}
	}
	if extra, ok := f.extra[sourceID]; ok {
		result = append(result, extra...)
	}
	return result
}

func (f *RelevanceFilter) LoadFromDB(db *sql.DB) error {
	f.db = db
	return f.Reload()
}

func (f *RelevanceFilter) Reload() error {
	if f.db == nil {
		return nil
	}
	rows, err := f.db.Query(`SELECT keyword, weight, scope FROM relevance_rules WHERE enabled`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var rules []Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.Keyword, &r.Weight, &r.Scope); err != nil {
			return err
		}
		rules = append(rules, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	tRows, err := f.db.Query(`SELECT source_id, level1_min_score, level2_min_score, extra_keywords FROM relevance_thresholds`)
	if err != nil {
		return err
	}
	defer tRows.Close()
	extra := make(map[string][]Rule)
	l1Min := make(map[string]int)
	l2Min := make(map[string]int)
	for tRows.Next() {
		var sid string
		var l1, l2 int
		var ek string
		if err := tRows.Scan(&sid, &l1, &l2, &ek); err != nil {
			return err
		}
		l1Min[sid] = l1
		l2Min[sid] = l2
		if ek != "" {
			for _, kw := range strings.Split(ek, ",") {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					extra[sid] = append(extra[sid], Rule{Keyword: kw, Weight: 1, Scope: "all"})
				}
			}
		}
	}

	f.mu.Lock()
	f.rules = rules
	f.extra = extra
	f.l1Min = l1Min
	f.l2Min = l2Min
	f.lastLoad = time.Now()
	f.mu.Unlock()
	log.Printf("[relevance] loaded %d rules, %d source overrides", len(rules), len(l1Min))
	return nil
}

func (f *RelevanceFilter) StartReloadLoop(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := f.Reload(); err != nil {
				log.Printf("[relevance] reload error: %v", err)
			}
		case <-stopCh:
			return
		}
	}
}
