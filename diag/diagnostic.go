package diag

import (
	"fmt"
	"sort"
	"strings"

	lcerrors "github.com/pt-main/lc/v2/public/errors"
	"github.com/pt-main/tycl/shared"
)

// Severity classifies how much a diagnostic matters.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Diagnostic is one flat, self-describing problem report.
type Diagnostic struct {
	Code     string       `json:"code"`
	Severity Severity     `json:"severity"`
	Message  string       `json:"message"`
	Hint     string       `json:"hint,omitempty"`
	Span     *Span        `json:"span,omitempty"`
	Path     string       `json:"path,omitempty"`
	Index    *int         `json:"index,omitempty"`
	Label    string       `json:"label,omitempty"`
	Causes   []Diagnostic `json:"causes,omitempty"`
}

// Error is a non-nil error carrying a list of diagnostics.
// It satisfies lc's core.ErrorInterface, so it flows through the same
// plumbing as parser and runtime errors.
type Error struct {
	Diagnostics []Diagnostic
}

func (e *Error) Error() string {
	if e == nil || len(e.Diagnostics) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e.Diagnostics))
	for _, d := range Flatten(e.Diagnostics) {
		parts = append(parts, d.Summary())
	}
	return strings.Join(parts, "; ")
}

// GetCode reports the tycl contract error code.
func (e *Error) GetCode() string { return string(shared.ContractError) }

// GetMsg returns the joined one-line summary.
func (e *Error) GetMsg() string { return e.Error() }

// GetMeta exposes the diagnostics for the generic metadata reader.
func (e *Error) GetMeta() map[lcerrors.ErrorMetaType]interface{} {
	if e == nil {
		return nil
	}
	return map[lcerrors.ErrorMetaType]interface{}{
		"diagnostics": e.Diagnostics,
	}
}

// Unwrap returns no cause: every problem is already a flat diagnostic.
func (e *Error) Unwrap() error { return nil }

// Format renders the diagnostics for a terminal. It is part of the lc
// error interface.
func (e *Error) Format() string { return Render(e.Diagnostics, "") }

// HasErrors reports whether the list contains at least one error.
func HasErrors(list []Diagnostic) bool {
	for _, d := range list {
		if d.Severity != SeverityWarning {
			return true
		}
	}
	return false
}

// Errorf builds an error holding a single diagnostic.
func Errorf(code string, format string, args ...any) *Error {
	return &Error{Diagnostics: []Diagnostic{{
		Code:     code,
		Severity: SeverityError,
		Message:  fmt.Sprintf(format, args...),
	}}}
}

// Summary renders a one-line description without position info.
func (d Diagnostic) Summary() string {
	msg := d.Code
	if d.Message != "" {
		msg = d.Message
	}
	if d.Span != nil {
		if loc := d.Span.String(); loc != "" {
			return fmt.Sprintf("%s: %s", loc, msg)
		}
	}
	return msg
}

// Sort orders diagnostics deterministically: file, then position, then code.
// Repeated runs on the same input always produce byte-identical output.
func Sort(list []Diagnostic) {
	sort.SliceStable(list, func(i, j int) bool {
		a, b := &list[i], &list[j]
		if af, bf := a.Span.fileKey(), b.Span.fileKey(); af != bf {
			return af < bf
		}
		if al, bl := a.Span.lineKey(), b.Span.lineKey(); al != bl {
			return al < bl
		}
		if ac, bc := a.Span.colKey(), b.Span.colKey(); ac != bc {
			return ac < bc
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
}

// Flatten expands nested Causes into a single list, keeping the order in
// which problems were reported.
func Flatten(list []Diagnostic) []Diagnostic {
	out := make([]Diagnostic, 0, len(list))
	var walk func(d Diagnostic)
	walk = func(d Diagnostic) {
		causes := d.Causes
		d.Causes = nil
		out = append(out, d)
		for _, c := range causes {
			walk(c)
		}
	}
	for _, d := range list {
		walk(d)
	}
	return out
}

func (s *Span) fileKey() string {
	if s == nil {
		return ""
	}
	return s.File
}

func (s *Span) lineKey() int {
	if s == nil {
		return 1 << 30
	}
	return s.Line
}

func (s *Span) colKey() int {
	if s == nil {
		return 1 << 30
	}
	return s.Column
}

// pluralize renders "1 problem" / "3 problems".
func pluralize(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// ordinal renders 1st/2nd/3rd/4th.
func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}

// dedupe removes diagnostics that are identical after sorting, so a single
// mistake is never reported twice.
func dedupe(list []Diagnostic) []Diagnostic {
	if len(list) < 2 {
		return list
	}
	out := list[:1]
	for _, d := range list[1:] {
		last := out[len(out)-1]
		if last.Code == d.Code && last.Message == d.Message &&
			last.Span.fileKey() == d.Span.fileKey() &&
			last.Span.lineKey() == d.Span.lineKey() &&
			last.Span.colKey() == d.Span.colKey() &&
			last.Path == d.Path {
			continue
		}
		out = append(out, d)
	}
	return out
}

// joinPath builds a dotted config path, skipping empty segments.
func joinPath(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, ".")
}
