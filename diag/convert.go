package diag

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/parsing/stringParsing/parser3"
	lcerrors "github.com/pt-main/lc/v2/public/errors"
	"github.com/pt-main/tycl/shared"
)

// Stable diagnostic codes reported by the JSON output.
const (
	CodeSyntax   = "syntax.parse"
	CodeLex      = "syntax.lex"
	CodeType     = "type.invalid"
	CodeContract = "contract.violation"
	CodeValue    = "value.invalid"
	CodeIO       = "io.error"
	CodeUsage    = "usage.error"
	CodeInternal = "internal.error"
	CodeGrammar  = "syntax.grammar"
)

// FromError flattens an lc/tycl error chain into flat diagnostics.
//
// Typed lc parse errors carry token offsets, so they resolve to exact
// file:line:column. tycl error codes are mapped to stable machine-readable
// codes. Nested context becomes structured Causes instead of one long string.
func FromError(src *Source, err error) []Diagnostic {
	if err == nil {
		return nil
	}
	out := flatten(src, err, nil, 0)
	Sort(out)
	return dedupe(out)
}

const maxDepth = 6

// pathCtx carries the config path and array index that are known at the
// point where an error was produced.
type pathCtx struct {
	path  string
	index *int
}

func flatten(src *Source, err error, ctx *pathCtx, depth int) []Diagnostic {
	if err == nil || depth > maxDepth {
		return nil
	}

	var pe *parser3.ParseError
	if asError(err, &pe) {
		return applyCtx(flattenParseError(src, pe), ctx)
	}

	var ge *parser3.GrammarError
	if asError(err, &ge) {
		d := Diagnostic{Code: CodeGrammar, Severity: SeverityError, Message: ge.Msg, Label: ge.Phase}
		out := []Diagnostic{d}
		if ge.Cause != nil {
			out = append(out, flatten(src, ge.Cause, ctx, depth+1)...)
		}
		return out
	}

	var ae *parser3.AdapterError
	if asError(err, &ae) {
		d := Diagnostic{Code: CodeInternal, Severity: SeverityError, Message: ae.Msg}
		if ae.Cause != nil {
			d.Hint = ae.Cause.Error()
		}
		return []Diagnostic{d}
	}

	if ei, ok := err.(core.ErrorInterface); ok {
		return applyCtx(flattenCore(src, ei, ctx, depth), ctx)
	}

	return applyCtx([]Diagnostic{{
		Code:     CodeInternal,
		Severity: SeverityError,
		Message:  err.Error(),
	}}, ctx)
}

func applyCtx(list []Diagnostic, ctx *pathCtx) []Diagnostic {
	if ctx == nil {
		return list
	}
	for i := range list {
		if list[i].Path == "" {
			list[i].Path = ctx.path
		}
		if list[i].Index == nil {
			list[i].Index = ctx.index
		}
	}
	return list
}

func flattenParseError(src *Source, pe *parser3.ParseError) []Diagnostic {
	head := Diagnostic{
		Code:     CodeSyntax,
		Severity: SeverityError,
		Message:  parseMessage(pe),
		Label:    pe.Phase,
		Span:     parseSpan(src, pe),
		Hint:     parseHint(pe),
	}

	// A parse failure is a chain of wrappers around one real mistake. The
	// innermost diagnostic knows the exact token, so it is reported alone.
	if causes := flatten(src, pe.Cause, nil, maxDepth); len(causes) > 0 {
		root := deepest(causes)
		if root.Message != "" && root.Message != "invalid syntax" {
			return []Diagnostic{root}
		}
		if root.Hint != "" {
			head.Hint = root.Hint
		}
		if root.Span != nil {
			head.Span = root.Span
		}
	}
	return []Diagnostic{head}
}

func deepest(list []Diagnostic) Diagnostic {
	cur := list[0]
	for len(cur.Causes) > 0 {
		cur = cur.Causes[0]
	}
	return cur
}

func parseSpan(src *Source, pe *parser3.ParseError) *Span {
	if src == nil {
		return nil
	}
	// The token index is authoritative: the raw text is searched only as a
	// fallback, because the same token text can appear many times in a file.
	if start, ok := tokenOffset(pe.TokenPos); ok {
		span := src.At(start, start+maxWidth(pe.Raw))
		span.Width = maxWidth(pe.Raw)
		span.EndColumn = span.Column + span.Width
		return span
	}
	if pe.Raw == "" {
		return nil
	}
	offset := src.locate(pe.Raw)
	if offset < 0 {
		return nil
	}
	span := src.At(offset, offset+len([]rune(pe.Raw)))
	// The lexer counts runes, not bytes, so resolve through the same unit.
	if line, col, ok := posFromTokenPos(src, pe.TokenPos); ok {
		span.Line = line
		span.Column = col
		span.EndColumn = col + len([]rune(pe.Raw))
		span.Width = len([]rune(pe.Raw))
	}
	return span
}

