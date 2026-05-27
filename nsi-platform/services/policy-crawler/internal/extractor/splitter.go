package extractor

import (
	"strings"
)

func splitDocument(text string, maxChunkSize int) []string {
	runes := []rune(text)
	if len(runes) <= maxChunkSize {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var current strings.Builder

	for _, para := range paragraphs {
		paraRunes := []rune(para)
		if len(paraRunes) > maxChunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			chunks = append(chunks, para)
			continue
		}
		if current.Len() > 0 && current.Len()+len(paraRunes)+2 > maxChunkSize {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteString("\n\n")
		}
		current.WriteString(para)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	if len(chunks) > 5 {
		chunks = chunks[:5]
	}

	return chunks
}
