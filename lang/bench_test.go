package lang

import (
	"fmt"
	"strings"
	"testing"

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
			fmt.Fprintf(&b, "\t\tratio%d: float = %d.5,\n", p, p)
			fmt.Fprintf(&b, "\t\tflag%d: bool = true,\n", p)
			fmt.Fprintf(&b, "\t\ttags%d: strings = [\"a\", \"b\", \"c\"],\n", p)
			fmt.Fprintf(&b, "\t\tports%d: ints = [%d, %d, %d],\n", p, p, p+1, p+2)
		}
		fmt.Fprintf(&b, "\t\tempty: int = null,\n")
		b.WriteString("\t},\n")
	}
	b.WriteString("}")
	return b.String()
}

func benchConfigOnce(b *testing.B, code string) *shared.Config {
	b.Helper()
	cfg, err := ParseConf(shared.NewNilConfig(), code, false)
	if err != nil {
		b.Fatalf("parse: %v", err)
	}
	return cfg
}

func BenchmarkParseConfSmall(b *testing.B) {
	code := benchConfig(1, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchConfigOnce(b, code)
	}
}

func BenchmarkParseConfMedium(b *testing.B) {
	code := benchConfig(4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchConfigOnce(b, code)
	}
}

func BenchmarkParseConfLarge(b *testing.B) {
	code := benchConfig(20, 40)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchConfigOnce(b, code)
	}
}

func BenchmarkResolvePath(b *testing.B) {
	cfg := benchConfigOnce(b, benchConfig(4, 10))
	paths := []string{
		"server0.host3",
		"server1.port5",
		"server2.tags7",
		"server3.ports9",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range paths {
			if _, ok := Resolve(cfg, p); !ok {
				b.Fatalf("path %q not found", p)
			}
		}
	}
}