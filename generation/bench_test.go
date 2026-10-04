package generation_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

func benchConfig(groups, pairs int) string {
	var b strings.Builder
	b.WriteString("{\n")
	for g := 0; g < groups; g++ {
		fmt.Fprintf(&b, "\tserver%d: object = {\n", g)
		for p := 0; p < pairs; p++ {
			fmt.Fprintf(&b, "\t\thost%d: string = \"host-%d-%d\",\n", p, g, p)
			fmt.Fprintf(&b, "\t\tport%d: int = %d,\n", p, 8000+p)
			fmt.Fprintf(&b, "\t\ttags%d: strings = [\"a\", \"b\", \"c\"],\n", p)
			fmt.Fprintf(&b, "\t\tports%d: ints = [%d, %d, %d],\n", p, p, p+1, p+2)
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}")
	return b.String()
}

func benchParse(b *testing.B, groups, pairs int) *shared.Config {
	b.Helper()
	cfg, err := lang.ParseConf(shared.NewNilConfig(), benchConfig(groups, pairs), false)
	if err != nil {
		b.Fatalf("parse: %v", err)
	}
	return cfg
}

func BenchmarkTycl(b *testing.B) {
	cfg := benchParse(b, 4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := generation.Tycl(cfg); err != nil {
			b.Fatalf("tycl: %v", err)
		}
	}
}

func BenchmarkJson(b *testing.B) {
	cfg := benchParse(b, 4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := generation.Json(cfg); err != nil {
			b.Fatalf("json: %v", err)
		}
	}
}

func BenchmarkYaml(b *testing.B) {
	cfg := benchParse(b, 4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := generation.Yaml(cfg); err != nil {
			b.Fatalf("yaml: %v", err)
		}
	}
}

