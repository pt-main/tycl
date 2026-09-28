package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validConfig = `{
	port: int = 8080,
	host: string = "localhost",
	debug: bool = true,
	timeout: int = null,
	ports: ints = [1, 2, 3],
	server: object = { host: string = "h", port: int = 80 },
}`

const brokenConfig = `{
	port: nope = 8080,
}`

// run executes a command line and captures its streams and exit code.
func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	os.Stdout, os.Stderr = outW, errW

	done := make(chan [2]string, 1)
	go func() {
		var outBuf, errBuf bytes.Buffer
		_, _ = outBuf.ReadFrom(outR)
		_, _ = errBuf.ReadFrom(errR)
		done <- [2]string{outBuf.String(), errBuf.String()}
	}()

	code = MainWithArgs(args)

	outW.Close()
	errW.Close()
	captured := <-done
	os.Stdout, os.Stderr = oldOut, oldErr

	return code, captured[0], captured[1]
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func decode(t *testing.T, out string) Envelope {
	t.Helper()
	var env Envelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("cannot decode envelope %q: %v", out, err)
	}
	return env
}

func TestValidSucceedsWithExitZero(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, _, _ := run(t, "valid", path)
	if code != CodeOK {
		t.Errorf("exit = %d, want %d", code, CodeOK)
	}
}

func TestValidFailsWithExitOne(t *testing.T) {
	path := writeFile(t, "app.tycl", brokenConfig)
	code, _, stderr := run(t, "valid", path)
	if code != CodeFail {
		t.Errorf("exit = %d, want %d", code, CodeFail)
	}
	if !strings.Contains(stderr, "nope") {
		t.Errorf("stderr does not explain the problem:\n%s", stderr)
	}
}

func TestValidJSONReportsDiagnostics(t *testing.T) {
	path := writeFile(t, "app.tycl", brokenConfig)
	code, stdout, stderr := run(t, "valid", path, "--json")

	if code != CodeFail {
		t.Errorf("exit = %d, want %d", code, CodeFail)
	}
	if stderr != "" {
		t.Errorf("stderr must stay empty in json mode, got: %s", stderr)
	}
	env := decode(t, stdout)
	if env.OK {
		t.Error("ok = true, want false")
	}
	if len(env.Diagnostics) == 0 {
		t.Fatal("no diagnostics reported")
	}
	d := env.Diagnostics[0]
	if d.Code == "" || d.Message == "" {
		t.Errorf("diagnostic lacks code or message: %+v", d)
	}
	if d.Span == nil || d.Span.Line != 2 {
		t.Errorf("diagnostic lacks a position: %+v", d.Span)
	}
}

func TestJSONEnvelopeIsStableOnSuccess(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "valid", path, "--json")
	if code != CodeOK {
		t.Fatalf("exit = %d, want %d", code, CodeOK)
	}
	env := decode(t, stdout)
	if !env.OK {
		t.Error("ok = false, want true")
	}
	if env.Diagnostics == nil {
		t.Error("diagnostics must be an empty array, not null")
	}
}

func TestContractViolationIsReported(t *testing.T) {
	cfg := writeFile(t, "app.tycl", validConfig)
	cont := writeFile(t, "schema.tycl", `strict { port: int, missing: string }`)

	code, stdout, _ := run(t, "valid", cfg, cont, "--json")
	if code != CodeFail {
		t.Errorf("exit = %d, want %d", code, CodeFail)
	}
	env := decode(t, stdout)
	found := false
	for _, d := range env.Diagnostics {
		if strings.Contains(d.Message, "missing") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing key not reported: %+v", env.Diagnostics)
	}
}

func TestSyntaxReportsEveryFile(t *testing.T) {
	good := writeFile(t, "good.tycl", validConfig)
	bad := writeFile(t, "bad.tycl", brokenConfig)

	code, stdout, _ := run(t, "syntax", good, bad, "--json")
	if code != CodeFail {
		t.Errorf("exit = %d, want %d", code, CodeFail)
	}
	env := decode(t, stdout)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected data: %+v", env.Data)
	}
	files, ok := data["files"].([]any)
	if !ok || len(files) != 2 {
		t.Fatalf("want 2 file results, got %+v", data["files"])
	}
}

func TestGetPrintsBareValue(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "get", path, "port")
	if code != CodeOK {
		t.Fatalf("exit = %d, want %d", code, CodeOK)
	}
	if strings.TrimSpace(stdout) != "8080" {
		t.Errorf("stdout = %q, want 8080", stdout)
	}
}

func TestGetNestedPath(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	_, stdout, _ := run(t, "get", path, "server.host")
	if strings.TrimSpace(stdout) != "h" {
		t.Errorf("stdout = %q, want h", stdout)
	}
}

func TestGetUnknownPathSuggestsAlternative(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, _, stderr := run(t, "get", path, "serverht")
	if code != CodeFail {
		t.Errorf("exit = %d, want %d", code, CodeFail)
	}
	if !strings.Contains(stderr, "did you mean") {
		t.Errorf("no suggestion offered:\n%s", stderr)
	}
}

