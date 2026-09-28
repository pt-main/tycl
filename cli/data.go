package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

// valuePayload is the JSON shape of a single value read from a config.
type valuePayload struct {
	Path  string `json:"path"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// queryHandler implements `tycl query`: read values by dot-path.
func queryHandler(ctx *Ctx, args []string) *Result {
	in, cfg, res := loadConfig(ctx, args[0])
	if res != nil {
		return res
	}
	paths := cleanOptional(args[1:])
	if len(paths) == 0 {
		return Ok(map[string]any{
			"file":  in.Name,
			"paths": ListPaths(cfg, ""),
			"tree":  structureOf(cfg),
		}, "")
	}

	values := []valuePayload{}
	missing := []string{}
	for _, path := range paths {
		entry, ok := lang.Resolve(cfg, path)
		if !ok {
			missing = append(missing, path)
			continue
		}
		values = append(values, valuePayload{Path: path, Type: entry.Type, Value: plainValue(entry)})
	}

	if len(values) == 0 {
		return Fail(lookupFailure(in.Name, missing, cfg)...)
	}

	// In human mode a single lookup prints only the value, so the command
	// composes with shell substitution.
	if ctx.Format != FormatJSON && len(values) == 1 {
		return Ok(scalarText(values[0].Type, values[0].Value), "")
	}

	payload := map[string]any{"values": values, "missing": missing}
	if len(values) == 1 {
		payload["value"] = values[0].Value
		payload["type"] = values[0].Type
	}
	return Ok(payload, "")
}

// scalarText renders a single value for terminal output.
func scalarText(vtype string, value any) string {
	switch v := value.(type) {
	case string:
		return v
	case nil:
		return "null"
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, scalarText(vtype, item))
		}
		return strings.Join(parts, ", ")
	case []int:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%d", item))
		}
		return strings.Join(parts, ", ")
	case []float64:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%g", item))
		}
		return strings.Join(parts, ", ")
	case []bool:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%t", item))
		}
		return strings.Join(parts, ", ")
	case []string:
		return strings.Join(v, ", ")
	case map[string]any:
		return renderTree(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// renderTree prints a nested value with indentation.
func renderTree(tree map[string]any) string {
	keys := make([]string, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		value := tree[k]
		switch v := value.(type) {
		case map[string]any:
			fmt.Fprintf(&b, "%s:\n%s", k, indent(renderTree(v), "  "))
		case []any:
			parts := make([]string, 0, len(v))
			for _, item := range v {
				if m, ok := item.(map[string]any); ok {
					parts = append(parts, renderTree(m))
					continue
				}
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			fmt.Fprintf(&b, "%s:\n%s", k, indent(strings.Join(parts, "\n"), "  "))
		default:
			fmt.Fprintf(&b, "%s: %v", k, v)
		}
	}
	return b.String()
}

func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

// getHandler implements `tycl get`: read a single value, raw or as JSON.
func getHandler(ctx *Ctx, args []string) *Result {
	return queryHandler(ctx, []string{args[0], args[1]})
}

// setHandler implements `tycl set`: write a value into a config file.
func setHandler(ctx *Ctx, args []string) *Result {
	path := args[0]
	in, cfg, res := loadConfig(ctx, path)
	if res != nil {
		return res
	}
	valuePath := args[1]
	vtype := args[2]
	raw := args[3]
	if vtype == "auto" {
		vtype = lang.InferType(raw)
	}
	if !lang.IsValidType(vtype) {
		return Usagef(
			"unknown type %q: use one of %s, or 'auto'",
			vtype, strings.Join(lang.ValidTypes(), ", "))
	}
	if err := lang.Set(cfg, valuePath, vtype, raw); err != nil {
		return Failf(diag.CodeValue, "%v", err)
	}
	return writeConfig(ctx, cfg, in.Name, fmt.Sprintf("set %s to %s", valuePath, vtype))
}

// removeHandler implements `tycl remove`: delete a key by path.
func removeHandler(ctx *Ctx, args []string) *Result {
	path := args[0]
	in, cfg, res := loadConfig(ctx, path)
	if res != nil {
		return res
	}
	key := args[1]
	parent, leaf := splitPath(cfg, key)
	if parent == nil {
		return Failf(diag.CodeValue, "path %q not found in %s", key, in.Name)
	}
	if !lang.Remove(parent, leaf) {
		return Failf(diag.CodeValue, "key %q not found in %s", key, in.Name)
	}
	return writeConfig(ctx, cfg, in.Name, "removed "+key)
}

// splitPath resolves a dot-path to the parent object and the final key.
func splitPath(cfg *shared.Config, path string) (*shared.Config, string) {
	segments := strings.Split(path, ".")
	current := cfg
	for idx := 0; idx < len(segments)-1; idx++ {
		key := segments[idx]
		if child, ok := current.InnerV[key]; ok {
			current = child
			continue
		}
		items, ok := current.InnerArrV[key]
		if !ok {
			return nil, ""
		}
		position, err := strconv.Atoi(segments[idx+1])
		if err != nil {
			return nil, ""
		}
		if position < 0 {
			position += len(items)
		}
		if position < 0 || position >= len(items) {
			return nil, ""
		}
		current = items[position]
		idx++
	}
	return current, segments[len(segments)-1]
}

// loadConfig reads a config file and parses it, reporting errors as diagnostics.
func loadConfig(ctx *Ctx, path string) (*Input, *shared.Config, *Result) {
	in, err := ReadInput(path)
	if err != nil {
		return nil, nil, failFromError(err)
	}
	ctx.track(in)
	cfg, diags := validateInput(ctx, in, "", "")
	if len(diags) > 0 {
		return nil, nil, Fail(diags...)
	}
	return in, cfg, nil
}

// writeConfig serialises a config back to TYCL and stores it.
func writeConfig(ctx *Ctx, cfg *shared.Config, name, message string) *Result {
	code, err := generation.Tycl(cfg)
	if err != nil {
		return failFromError(err)
	}
	if err := WriteOutput(name, code); err != nil {
		return failFromError(err)
	}
	return Ok(nil, message)
}

// lookupFailure builds diagnostics for paths that could not be resolved.
func lookupFailure(name string, missing []string, cfg *shared.Config) []diag.Diagnostic {
	out := make([]diag.Diagnostic, 0, len(missing))
	for _, path := range missing {
		hint := "run 'tycl query " + name + " --json' to list the existing paths"
		if near := closestPath(cfg, path); near != "" {
			hint = "did you mean " + near + "?"
		}
		out = append(out, diag.Diagnostic{
			Code:     diag.CodeValue,
			Severity: diag.SeverityError,
			Message:  fmt.Sprintf("path %q not found", path),
			Path:     path,
			Hint:     hint,
		})
	}
	return out
}

// closestPath suggests a known path when the requested one is a typo.
func closestPath(cfg *shared.Config, path string) string {
	known := ListPaths(cfg, "")
	best := ""
	bestScore := 1 << 30
	for _, candidate := range known {
		score := editDistance(strings.ToLower(path), strings.ToLower(candidate))
		if score < bestScore {
			bestScore = score
			best = candidate
		}
	}
	if bestScore > 3 {
		return ""
	}
	return best
}

// editDistance is the Levenshtein distance used for path suggestions.
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
			curr[j] = minOf(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minOf(a, b, c int) int { return min(min(a, b), c) }

// plainValue converts typed values into plain Go values for JSON encoding.
func plainValue(entry *lang.Entry) any {
	switch v := entry.Value.(type) {
	case *shared.Config:
		return structureOf(v)
	case []*shared.Config:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, structureOf(item))
		}
		return out
	}
	return entry.Value
}

// ListPaths returns every readable path of a config, sorted.
func ListPaths(cfg *shared.Config, prefix string) []string {
	if cfg == nil {
		return nil
	}
	out := []string{}
	for _, group := range []struct {
		keys []string
	}{
		{sortedKeys(cfg.IntV)},
		{sortedKeys(cfg.FloatV)},
		{sortedKeys(cfg.BoolV)},
		{sortedKeys(cfg.StringV)},
		{sortedKeys(cfg.NullV)},
		{sortedKeys(cfg.IntArrV)},
		{sortedKeys(cfg.FloatArrV)},
		{sortedKeys(cfg.BoolArrV)},
		{sortedKeys(cfg.StringArrV)},
	} {
		for _, key := range group.keys {
			out = append(out, join(prefix, key))
		}
	}

	nested := sortedKeys(cfg.InnerV)
	for _, key := range nested {
		path := join(prefix, key)
		out = append(out, path)
		out = append(out, ListPaths(cfg.InnerV[key], path)...)
	}

	arrays := sortedKeys(cfg.InnerArrV)
	for _, key := range arrays {
		path := join(prefix, key)
		out = append(out, path)
		for idx, item := range cfg.InnerArrV[key] {
			itemPath := fmt.Sprintf("%s.%d", path, idx)
			out = append(out, itemPath)
			out = append(out, ListPaths(item, itemPath)...)
		}
	}

	sort.Strings(out)
	return out
}

func join(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// structureOf converts a config into a plain map for JSON output.
func structureOf(cfg *shared.Config) map[string]any {
	if cfg == nil {
		return map[string]any{}
	}
	res := map[string]any{}
	for k, v := range cfg.IntV {
		res[k] = v
	}
	for k, v := range cfg.FloatV {
		res[k] = v
	}
	for k, v := range cfg.BoolV {
		res[k] = v
	}
	for k, v := range cfg.StringV {
		res[k] = v
	}
	for k := range cfg.NullV {
		res[k] = nil
	}
	for k, v := range cfg.IntArrV {
		res[k] = v
	}
	for k, v := range cfg.FloatArrV {
		res[k] = v
	}
	for k, v := range cfg.BoolArrV {
		res[k] = v
	}
	for k, v := range cfg.StringArrV {
		res[k] = v
	}
	for k, v := range cfg.InnerV {
		res[k] = structureOf(v)
	}
	for k, v := range cfg.InnerArrV {
		items := make([]any, 0, len(v))
		for _, item := range v {
			items = append(items, structureOf(item))
		}
		res[k] = items
	}
	return res
}
