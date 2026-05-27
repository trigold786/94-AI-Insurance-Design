package extractor

import (
	"strings"
	"testing"
)

func TestSplitDocument_Short(t *testing.T) {
	text := "短文档内容，不需要分片。"
	chunks := splitDocument(text, 4000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != text {
		t.Fatalf("chunk content mismatch")
	}
}

func TestSplitDocument_Long(t *testing.T) {
	paras := make([]string, 20)
	for i := range paras {
		paras[i] = strings.Repeat("这是第"+string(rune('A'+i))+"段内容。", 100)
	}
	text := strings.Join(paras, "\n\n")
	chunks := splitDocument(text, 4000)
	if len(chunks) < 2 {
		t.Fatalf("expected >= 2 chunks for long doc, got %d", len(chunks))
	}
	for i, c := range chunks {
		if len([]rune(c)) > 4000 {
			t.Fatalf("chunk %d too long: %d chars", i, len([]rune(c)))
		}
	}
	totalLen := 0
	for _, c := range chunks {
		totalLen += len(c)
	}
	if totalLen < len(text)*4/10 {
		t.Fatalf("too much content lost: original=%d chunks_total=%d", len(text), totalLen)
	}
}

func TestSplitDocument_MaxChunks(t *testing.T) {
	paras := make([]string, 30)
	for i := range paras {
		paras[i] = strings.Repeat("段落内容"+string(rune('A'+i)), 200)
	}
	text := strings.Join(paras, "\n\n")
	chunks := splitDocument(text, 2000)
	if len(chunks) > 5 {
		t.Fatalf("expected max 5 chunks, got %d", len(chunks))
	}
}

func TestSplitDocument_SingleHugeParagraph(t *testing.T) {
	text := strings.Repeat("超长段落不分段。", 2000)
	chunks := splitDocument(text, 4000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk for single para, got %d", len(chunks))
	}
}