func TestSetThenGetRoundTrip(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)

	if code, _, stderr := run(t, "set", path, "port", "auto", "9090"); code != CodeOK {
		t.Fatalf("set failed: %d\n%s", code, stderr)
	}
	_, stdout, _ := run(t, "get", path, "port")
	if strings.TrimSpace(stdout) != "9090" {
		t.Errorf("value = %q, want 9090", stdout)
	}
}

func TestSetWritesIntoNestedObject(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	if code, _, _ := run(t, "set", path, "server.port", "int", "9091"); code != CodeOK {
		t.Fatalf("set failed")
	}
	_, stdout, _ := run(t, "get", path, "server.port")
	if strings.TrimSpace(stdout) != "9091" {
		t.Errorf("value = %q, want 9091", stdout)
	}
}

func TestSetRejectsUnknownTypeWithUsageExit(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, _, _ := run(t, "set", path, "port", "number", "1")
	if code != CodeUsage {
		t.Errorf("exit = %d, want %d", code, CodeUsage)
	}
}

func TestRemoveDeletesKey(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	if code, _, _ := run(t, "remove", path, "port"); code != CodeOK {
		t.Fatalf("remove failed")
	}
	code, _, _ := run(t, "get", path, "port")
	if code != CodeFail {
		t.Errorf("key survived removal: exit = %d", code)
	}
}

func TestGenWritesJSONFile(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	out := filepath.Join(t.TempDir(), "app.json")

	if code, _, stderr := run(t, "gen", path, out, "json"); code != CodeOK {
		t.Fatalf("gen failed: %d\n%s", code, stderr)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("output is not valid json: %v", err)
	}
	if decoded["port"] != float64(8080) {
		t.Errorf("port = %v, want 8080", decoded["port"])
	}
}

func TestGenToStdoutStreamsTheDocument(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "gen", path, "-", "json")
	if code != CodeOK {
		t.Fatalf("exit = %d, want %d", code, CodeOK)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid json: %v\n%s", err, stdout)
	}
}

func TestGenToStdoutIgnoresJSONFlag(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "gen", path, "-", "json", "--json")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stdout, "\"port\"") {
		t.Errorf("stdout is not the generated document:\n%s", stdout)
	}
}

func TestContractGenerationIsStable(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	first := ""
	for i := 0; i < 5; i++ {
		code, stdout, _ := run(t, "contract", path, "-", "strict")
		if code != CodeOK {
			t.Fatalf("contract failed")
		}
		if i == 0 {
			first = stdout
			continue
		}
		if stdout != first {
			t.Fatalf("contract output is not deterministic:\n%s\n---\n%s", first, stdout)
		}
	}
}

func TestMergeOverridesLaterFiles(t *testing.T) {
	base := writeFile(t, "base.tycl", `{
		port: int = 8080,
		server: object = { host: string = "h", port: int = 80 },
	}`)
	over := writeFile(t, "over.tycl", `{
		port: int = 443,
		server: object = { port: int = 443 },
	}`)

	code, stdout, stderr := run(t, "merge", base, over)
	if code != CodeOK {
		t.Fatalf("merge failed: %d\n%s", code, stderr)
	}
	for _, want := range []string{"port: int = 443", `host: string = "h"`} {
		if !strings.Contains(stdout, want) {
			t.Errorf("merge output misses %q:\n%s", want, stdout)
		}
	}
}

func TestMergeRequiresTwoFiles(t *testing.T) {
	base := writeFile(t, "base.tycl", validConfig)
	code, _, _ := run(t, "merge", base)
	if code != CodeUsage {
		t.Errorf("exit = %d, want %d", code, CodeUsage)
	}
}

func TestAstIsValidJSON(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "ast", path, "--json")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	env := decode(t, stdout)
	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected data: %+v", env.Data)
	}
	tree, ok := data["ast"].(map[string]any)
	if !ok || tree["type"] != "config" {
		t.Errorf("ast root is not a config node: %+v", data["ast"])
	}
}

func TestFormatIsIdempotent(t *testing.T) {
	path := writeFile(t, "app.tycl", `{
  port:int=8080,
  server:object={host:string="h",port:int=80},
}`)
	if code, _, _ := run(t, "fmt", "conf", path); code != CodeOK {
		t.Fatalf("first format failed")
	}
	first, _ := os.ReadFile(path)
	if code, _, _ := run(t, "fmt", "conf", path); code != CodeOK {
		t.Fatalf("second format failed")
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("formatting is not idempotent:\n%s\n---\n%s", first, second)
	}
}

func TestStructureListsEveryPath(t *testing.T) {
	path := writeFile(t, "app.tycl", validConfig)
	code, stdout, _ := run(t, "structure", path, "--json")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	env := decode(t, stdout)
	data := env.Data.(map[string]any)
	paths, ok := data["paths"].([]any)
	if !ok {
		t.Fatalf("paths missing: %+v", data)
	}
	want := map[string]bool{"port": false, "server.host": false, "ports": false}
	for _, p := range paths {
		if _, tracked := want[p.(string)]; tracked {
			want[p.(string)] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("path %q not listed: %+v", path, paths)
		}
	}
}