func maxWidth(raw string) int {
	if raw == "" {
		return 1
	}
	return len([]rune(raw))
}

func posFromTokenPos(src *Source, tokenPos string) (line, col int, ok bool) {
	start, found := tokenOffset(tokenPos)
	if !found {
		return 0, 0, false
	}
	line, col = src.OffsetToPos(start)
	return line, col, true
}

func tokenOffset(tokenPos string) (int, bool) {
	idx := strings.Index(tokenPos, "start=")
	if idx < 0 {
		return 0, false
	}
	rest := tokenPos[idx+len("start="):]
	end := strings.IndexAny(rest, "- ")
	if end < 0 {
		end = len(rest)
	}
	value, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0, false
	}
	return value, true
}

func parseMessage(pe *parser3.ParseError) string {
	if pe.Msg != "" && pe.Expected == "" {
		return pe.Msg
	}
	switch {
	case pe.Expected != "" && pe.Got != "":
		return fmt.Sprintf("expected %s, got %s", quote(pe.Expected), quote(pe.Got))
	case pe.Expected != "":
		return fmt.Sprintf("expected %s", quote(pe.Expected))
	case pe.Got != "":
		return fmt.Sprintf("unexpected %s", quote(pe.Got))
	case pe.Msg != "":
		return pe.Msg
	default:
		return "invalid syntax"
	}
}

func parseHint(pe *parser3.ParseError) string {
	if pe.Phase != parser3.PhaseExpect || pe.Expected == "" {
		return ""
	}
	// A file-level comment before the opening brace is the most common
	// mistake: the language only accepts comments inside an object.
	if pe.Expected == "LBRACE" && pe.Got == "COMMENT" {
		return "comments belong inside the object: move this one after '{' or before the last '}'"
	}
	switch pe.Expected {
	case "RBRACE":
		return "a value is missing before '}', or the previous pair needs a trailing ','"
	case "RBRACK":
		return "a value is missing before ']', or the previous item needs a trailing ','"
	case "SEPARATOR":
		return "pairs and array items must be separated by ','"
	case "IDENT":
		return "keys and type names must start with a letter or '_'"
	case "ASSIGN":
		return "every key needs a value: 'key: type = value' or 'key = value'"
	case "COLON":
		return "type assertions use ':' before the type, e.g. 'port: int = 8080'"
	}
	return ""
}

func quote(s string) string {
	if s == "" {
		return s
	}
	return "'" + s + "'"
}

func flattenCore(src *Source, ei core.ErrorInterface, ctx *pathCtx, depth int) []Diagnostic {
	code := lcerrors.ErrorCodeType(ei.GetCode())

	switch code {
	case shared.ContextedError:
		return flattenContexted(src, ei, depth)
	case lcerrors.ParsingError:
		d := Diagnostic{Code: CodeLex, Severity: SeverityError, Message: ei.GetMsg()}
		applyLexerPos(&d, ei, src)
		return []Diagnostic{d}
	case shared.ContractError:
		// A contract error aggregates the per-pair problems it found.
		if nested, ok := ei.GetMeta()["errs"].([]core.ErrorInterface); ok && len(nested) > 0 {
			out := []Diagnostic{}
			for _, sub := range nested {
				out = append(out, flatten(src, sub, ctx, depth+1)...)
			}
			return out
		}
		return []Diagnostic{fromMeta(ei, src, CodeContract, CodeContract)}
	default:
		return []Diagnostic{fromMeta(ei, src, mapCoreCode(code), CodeValue)}
	}
}

func fromMeta(ei core.ErrorInterface, src *Source, code, fallback string) Diagnostic {
	msg := ei.GetMsg()
	if msg == "" {
		msg = fallback
	}
	d := Diagnostic{Code: code, Severity: SeverityError, Message: msg}

	if hint := metaString(ei, "hint"); hint != "" {
		d.Hint = hint
	}
	if path := metaString(ei, "path"); path != "" {
		d.Path = path
	}
	if label := metaString(ei, "label"); label != "" {
		d.Label = label
	}
	if idx, ok := metaInt(ei, "index"); ok {
		value := idx
		d.Index = &value
	}
	if start, ok := metaInt(ei, "start"); ok {
		end := start + 1
		if e, ok := metaInt(ei, "end"); ok {
			end = e
		}
		if src != nil {
			d.Span = src.At(start, end)
		}
	}
	if d.Span == nil {
		if raw := metaString(ei, "raw"); raw != "" {
			d.Span = locateSpan(src, raw)
		}
	}
	return d
}

