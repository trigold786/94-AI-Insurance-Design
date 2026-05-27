package crawler

import "testing"

func TestRelevanceScoreBasic(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "社保", Weight: 2, Scope: "all"},
		{Keyword: "养老", Weight: 2, Scope: "all"},
		{Keyword: "补贴", Weight: 1, Scope: "all"},
		{Keyword: "职工", Weight: 1, Scope: "all"},
	})
	score, matched := filter.Score("上海社保缴费基数调整通知", "DOUYIN-test", "douyin")
	if score < 2 {
		t.Errorf("expected score >= 2 for 社保+调整, got %d", score)
	}
	if len(matched) == 0 {
		t.Error("expected at least one match")
	}
}

func TestRelevanceScoreIrrelevant(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "社保", Weight: 2, Scope: "all"},
		{Keyword: "养老", Weight: 2, Scope: "all"},
	})
	score, _ := filter.Score("番茄畅听免费看后续温暖医生小说", "DOUYIN-test", "douyin")
	if score != 0 {
		t.Errorf("expected score 0 for irrelevant text, got %d", score)
	}
}

func TestRelevanceScoreScope(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "社保", Weight: 2, Scope: "douyin"},
		{Keyword: "养老", Weight: 2, Scope: "wechat"},
	})
	score1, _ := filter.Score("社保缴费", "SRC", "douyin")
	if score1 != 2 {
		t.Errorf("expected 2 for douyin scope match, got %d", score1)
	}
	score2, _ := filter.Score("社保缴费", "SRC", "wechat")
	if score2 != 0 {
		t.Errorf("expected 0 for wechat scope miss, got %d", score2)
	}
	score3, _ := filter.Score("养老", "SRC", "wechat")
	if score3 != 2 {
		t.Errorf("expected 2 for wechat scope match, got %d", score3)
	}
}

func TestRelevanceExtraKeywords(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "社保", Weight: 2, Scope: "all"},
	})
	filter.SetExtraKeywords("SRC1", []string{"闵行", "浦东"})
	score, _ := filter.Score("闵行区社保中心", "SRC1", "douyin")
	if score < 3 {
		t.Errorf("expected score >= 3 (社保+闵行), got %d", score)
	}
	score2, _ := filter.Score("闵行区社保中心", "SRC2", "douyin")
	if score2 != 2 {
		t.Errorf("expected score 2 (only 社保, no extra for SRC2), got %d", score2)
	}
}

func TestRelevanceThresholds(t *testing.T) {
	filter := NewRelevanceFilter([]Rule{
		{Keyword: "社保", Weight: 2, Scope: "all"},
	})
	filter.SetThresholds("SRC1", 3, 5)
	if filter.MinScore("SRC1", "level1") != 3 {
		t.Error("level1 threshold should be 3")
	}
	if filter.MinScore("SRC1", "level2") != 5 {
		t.Error("level2 threshold should be 5")
	}
	if filter.MinScore("SRC2", "level1") != 1 {
		t.Error("default level1 threshold should be 1")
	}
}
