package lang

import (
	"strings"
	"testing"

	"github.com/pt-main/tycl/shared"
)

const sampleConfig = `{
	port: int = 8080,
	host: string = "localhost",
	debug: bool = true,
	timeout: int = null,
	ports: ints = [1, 2, 3],
	server: object = { host: string = "h", port: int = 80 },
	servers: objects = [{ host: string = "a" }, { host: string = "b" }],
}`

func load(t *testing.T, code string) *shared.Config {
	t.Helper()
	cfg, err := ParseConf(shared.NewNilConfig(), code, false)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return cfg
}

func TestResolveTopLevel(t *testing.T) {
	cfg := load(t, sampleConfig)
	entry, ok := Resolve(cfg, "port")
	if !ok {
		t.Fatal("port not found")
	}
	if entry.Type != "int" || entry.Value != 8080 {
		t.Errorf("got %s %v, want int 8080", entry.Type, entry.Value)
	}
}

func TestResolveNested(t *testing.T) {
	cfg := load(t, sampleConfig)
	entry, ok := Resolve(cfg, "server.host")
	if !ok {
		t.Fatal("server.host not found")
	}
	if entry.Value != "h" {
		t.Errorf("got %v, want h", entry.Value)
	}
}

func TestResolveArrayIndex(t *testing.T) {
	cfg := load(t, sampleConfig)
	entry, ok := Resolve(cfg, "servers.1.host")
	if !ok {
		t.Fatal("servers.1.host not found")
	}
	if entry.Value != "b" {
		t.Errorf("got %v, want b", entry.Value)
	}
}

func TestResolveTypedNull(t *testing.T) {
	cfg := load(t, sampleConfig)
	entry, ok := Resolve(cfg, "timeout")
	if !ok {
		t.Fatal("timeout not found")
	}
	if entry.Type != "int" || entry.Value != nil {
		t.Errorf("got %s %v, want int nil", entry.Type, entry.Value)
	}
}

func TestResolveUnknownPath(t *testing.T) {
	cfg := load(t, sampleConfig)
	if _, ok := Resolve(cfg, "serverht"); ok {
		t.Error("unknown path resolved")
	}
}

func TestSetWritesTypedValue(t *testing.T) {
	cfg := load(t, sampleConfig)
	if err := Set(cfg, "port", "int", "9090"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if cfg.IntV["port"] != 9090 {
		t.Errorf("port = %d, want 9090", cfg.IntV["port"])
	}
}

func TestSetReplacesTypeOfExistingKey(t *testing.T) {
	cfg := load(t, sampleConfig)
	if err := Set(cfg, "port", "string", "http"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, stale := cfg.IntV["port"]; stale {
		t.Error("old int entry survived a type change")
	}
	if cfg.StringV["port"] != "http" {
		t.Errorf("port = %q, want http", cfg.StringV["port"])
	}
}

func TestSetTypedNull(t *testing.T) {
	cfg := load(t, sampleConfig)
	if err := Set(cfg, "timeout", "int", "null"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if cfg.NullV["timeout"] != "int" {
		t.Errorf("timeout = %q, want int", cfg.NullV["timeout"])
	}
}

func TestSetRejectsUnknownType(t *testing.T) {
	cfg := load(t, sampleConfig)
	if err := Set(cfg, "port", "number", "1"); err == nil {
		t.Error("unknown type accepted")
	}
}

func TestSetArrayKeepsNestedObjects(t *testing.T) {
	cfg := load(t, sampleConfig)
	value := `[{host: string = "a", port: int = 1}, {host: string = "b", port: int = 2}]`
	if err := Set(cfg, "servers", "objects", value); err != nil {
		t.Fatalf("set: %v", err)
	}
	arr := cfg.InnerArrV["servers"]
	if len(arr) != 2 {
		t.Fatalf("got %d elements, want 2", len(arr))
	}
	if arr[0].StringV["host"] != "a" || arr[1].IntV["port"] != 2 {
		t.Errorf("unexpected element values: %+v %+v", arr[0].StringV, arr[1].IntV)
	}
}

func TestRemoveKey(t *testing.T) {
	cfg := load(t, sampleConfig)
	if !Remove(cfg, "port") {
		t.Fatal("port not removed")
	}
	if _, ok := cfg.IntV["port"]; ok {
		t.Error("port still present")
	}
	if Remove(cfg, "port") {
		t.Error("second remove reported success")
	}
}

func TestMergeIntoOverridesScalarsAndMergesObjects(t *testing.T) {
	dst := load(t, sampleConfig)
	src := load(t, `{ port: int = 443, server: object = { port: int = 443 } }`)

	MergeInto(dst, src)

	if dst.IntV["port"] != 443 {
		t.Errorf("port = %d, want 443", dst.IntV["port"])
	}
	if dst.InnerV["server"].IntV["port"] != 443 {
		t.Errorf("server.port not overridden")
	}
	if dst.InnerV["server"].StringV["host"] != "h" {
		t.Error("server.host was lost during merge")
	}
	if dst.StringV["host"] != "localhost" {
		t.Error("untouched key was dropped during merge")
	}
}

func TestInferType(t *testing.T) {
	cases := map[string]string{
		"8080":     "int",
		"1.5":      "float",
		"true":     "bool",
		"text":     "string",
		"null":     "string",
		`"quoted"`: "string",
	}
	for raw, want := range cases {
		if got := InferType(raw); got != want {
			t.Errorf("InferType(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestSplitTopLevelIgnoresNestedCommas(t *testing.T) {
	got := splitTopLevel(`{a: int = 1, b: int = 2}, {a: int = 3}`)
	if len(got) != 2 {
		t.Fatalf("got %d parts, want 2: %q", len(got), got)
	}
}

func TestTypeSuggestionFindsTypos(t *testing.T) {
	if hint := typeSuggestion("in"); hint == "" {
		t.Error("no hint for a close typo")
	}
	if hint := typeSuggestion("qqqqqq"); hint == "" {
		t.Error("no fallback hint for an unrelated type")
	}
}

func TestParseReportsEveryBadPair(t *testing.T) {
	code := `{
		port: int = 8080,
		timeout = null,
		rate: int = 1.5,
		numbr: strig = 5,
	}`
	_, err := ParseConf(shared.NewNilConfig(), code, false)
	if err == nil {
		t.Fatal("expected errors")
	}
	list := diagList(code, err)
	if len(list) != 3 {
		t.Fatalf("want 3 problems, got %d: %+v", len(list), list)
	}
}

func TestParseLabelsTheRightPair(t *testing.T) {
	code := "{\n\tok: int = 1,\n\tbad: strig = 5,\n}"
	_, err := ParseConf(shared.NewNilConfig(), code, false)
	if err == nil {
		t.Fatal("expected errors")
	}
	list := diagList(code, err)
	if len(list) != 1 {
		t.Fatalf("want 1 problem, got %d", len(list))
	}
	if list[0].Label != "2nd pair" {
		t.Errorf("label = %q, want %q", list[0].Label, "2nd pair")
	}
}

func TestStrictKeysRejectsDuplicates(t *testing.T) {
	code := "{\n\tdup: int = 1,\n\tdup: int = 2,\n}"
	_, err := ParseConf(shared.NewNilConfig(), code, true)
	if err == nil {
		t.Fatal("strict keys must reject duplicates")
	}
	list := diagList(code, err)
	if len(list) == 0 {
		t.Fatal("no diagnostics")
	}
	if !strings.Contains(list[0].Message, "Duplicate") {
		t.Errorf("unexpected message: %q", list[0].Message)
	}
}