// flattenContexted expands a per-pair context error into the problems it
// contains. The wrapper carries no information a user can act on, so the
// causes become the top-level diagnostics and inherit the pair position.
func flattenContexted(src *Source, ei core.ErrorInterface, depth int) []Diagnostic {
	span := spanOf(ei, src)
	label := ""
	if idx, ok := metaInt(ei, "idx"); ok {
		label = ordinal(idx+1) + " pair"
	}

	children := []Diagnostic{}
	if errs, ok := ei.GetMeta()["errs"].([]core.ErrorInterface); ok {
		for _, sub := range errs {
			children = append(children, flatten(src, sub, nil, depth+1)...)
		}
	}

	if len(children) == 0 {
		d := Diagnostic{
			Code:     CodeType,
			Severity: SeverityError,
			Message:  ei.GetMsg(),
			Span:     span,
			Label:    label,
		}
		if raw := metaString(ei, "raw"); d.Message == "" && raw != "" {
			d.Message = strings.TrimSpace(raw)
		}
		return []Diagnostic{d}
	}

	for i := range children {
		if children[i].Span == nil {
			children[i].Span = span
		}
		if children[i].Label == "" {
			children[i].Label = label
		}
	}
	return children
}

// spanOf resolves the position recorded on an error, falling back to a text
// search for the raw snippet it reported.
func spanOf(ei core.ErrorInterface, src *Source) *Span {
	if start, ok := metaInt(ei, "start"); ok {
		end := start + 1
		if e, ok := metaInt(ei, "end"); ok {
			end = e
		}
		if src != nil {
			return src.At(start, end)
		}
	}
	return locateSpan(src, metaString(ei, "raw"))
}

func locateSpan(src *Source, raw string) *Span {
	if src == nil || raw == "" {
		return nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	offset := src.locate(trimmed)
	if offset < 0 {
		return nil
	}
	return src.At(offset, offset+len([]rune(trimmed)))
}

func applyLexerPos(d *Diagnostic, ei core.ErrorInterface, src *Source) {
	meta := ei.GetMeta()
	if meta == nil {
		return
	}
	line, ok1 := metaIntValue(meta, string(core.EMK(0, "int")))
	col, ok2 := metaIntValue(meta, string(core.EMK(1, "int")))
	if ok1 && ok2 && src != nil {
		offset := src.OffsetOf(line, col)
		d.Span = src.AtOffset(offset, offset+1)
		d.Span.Line = line
		d.Span.Column = col
	}
	if snippet, ok := meta[core.EMK(2, "string")].(string); ok && snippet != "" {
		d.Hint = "unexpected text near " + strconv.Quote(snippet)
	}
}

func metaInt(ei core.ErrorInterface, key string) (int, bool) {
	return metaIntValue(ei.GetMeta(), key)
}

func metaIntValue(meta map[lcerrors.ErrorMetaType]interface{}, key string) (int, bool) {
	if meta == nil {
		return 0, false
	}
	switch v := meta[lcerrors.ErrorMetaType(key)].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	}
	return 0, false
}

func metaString(ei core.ErrorInterface, key string) string {
	meta := ei.GetMeta()
	if meta == nil {
		return ""
	}
	if s, ok := meta[lcerrors.ErrorMetaType(key)].(string); ok {
		return s
	}
	return ""
}

func mapCoreCode(code lcerrors.ErrorCodeType) string {
	switch code {
	case shared.RuntimeError:
		return CodeValue
	case shared.ProcessingError:
		return CodeType
	case shared.WrappedError:
		return CodeInternal
	case shared.ContractError:
		return CodeContract
	}
	if code == "" {
		return CodeInternal
	}
	return CodeInternal
}

func asError[T error](err error, target *T) bool {
	for err != nil {
		if v, ok := err.(T); ok {
			*target = v
			return true
		}
		unwrapped, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrapped.Unwrap()
	}
	return false
}

// NodeSpan resolves a parsed token to a span, so a diagnostic can point at
// the exact token that caused it.
func NodeSpan(src *Source, node *stringParsing.ParsedNode) *Span {
	if src == nil || node == nil {
		return nil
	}
	start, ok := node.Metadata["__start"].(int)
	if !ok {
		return nil
	}
	end, _ := node.Metadata["__end"].(int)
	if end <= start {
		end = start + 1
	}
	return src.At(start, end)
}
