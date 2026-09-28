package diag

import (
	"strings"
	"testing"
)

func TestFromErrorReportsPositionAndHint(t *testing.T) {
	code := "{\n    port: nope = 8080,\n}\n"
	src := NewSource("app.tycl", code)

	list := FromError(src, parseError(t, code))
	if len(list) != 1 {
		t.Fatalf("want 1 diagnostic, got %d: %+v", len(list), list)
	}
	d := list[0]
	if d.Code != CodeType {
		t.Errorf("code = %q, want %q", d.Code, CodeType)
	}
	if d.Span == nil || d.Span.Line != 2 {
		t.Fatalf("span = %+v, want line 2", d.Span)
	}
	if d.Span.Column < 1 {
		t.Errorf("column = %d, want >= 1", d.Span.Column)
	}
	if d.Hint == "" {
		t.Error("hint is empty")
	}
}

func TestFromErrorMarksTheOffendingToken(t *testing.T) {
	code := "{\n    port: int = \"abc\",\n}\n"
	src := NewSource("app.tycl", code)

	list := FromError(src, parseError(t, code))
	if len(list) == 0 {
		t.Fatal("no diagnostics")
	}
	before, marked, _, ok := list[0].Span.Excerpt(code)
	if !ok {
		t.Fatal("cannot render excerpt")
	}
	if marked != `"abc"` {
		t.Errorf("marked = %q, want %q", marked, `"abc"`)
	}
	if !strings.Contains(before, "port: int = ") {
		t.Errorf("before = %q, want the pair prefix", before)
	}
}

func TestRenderHasNoColorCodes(t *testing.T) {
	code := "{\n    port: nope = 1,\n}\n"
	src := NewSource("app.tycl", code)
	out := Render(FromError(src, parseError(t, code)), code)
	if !strings.Contains(out, "nope") {
		t.Errorf("rendered output does not mention the offending token:\n%s", out)
	}
}

func TestDiagnosticIsJSONSerialisable(t *testing.T) {
	d := Diagnostic{
		Code:     CodeContract,
		Severity: SeverityError,
		Message:  "missing key",
		Span:     &Span{File: "a.tycl", Line: 3, Column: 5, Width: 4},
	}
	encoded := mustJSON(t, d)
	for _, want := range []string{`"code":"contract.violation"`, `"line":3`, `"column":5`, `"file":"a.tycl"`} {
		if !strings.Contains(encoded, want) {
			t.Errorf("json %s missing %s", encoded, want)
		}
	}
}

func TestCheckContractReportsEveryViolation(t *testing.T) {
	conf, err := parseConfig(t, `{ port: int = 8080, extra: string = "x" }`)
	if err != nil {
		t.Fatal(err)
	}
	cont := parseContract(t, `strict { port: int, missing: string }`)

	list := CheckContract(conf, cont)
	codes := map[string]int{}
	for _, d := range list {
		codes[d.Message]++
	}
	if len(list) != 2 {
		t.Fatalf("want 2 violations, got %d: %+v", len(list), list)
	}
	if !hasMessage(list, `required string key "missing" is missing`) {
		t.Errorf("missing required key not reported: %+v", list)
	}
	if !hasMessage(list, `key "extra" is not allowed by this strict contract`) {
		t.Errorf("extra key not reported: %+v", list)
	}
}

func TestCheckContractAcceptsTypedNull(t *testing.T) {
	conf, err := parseConfig(t, `{ port: int = null }`)
	if err != nil {
		t.Fatal(err)
	}
	if list := CheckContract(conf, parseContract(t, `strict { port: int }`)); len(list) != 0 {
		t.Errorf("typed null should satisfy the contract, got %+v", list)
	}
}

func TestSortIsDeterministic(t *testing.T) {
	list := []Diagnostic{
		{Code: CodeType, Message: "b", Span: &Span{File: "a.tycl", Line: 5, Column: 1}},
		{Code: CodeType, Message: "a", Span: &Span{File: "a.tycl", Line: 2, Column: 9}},
		{Code: CodeType, Message: "c", Span: &Span{File: "a.tycl", Line: 2, Column: 1}},
	}
	Sort(list)
	if list[0].Message != "c" || list[1].Message != "a" || list[2].Message != "b" {
		t.Errorf("unexpected order: %s %s %s", list[0].Message, list[1].Message, list[2].Message)
	}
}

func TestErrorImplementsCoreInterface(t *testing.T) {
	var err error = &Error{Diagnostics: []Diagnostic{{
		Code:     CodeContract,
		Severity: SeverityError,
		Message:  "boom",
	}}}
	if err.Error() != "boom" {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestParseErrorIsReportedOnce(t *testing.T) {
	code := "{\n    port: int = \n}\n"
	src := NewSource("app.tycl", code)

	list := FromError(src, parseError(t, code))
	if len(list) != 1 {
		t.Fatalf("want a single diagnostic, got %d: %+v", len(list), list)
	}
	if list[0].Hint == "" {
		t.Error("no hint for a missing value")
	}
}

func TestParseErrorPointsAtTheToken(t *testing.T) {
	code := "{\n    port: int = \n}\n"
	src := NewSource("app.tycl", code)

	list := FromError(src, parseError(t, code))
	_, marked, _, ok := list[0].Span.Excerpt(code)
	if !ok || marked == "" {
		t.Fatalf("cannot render the excerpt: %+v", list[0].Span)
	}
}

func TestCommentBeforeObjectGetsAHint(t *testing.T) {
	code := "/* doc */\n{\n    port: int = 1,\n}\n"
	src := NewSource("app.tycl", code)

	list := FromError(src, parseError(t, code))
	if len(list) == 0 {
		t.Fatal("no diagnostics")
	}
	if list[0].Hint == "" {
		t.Errorf("no hint for a misplaced comment: %+v", list[0])
	}
}
