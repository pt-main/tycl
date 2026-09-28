package diag

import (
	"fmt"
	"strings"
)

// Source indexes the text of one input so a span can be resolved to
// a human-readable line/column and to a rendered excerpt.
type Source struct {
	Name  string
	Text  string
	Lines []string
}

// NewSource indexes text by lines.
func NewSource(name, text string) *Source {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	return &Source{
		Name:  name,
		Text:  normalized,
		Lines: strings.Split(normalized, "\n"),
	}
}

// locate finds the first occurrence of sub in the file text.
func (s *Source) locate(sub string) int {
	if sub == "" {
		return 0
	}
	idx := strings.Index(s.Text, sub)
	if idx < 0 {
		return -1
	}
	return len([]rune(s.Text[:idx]))
}

// OffsetToPos converts a rune offset into 1-based line and column.
func (s *Source) OffsetToPos(offset int) (line, col int) {
	if s == nil {
		return 0, 0
	}
	runes := []rune(s.Text)
	if offset < 0 {
		offset = 0
	}
	if offset > len(runes) {
		offset = len(runes)
	}
	consumed := 0
	for idx, text := range s.Lines {
		size := len([]rune(text))
		if offset <= consumed+size {
			return idx + 1, offset - consumed + 1
		}
		consumed += size + 1
	}
	return len(s.Lines), 1
}

// OffsetOf converts a 1-based line and column into a rune offset.
func (s *Source) OffsetOf(line, col int) int {
	if s == nil {
		return 0
	}
	if line < 1 {
		line = 1
	}
	if col < 1 {
		col = 1
	}
	offset := 0
	for idx := 1; idx < line && idx <= len(s.Lines); idx++ {
		offset += len([]rune(s.Lines[idx-1])) + 1
	}
	return offset + col - 1
}

// At builds a span for a half-open rune range expressed in this source text.
func (s *Source) At(start, end int) *Span {
	if s == nil {
		return nil
	}
	line, col := s.OffsetToPos(start)
	endLine, endCol := s.OffsetToPos(end)
	span := &Span{File: s.Name, Line: line, Column: col}
	if endLine == line && endCol > col {
		span.EndColumn = endCol
		span.Width = endCol - col
	}
	return span
}

// AtOffset builds a span from an offset already expressed in file coordinates.
func (s *Source) AtOffset(start, end int) *Span {
	if s == nil {
		return nil
	}
	return newFileSpan(s.Name, s.Text, start, end)
}

func newFileSpan(name, text string, start, end int) *Span {
	runes := []rune(text)
	line, col := posOf(runes, start)
	endLine, endCol := posOf(runes, end)
	span := &Span{File: name, Line: line, Column: col}
	if endLine == line && endCol > col {
		span.EndColumn = endCol
		span.Width = endCol - col
	}
	return span
}

func posOf(runes []rune, offset int) (line, col int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(runes) {
		offset = len(runes)
	}
	line = 1
	lineStart := 0
	for i := 0; i < offset; i++ {
		if runes[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return line, offset - lineStart + 1
}

// Span is a 1-based inclusive source range.
type Span struct {
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
	EndColumn int    `json:"endColumn,omitempty"`
	Width     int    `json:"width,omitempty"`
}

// String renders a span as file:line:column.
func (s *Span) String() string {
	if s == nil {
		return ""
	}
	switch {
	case s.File != "" && s.Line > 0:
		return fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.Column)
	case s.File != "":
		return s.File
	case s.Line > 0:
		return fmt.Sprintf("%d:%d", s.Line, s.Column)
	default:
		return ""
	}
}

// Excerpt renders the referenced line with the span underlined.
//
// It returns the line split into three parts so the caller can colour only
// the offending range.
func (s *Span) Excerpt(text string) (before, marked, after string, ok bool) {
	if s == nil || s.Line < 1 {
		return "", "", "", false
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if s.Line > len(lines) {
		return "", "", "", false
	}
	runes := []rune(lines[s.Line-1])

	col := s.Column - 1
	if col < 0 {
		col = 0
	}
	if col > len(runes) {
		col = len(runes)
	}
	width := s.Width
	if width < 1 {
		width = 1
	}
	end := col + width
	if end > len(runes) {
		end = len(runes)
	}
	if end <= col {
		end = min(col+1, len(runes))
	}
	return string(runes[:col]), string(runes[col:end]), string(runes[end:]), true
}

// CaretLine builds the underline for the span, aligned under Excerpt.
func (s *Span) CaretLine(before, marked string) string {
	caret := strings.Repeat(" ", len([]rune(before))) + strings.Repeat("^", len([]rune(marked)))
	if marked == "" {
		caret = strings.Repeat(" ", len([]rune(before))) + "^"
	}
	return caret
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
