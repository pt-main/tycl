package lang

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	"github.com/pt-main/tycl/shared"
)

// offsetOfFirstToken returns the position of the left-most token of a node.
//
// Composite nodes are rebuilt from their children during parsing and carry no
// position of their own, so the first leaf is used to anchor diagnostics.
func (cp *configParser) offsetOfFirstToken(node *stringParsing.ParsedNode) int {
	if start, ok := cp.firstToken(node); ok {
		return start
	}
	return cp.Base
}

func (cp *configParser) firstToken(node *stringParsing.ParsedNode) (int, bool) {
	if node == nil {
		return 0, false
	}
	if start, ok := node.Metadata["__start"].(int); ok {
		return cp.Base + start, true
	}
	for _, child := range astools.GetChildren(node) {
		if start, ok := cp.firstToken(&child); ok {
			return start, true
		}
	}
	return 0, false
}

func duplicateKey(key, vtype string) core.ErrorInterface {
	return shared.HintError(
		core.Err(shared.RuntimeError, "Key %q is already defined as %s", key, vtype),
		"give the key a unique name, or remove the earlier definition",
	)
}

var knownTypes = []string{
	"null", "bool", "int", "float", "string", "object",
	"bools", "ints", "floats", "strings", "objects",
}

// typeSuggestion turns an unknown type into an actionable hint.
func typeSuggestion(vtype string) string {
	if vtype == "" {
		return "every value needs a type: 'key: int = 1' or 'key = 1'"
	}
	if near := closestType(vtype); near != "" {
		return fmt.Sprintf("unknown type %q, did you mean %q?", vtype, near)
	}
	if strings.HasSuffix(vtype, "s") {
		return fmt.Sprintf("arrays are declared with a plural type: int, float, bool, string, object (or %ss)", strings.TrimSuffix(vtype, "s"))
	}
	return fmt.Sprintf("known types: %s", strings.Join(knownTypes, ", "))
}

// closestType finds the known type with the smallest edit distance.
func closestType(vtype string) string {
	best := ""
	bestScore := 3
	lower := strings.ToLower(vtype)
	for _, known := range knownTypes {
		if score := editDistance(lower, known); score < bestScore {
			bestScore = score
			best = known
		}
	}
	return best
}

// editDistance is the Levenshtein distance between two short words.
func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	return min(min(a, b), c)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ValidTypes lists the types accepted by file/contract subcommands.
func ValidTypes() []string {
	out := make([]string, len(knownTypes))
	copy(out, knownTypes)
	return out
}

// valueSuggestion explains why a value does not match its declared type.
func valueSuggestion(vtype, value string) string {
	trimmed := strings.TrimSpace(value)
	switch vtype {
	case "int":
		if strings.Contains(trimmed, ".") {
			return "the value looks like a float: declare it as 'float' or drop the fraction"
		}
		if trimmed == "true" || trimmed == "false" {
			return "the value is a bool: declare it as 'bool' or use 1/0"
		}
		return "write a whole number, e.g. 8080"
	case "float":
		if trimmed != "" && !strings.ContainsAny(trimmed, ".eE") {
			return "a float needs a fraction or exponent, e.g. 1.0"
		}
		return "write a number, e.g. 1.5"
	case "bool":
		if trimmed == "1" || trimmed == "0" {
			return "write 'true' or 'false' instead of 1 or 0"
		}
		return "write 'true' or 'false'"
	case "string":
		if trimmed == "null" {
			return "'null' is a null value: declare the type explicitly, e.g. 'key: string = null'"
		}
		return "wrap the value in quotes: \"text\" or 'text'"
	case "object":
		return "an object needs a body: 'key: object = { other: int = 1 }'"
	}
	return ""
}

// IsValidType reports whether a TYCL type name is known.
func IsValidType(vtype string) bool {
	return slices.Contains(knownTypes, vtype)
}