func TestTypesListsTheTypeSystem(t *testing.T) {
	code, stdout, _ := run(t, "types", "--json")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	env := decode(t, stdout)
	if env.Data == nil {
		t.Fatal("no data returned")
	}
}

func TestMissingFileIsAnIOError(t *testing.T) {
	code, _, stderr := run(t, "valid", filepath.Join(t.TempDir(), "nope.tycl"))
	if code != CodeIO {
		t.Errorf("exit = %d, want %d", code, CodeIO)
	}
	if !strings.Contains(stderr, "nope.tycl") {
		t.Errorf("stderr does not name the file:\n%s", stderr)
	}
}

func TestVersionCommand(t *testing.T) {
	code, stdout, _ := run(t, "version")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(stdout, "tycl") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestStdinIsAcceptedAsInput(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	go func() {
		_, _ = w.WriteString(validConfig)
		w.Close()
	}()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	code, stdout, stderr := run(t, "valid", "-")
	if code != CodeOK {
		t.Fatalf("exit = %d, want %d (%s)", code, CodeOK, stderr)
	}
	if !strings.Contains(stdout, "is valid") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestStdinToStdoutPipeline(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	go func() {
		_, _ = w.WriteString(validConfig)
		w.Close()
	}()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	code, stdout, _ := run(t, "gen", "-", "-", "json")
	if code != CodeOK {
		t.Fatalf("exit = %d", code)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid json: %v\n%s", err, stdout)
	}
	if decoded["port"] != float64(8080) {
		t.Errorf("port = %v", decoded["port"])
	}
}

func TestNoColorStripsAnsi(t *testing.T) {
	path := writeFile(t, "app.tycl", brokenConfig)
	_, _, stderr := run(t, "valid", path, "--no-color")
	if strings.Contains(stderr, "\x1b[") {
		t.Errorf("ansi codes present with --no-color:\n%q", stderr)
	}
}

func TestErrorOutputCarriesTheSourceLine(t *testing.T) {
	code := "{\n    port: nope = 8080,\n}\n"
	path := writeFile(t, "app.tycl", code)
	_, _, stderr := run(t, "valid", path, "--no-color")
	if !strings.Contains(stderr, "port: nope = 8080") {
		t.Errorf("source line not shown:\n%s", stderr)
	}
	if !strings.Contains(stderr, "^^^^") {
		t.Errorf("offending token not underlined:\n%s", stderr)
	}
}

func TestJSONDiagnosticNamesTheFileAndLine(t *testing.T) {
	path := writeFile(t, "app.tycl", brokenConfig)
	_, stdout, _ := run(t, "valid", path, "--json")
	env := decode(t, stdout)
	d := env.Diagnostics[0]
	if d.Span == nil {
		t.Fatal("no span")
	}
	if d.Span.File != path {
		t.Errorf("file = %q, want %q", d.Span.File, path)
	}
	if d.Span.Line != 2 {
		t.Errorf("line = %d, want 2", d.Span.Line)
	}
}

func TestSetThenFormatKeepsObjectArraysValid(t *testing.T) {
	const code = `{
		servers: objects = [
			{ host: string = "a", port: int = 80 },
			{ host: string = "b", port: int = 443 }
		],
	}`
	path := writeFile(t, "app.tycl", code)

	if status, _, stderr := run(t, "set", path, "servers.0.port", "int", "8081"); status != CodeOK {
		t.Fatalf("set failed: %d\n%s", status, stderr)
	}
	if status, _, stderr := run(t, "fmt", "conf", path); status != CodeOK {
		t.Fatalf("format failed: %d\n%s", status, stderr)
	}
	// The rewritten file must still parse and keep both elements.
	status, rendered, stderr := run(t, "gen", path, "-", "tycl")
	if status != CodeOK {
		t.Fatalf("regenerated file does not parse: %s", stderr)
	}
	if !strings.Contains(rendered, `host: string = "b"`) {
		t.Errorf("second element lost:\n%s", rendered)
	}
	if !strings.Contains(rendered, "port: int = 8081") {
		t.Errorf("edit lost:\n%s", rendered)
	}
}

func TestGenTyclRoundTrips(t *testing.T) {
	const code = `{
		port: int = 8080,
		server: object = { host: string = "h" },
		servers: objects = [{ host: string = "a" }, { host: string = "b" }],
	}`
	path := writeFile(t, "app.tycl", code)

	status, first, stderr := run(t, "gen", path, "-", "tycl")
	if status != CodeOK {
		t.Fatalf("generation failed: %s", stderr)
	}
	_, second, _ := run(t, "gen", path, "-", "tycl")
	if first != second {
		t.Errorf("generation is not deterministic:\n%s\n---\n%s", first, second)
	}
	if !strings.Contains(first, "},") {
		t.Errorf("object array items are not separated:\n%s", first)
	}
}
