package diag

import (
	"fmt"
	"strings"

	"github.com/pt-main/tap/go/color"
)

const (
	accent = "[?YW]"
	strong = "[?BBK]"
	bad    = "[?RD]"
	good   = "[?GN]"
	info   = "[?BE]"
	reset  = "[?RT]"
)

// Render formats diagnostics for a terminal, with source excerpts and carets.
//
// Colour codes are resolved once, at the end, so --no-color and redirected
// output need no separate code path.
func Render(list []Diagnostic, text string) string {
	return RenderFrom(list, map[string]string{"": text})
}

// RenderFrom formats diagnostics, resolving each excerpt from the file named
// in its span. A diagnostic that names no file falls back to "".
func RenderFrom(list []Diagnostic, sources map[string]string) string {
	if len(list) == 0 {
		return ""
	}
	parts := make([]string, 0, len(list))
	for _, d := range list {
		parts = append(parts, renderOne(d, sourceFor(d, sources)))
	}
	return color.Set(strings.Join(parts, "\n"))
}

// sourceFor returns the text a diagnostic should be rendered against.
func sourceFor(d Diagnostic, sources map[string]string) string {
	if d.Span != nil && d.Span.File != "" {
		if text, ok := sources[d.Span.File]; ok {
			return text
		}
	}
	return sources[""]
}

func renderOne(d Diagnostic, text string) string {
	var b strings.Builder

	head := d.Message
	if d.Label != "" && !strings.Contains(head, d.Label) {
		head = strings.TrimSpace(d.Label + ": " + head)
	}
	tag := "error"
	if d.Severity == SeverityWarning {
		tag = "warning"
	}
	fmt.Fprintf(&b, "%s%s %s%s\n", bad, tag, reset, head)

	if loc := location(d); loc != "" {
		fmt.Fprintf(&b, "  %s%s%s\n", info, loc, reset)
	}
	if d.Hint != "" {
		fmt.Fprintf(&b, "  %shint:%s %s\n", accent, reset, d.Hint)
	}

	if d.Span != nil && d.Span.Line > 0 {
		if before, marked, after, ok := d.Span.Excerpt(text); ok {
			num := fmt.Sprintf("%4d", d.Span.Line)
			gutter := info + strings.Repeat(" ", len(num)) + reset + " | "
			fmt.Fprintf(&b, "  %s%s%s | %s%s%s%s\n",
				info, num, reset, before, bad, marked, after)
			fmt.Fprintf(&b, "  %s%s%s%s%s\n",
				gutter, strong, d.Span.CaretLine(before, marked), reset, "")
		}
	}

	for _, c := range d.Causes {
		b.WriteString("\n")
		for _, line := range strings.Split(renderOne(c, text), "\n") {
			if line != "" {
				b.WriteString("  " + line + "\n")
			}
		}
	}
	return b.String()
}

func location(d Diagnostic) string {
	parts := []string{}
	if loc := d.Span.String(); loc != "" {
		parts = append(parts, loc)
	}
	if d.Path != "" {
		parts = append(parts, "path "+d.Path)
	}
	if d.Index != nil {
		parts = append(parts, fmt.Sprintf("index %d", *d.Index))
	}
	return strings.Join(parts, " ")
}

// Summary renders a count line such as "2 errors in app.tycl".
func Summary(list []Diagnostic) string {
	if len(list) == 0 {
		return ""
	}
	errs, warns := 0, 0
	files := map[string]bool{}
	for _, d := range list {
		if d.Severity == SeverityWarning {
			warns++
		} else {
			errs++
		}
		if d.Span != nil && d.Span.File != "" {
			files[d.Span.File] = true
		}
	}
	parts := []string{}
	if errs > 0 {
		parts = append(parts, pluralize(errs, "error"))
	}
	if warns > 0 {
		parts = append(parts, pluralize(warns, "warning"))
	}
	out := strings.Join(parts, ", ")
	switch len(files) {
	case 1:
		for name := range files {
			out += " in " + name
		}
	case 0:
	default:
		out += fmt.Sprintf(" in %d files", len(files))
	}
	return out
}

// Success renders a confirmation line.
func Success(text string) string {
	return color.Set(good + text + reset)
}

// Failure renders an error summary line.
func Failure(text string) string {
	return color.Set(bad + text + reset)
}
