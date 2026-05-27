package crawler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractPDFText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("pdf open: %w", err)
	}
	var buf strings.Builder
	n := reader.NumPage()
	for i := 1; i <= n; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	result := strings.TrimSpace(buf.String())
	if len(result) == 0 {
		return "", fmt.Errorf("pdf: no text extracted from %d pages", n)
	}
	return result, nil
}

func ExtractDOCXText(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("docx open: %w", err)
	}
	var buf strings.Builder
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("docx read document.xml: %w", err)
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("docx read content: %w", err)
			}
			text := stripXMLTags(string(content))
			text = strings.TrimSpace(text)
			if len(text) > 0 {
				buf.WriteString(text)
			}
		}
	}
	result := buf.String()
	if len(result) == 0 {
		return "", fmt.Errorf("docx: no text extracted")
	}
	return result, nil
}

func stripXMLTags(s string) string {
	var buf strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			buf.WriteString(" ")
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}
	result := buf.String()
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	for strings.Contains(result, "\n ") {
		result = strings.ReplaceAll(result, "\n ", "\n")
	}
	result = strings.ReplaceAll(result, "\t", " ")
	return result
}

func IsPDFContentType(ct string) bool {
	return strings.Contains(ct, "application/pdf")
}

func IsDOCXContentType(ct string) bool {
	return strings.Contains(ct, "officedocument.wordprocessingml") ||
		strings.Contains(ct, "msword") ||
		strings.Contains(ct, "application/doc")
}

func IsDocumentContentType(ct string) bool {
	return IsPDFContentType(ct) || IsDOCXContentType(ct)
}
