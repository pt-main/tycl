package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pt-main/tap/go/color"
	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/utils"
)

// Exit codes follow the usual convention: 0 success, 1 data error,
// 2 usage error, 3 I/O error.
const (
	CodeOK    = 0
	CodeFail  = 1
	CodeUsage = 2
	CodeIO    = 3
)

// StdinPath is the conventional placeholder for standard input.
const StdinPath = "-"

// ExitError carries a process exit code through the error chain.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit status %d", e.Code)
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }

// exitCode extracts the intended exit status from an error.
func exitCode(err error) int {
	if err == nil {
		return CodeOK
	}
	var ee *ExitError
	if asExit(err, &ee) {
		return ee.Code
	}
	return CodeFail
}

func asExit(err error, target **ExitError) bool {
	for err != nil {
		if v, ok := err.(*ExitError); ok {
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

func usage(format string, args ...any) error {
	return &ExitError{Code: CodeUsage, Err: diag.Errorf(diag.CodeUsage, format, args...)}
}

// IOError reports a file that could not be read or written.
func IOError(path string, err error) error {
	return &ExitError{
		Code: CodeIO,
		Err:  diag.Errorf(diag.CodeIO, "%s: %v", path, err),
	}
}

// Input is a resolved source of TYCL text.
type Input struct {
	Name   string
	Text   string
	Source *diag.Source
}

// ReadInput loads a file, or standard input when the path is "-".
func ReadInput(path string) (*Input, error) {
	if path == StdinPath {
		text, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, IOError("stdin", err)
		}
		name := "<stdin>"
		return &Input{Name: name, Text: string(text), Source: diag.NewSource(name, string(text))}, nil
	}
	text, err := utils.OpenF(path)
	if err != nil {
		return nil, IOError(path, err)
	}
	return &Input{Name: path, Text: text, Source: diag.NewSource(path, text)}, nil
}

// WriteOutput writes to a file, or to standard output when the path is "-".
func WriteOutput(path, data string) error {
	if path == StdinPath || path == "" {
		if !strings.HasSuffix(data, "\n") {
			data += "\n"
		}
		_, err := os.Stdout.WriteString(data)
		return err
	}
	if err := utils.WriteF(path, data); err != nil {
		return IOError(path, err)
	}
	return nil
}

// Envelope is the stable JSON shape returned by every command.
type Envelope struct {
	OK          bool              `json:"ok"`
	Command     string            `json:"command"`
	Diagnostics []diag.Diagnostic `json:"diagnostics"`
	Data        any               `json:"data,omitempty"`
}

// Result is what a handler returns: payload plus diagnostics, without
// deciding how they should be printed.
type Result struct {
	Data        any
	Diagnostics []diag.Diagnostic
	Message     string
	// Stream marks a payload that is the command's real output: it is
	// written verbatim even in JSON mode, because the caller asked for a
	// document on stdout.
	Stream bool
	// Exit overrides the status used when the result carries diagnostics.
	Exit int
	// Sources maps a file name to the text a diagnostic in that file should
	// be rendered against, so the excerpt shows the real source line.
	Sources map[string]string
	// Source is the fallback text when a span names no known file.
	Source string
}

// Ok builds a successful Result.
func Ok(data any, message string) *Result {
	return &Result{Data: data, Message: message}
}

// Document builds a Result whose payload is printed as-is.
func Document(text string) *Result {
	return &Result{Data: text, Stream: true}
}

// Fail builds a Result carrying diagnostics.
func Fail(list ...diag.Diagnostic) *Result {
	return &Result{Diagnostics: list}
}

// Failf builds a Result carrying a single diagnostic.
func Failf(code, format string, args ...any) *Result {
	return &Result{Diagnostics: []diag.Diagnostic{{
		Code:     code,
		Severity: diag.SeverityError,
		Message:  fmt.Sprintf(format, args...),
	}}}
}

// Usagef reports a misuse of the command line.
func Usagef(format string, args ...any) *Result {
	res := Failf(diag.CodeUsage, format, args...)
	res.Exit = CodeUsage
	return res
}

// ctxSources picks the text a diagnostic should be rendered against: the
// file named in its span when it is known, otherwise the default text.
func ctxSources(res *Result, fallback string) map[string]string {
	out := map[string]string{}
	for name, body := range res.Sources {
		out[name] = body
	}
	if res.Source != "" {
		out[""] = res.Source
	} else {
		out[""] = fallback
	}
	return out
}

// Report prints a result in the requested format and returns the exit code.
//
// In JSON mode nothing but the envelope reaches stdout and stderr stays
// empty, so a caller can always parse stdout without filtering noise.
func Report(command string, res *Result, format string, text string) int {
	if format == FormatJSON {
		return reportJSON(command, res)
	}
	return reportHuman(res, text)
}

func reportJSON(command string, res *Result) int {
	// A streamed document stays the sole content of stdout, so pipelines
	// like `tycl gen - - json | jq` keep working with --json set.
	if res.Stream {
		if text, ok := res.Data.(string); ok {
			fmt.Println(text)
		}
		if !diag.HasErrors(res.Diagnostics) {
			return CodeOK
		}
	}

	env := Envelope{
		OK:          !diag.HasErrors(res.Diagnostics),
		Command:     command,
		Diagnostics: res.Diagnostics,
		Data:        res.Data,
	}
	if env.Diagnostics == nil {
		env.Diagnostics = []diag.Diagnostic{}
	}
	encoded, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, diag.Failure("cannot encode result: "+err.Error()))
		return CodeFail
	}
	fmt.Println(string(encoded))
	if env.OK {
		return CodeOK
	}
	return statusOf(res)
}

// statusOf returns the exit code a failed result should produce.
func statusOf(res *Result) int {
	if res.Exit != 0 {
		return res.Exit
	}
	return CodeFail
}

func reportHuman(res *Result, text string) int {
	if diag.HasErrors(res.Diagnostics) {
		if rendered := diag.RenderFrom(res.Diagnostics, ctxSources(res, text)); rendered != "" {
			fmt.Fprint(os.Stderr, rendered)
		}
		fmt.Fprintln(os.Stderr, diag.Failure(diag.Summary(res.Diagnostics)))
		return statusOf(res)
	}
	for _, d := range res.Diagnostics {
		if d.Severity == diag.SeverityWarning {
			fmt.Fprintln(os.Stderr, diag.Render([]diag.Diagnostic{d}, text))
		}
	}
	switch payload := res.Data.(type) {
	case nil:
	case string:
		// A command that streams a document prints it verbatim.
		if payload != "" {
			fmt.Print(payload)
			if !strings.HasSuffix(payload, "\n") {
				fmt.Println()
			}
			return CodeOK
		}
	default:
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err == nil {
			fmt.Println(string(encoded))
			return CodeOK
		}
	}
	if res.Message != "" {
		fmt.Println(diag.Success(res.Message))
	}
	return CodeOK
}

// applyColorMode honours --json and --no-color before any output is written.
//
// Both the documented --no-color spelling and the --no_color one tap itself
// understands are accepted, so a flag copied from another tap-based tool
// keeps working.
func applyColorMode(format string, flags map[string]string) {
	if format == FormatJSON {
		color.ColorEnabled = false
		return
	}
	if _, ok := flags["no-color"]; ok {
		color.ColorEnabled = false
		return
	}
	if _, ok := flags["no_color"]; ok {
		color.ColorEnabled = false
	}
}
