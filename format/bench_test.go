package format

import (
	"fmt"
	"strings"
	"testing"
)

func benchCode(groups, pairs int) string {
	var b strings.Builder
	b.WriteString("{\n")
	for g := 0; g < groups; g++ {
		fmt.Fprintf(&b, "\tserver%d: object = {\n", g)
		for p := 0; p < pairs; p++ {
			fmt.Fprintf(&b, "\t\thost%d: string = \"host-%d-%d\",\n", p, g, p)
			fmt.Fprintf(&b, "\t\tport%d: int = %d,\n", p, 8000+p)
			fmt.Fprintf(&b, "\t\ttags%d: strings = [\"a\", \"b\", \"c\"],\n", p)
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}")
	return b.String()
}

func BenchmarkFormConfigSmall(b *testing.B) {
	code := benchCode(1, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := FormConfig(code); err != nil {
			b.Fatalf("format: %v", err)
		}
	}
}

func BenchmarkFormConfigMedium(b *testing.B) {
	code := benchCode(4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := FormConfig(code); err != nil {
			b.Fatalf("format: %v", err)
		}
	}
}

func BenchmarkFormConfigLarge(b *testing.B) {
	code := benchCode(20, 40)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := FormConfig(code); err != nil {
			b.Fatalf("format: %v", err)
		}
	}
}