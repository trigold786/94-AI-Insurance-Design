package crawler

import (
	"testing"
	"time"
)

const rssSample = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>Test Feed</title>
<item>
<title>各地社保补贴政策汇总</title>
<link>https://example.com/policy1</link>
<description>2025年各地社保补贴政策最新汇总</description>
<content:encoded xmlns:content="http://purl.org/rss/1.0/modules/content/">&lt;p&gt;详细内容&lt;/p&gt;</content:encoded>
</item>
<item>
<title>灵活就业人员养老保险新规</title>
<link>https://example.com/policy2</link>
<description>灵活就业人员参加养老保险的最新规定</description>
</item>
</channel>
</rss>`

const atomSample = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Test Atom Feed</title>
<entry>
<title>养老保险全国统筹实施方案</title>
<link href="https://example.com/atom1" rel="alternate"/>
<summary>养老保险全国统筹最新实施方案</summary>
<content type="html">&lt;p&gt;详细实施内容&lt;/p&gt;</content>
</entry>
<entry>
<title>医保异地结算新政策</title>
<link href="https://example.com/atom2"/>
<summary>跨省异地就医直接结算新政策</summary>
</entry>
</feed>`

func TestParseRSS(t *testing.T) {
	items, err := ParseFeed([]byte(rssSample))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "各地社保补贴政策汇总" {
		t.Fatalf("unexpected title: %s", items[0].Title)
	}
	if items[0].Link != "https://example.com/policy1" {
		t.Fatalf("unexpected link: %s", items[0].Link)
	}
	if items[0].Content != "<p>详细内容</p>" {
		t.Fatalf("unexpected content: %s", items[0].Content)
	}
	if items[1].Description != "灵活就业人员参加养老保险的最新规定" {
		t.Fatalf("unexpected desc: %s", items[1].Description)
	}
}

func TestParseAtom(t *testing.T) {
	items, err := ParseFeed([]byte(atomSample))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Title != "养老保险全国统筹实施方案" {
		t.Fatalf("unexpected title: %s", items[0].Title)
	}
	if items[0].Link != "https://example.com/atom1" {
		t.Fatalf("unexpected link: %s", items[0].Link)
	}
	if items[0].Content != "<p>详细实施内容</p>" {
		t.Fatalf("unexpected content: %s", items[0].Content)
	}
	if items[1].Link != "https://example.com/atom2" {
		t.Fatalf("unexpected link: %s", items[1].Link)
	}
}

func TestParseEmptyFeed(t *testing.T) {
	items, err := ParseFeed([]byte(`<rss version="2.0"><channel></channel></rss>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestRSSCrawler_SourceID(t *testing.T) {
	cfg := SourceConfig{SourceID: "RSS-TEST", SourceLevel: "MEDIUM"}
	r := NewRSSCrawler(cfg)
	if r.SourceID() != "RSS-TEST" {
		t.Fatalf("expected RSS-TEST, got %s", r.SourceID())
	}
	if r.SourceLevel() != "MEDIUM" {
		t.Fatalf("expected MEDIUM, got %s", r.SourceLevel())
	}
}

func TestRSSCrawler_Interval(t *testing.T) {
	cfg := SourceConfig{SourceLevel: "MEDIUM", IntervalSec: 604800}
	r := NewRSSCrawler(cfg)
	if r.Interval() != 604800*time.Second {
		t.Fatalf("expected 168h, got %v", r.Interval())
	}
}

func TestRSSCrawler_DefaultInterval(t *testing.T) {
	cfg := SourceConfig{SourceLevel: "MEDIUM"}
	r := NewRSSCrawler(cfg)
	if r.Interval() != 168*time.Hour {
		t.Fatalf("expected 168h default, got %v", r.Interval())
	}
}
