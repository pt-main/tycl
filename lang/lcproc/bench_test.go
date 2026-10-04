package lcproc

import (
	"fmt"
	"strings"
	"testing"
)

// benchCode builds a config of the same shape the lang benchmarks use.
func benchCode(groups, pairs int) string {
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

// BenchmarkLexParse measures only the lc lexer+parser, with no tycl layer on
// top. It is the floor cost of the dependency for a given input size.
func BenchmarkLexParseSmall(b *testing.B) {
	code := benchCode(1, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := NewParser()
		if err != nil {
			b.Fatalf("parser: %v", err)
		}

		if _, err := p.Parse(code); err != nil {
			b.Fatalf("parse: %v", err)
		}
	}
}

func BenchmarkLexParseMedium(b *testing.B) {
	code := benchCode(4, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := NewParser()
		if err != nil {
			b.Fatalf("parser: %v", err)
		}

		if _, err := p.Parse(code); err != nil {
			b.Fatalf("parse: %v", err)
		}
	}
}

func BenchmarkLexParseLarge(b *testing.B) {
	code := benchCode(20, 40)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, err := NewParser()
		if err != nil {
			b.Fatalf("parser: %v", err)
		}
		if _, err := p.Parse(code); err != nil {
			b.Fatalf("parse: %v", err)
		}
	}
}

// BenchmarkNewParser isolates the fixed cost of building a parser, which
// recompiles every regexp on each call.
func BenchmarkNewParser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewParser(); err != nil {
			b.Fatalf("parser: %v", err)
		}
	}
}

func BenchmarkNewLexer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewLexer(); err != nil {
			b.Fatalf("lexer: %v", err)
		}
	}
}

func BenchmarkNewGrammar(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = createGrammar()
	}
}
