package lang

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/tycl/shared"
)

var errPath = errors.New("path not found")

// ErrType reports a value that does not match the requested type.
var ErrType = errors.New("value does not match type")

// ErrUnknownType reports a TYCL type name outside the language.
var ErrUnknownType = errors.New("unknown type")

// setTyped writes a value of the given TYCL type into a config.
//
// The key is first cleared from every other type map, so switching the type
// of an existing key never leaves a stale duplicate behind.
func setTyped(cfg *shared.Config, key, vtype, raw string) error {
	if cfg == nil {
		return errPath
	}
	if !IsValidType(vtype) {
		return fmt.Errorf("%w: %q (known: %s)", ErrUnknownType, vtype, strings.Join(ValidTypes(), ", "))
	}
	clearKey(cfg, key)

	if raw == "null" || raw == "" {
		cfg.NullV[key] = vtype
		return nil
	}

	single := vtype[:len(vtype)-1]
	if len(vtype) > 1 && isArrayType(vtype) {
		items, err := parseList(single, raw)
		if err != nil {
			return err
		}
		switch single {
		case "int":
			cfg.IntArrV[key] = items.([]int)
		case "float":
			cfg.FloatArrV[key] = items.([]float64)
		case "bool":
			cfg.BoolArrV[key] = items.([]bool)
		case "string":
			cfg.StringArrV[key] = items.([]string)
		case "object":
			cfg.InnerArrV[key] = items.([]*shared.Config)
		}
		return nil
	}

	switch vtype {
	case "int":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%w: %q is not an int", ErrType, raw)
		}
		cfg.IntV[key] = v
	case "float":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%w: %q is not a float", ErrType, raw)
		}
		cfg.FloatV[key] = v
	case "bool":
		v, err := parseBool(raw)
		if err != nil {
			return err
		}
		cfg.BoolV[key] = v
	case "string":
		v, err := unquote(raw)
		if err != nil {
			return err
		}
		cfg.StringV[key] = v
	case "object":
		obj := shared.NewNilConfig()
		if _, err := ParseConf(obj, raw, false); err != nil {
			return fmt.Errorf("%w: %v", ErrType, err)
		}
		obj.MainConf = cfg.MainConf
		cfg.InnerV[key] = obj
	}
	return nil
}

func isArrayType(vtype string) bool {
	switch vtype {
	case "ints", "floats", "bools", "strings", "objects":
		return true
	}
	return false
}

// clearKey removes a key from every type map of a config.
func clearKey(cfg *shared.Config, key string) {
	delete(cfg.IntV, key)
	delete(cfg.FloatV, key)
	delete(cfg.BoolV, key)
	delete(cfg.StringV, key)
	delete(cfg.NullV, key)
	delete(cfg.IntArrV, key)
	delete(cfg.FloatArrV, key)
	delete(cfg.BoolArrV, key)
	delete(cfg.StringArrV, key)
	delete(cfg.InnerV, key)
	delete(cfg.InnerArrV, key)
}

// Remove deletes a key of any type.
func Remove(cfg *shared.Config, key string) bool {
	before := countTyped(cfg, key)
	if before == 0 {
		return false
	}
	clearKey(cfg, key)
	return true
}

func countTyped(cfg *shared.Config, key string) int {
	n := 0
	for _, found := range []bool{
		has(cfg.IntV, key), has(cfg.FloatV, key), has(cfg.BoolV, key),
		has(cfg.StringV, key), has(cfg.NullV, key), has(cfg.IntArrV, key),
		has(cfg.FloatArrV, key), has(cfg.BoolArrV, key), has(cfg.StringArrV, key),
		has(cfg.InnerV, key), has(cfg.InnerArrV, key),
	} {
		if found {
			n++
		}
	}
	return n
}

func has[V any](m map[string]V, key string) bool {
	_, ok := m[key]
	return ok
}

// parseList parses a comma-separated list of values of a single type.
func parseList(single, raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	parts := splitTopLevel(raw)
	switch single {
	case "int":
		out := make([]int, 0, len(parts))
		for _, p := range parts {
			v, err := strconv.Atoi(p)
			if err != nil {
				return nil, fmt.Errorf("%w: %q is not an int", ErrType, p)
			}
			out = append(out, v)
		}
		return out, nil
	case "float":
		out := make([]float64, 0, len(parts))
		for _, p := range parts {
			v, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: %q is not a float", ErrType, p)
			}
			out = append(out, v)
		}
		return out, nil
	case "bool":
		out := make([]bool, 0, len(parts))
		for _, p := range parts {
			v, err := parseBool(p)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case "string":
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			v, err := unquote(p)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case "object":
		out := make([]*shared.Config, 0, len(parts))
		for _, p := range parts {
			obj := shared.NewNilConfig()
			if _, err := ParseConf(obj, p, false); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrType, err)
			}
			out = append(out, obj)
		}
		return out, nil
	}
	return nil, fmt.Errorf("%w: %q", ErrUnknownType, single)
}

// splitTopLevel splits a comma-separated list, ignoring commas nested
// inside objects, arrays, strings or action calls.
func splitTopLevel(raw string) []string {
	parts := []string{}
	depth := 0
	var quote rune
	start := 0

	runes := []rune(raw)
	for i, r := range runes {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
		case r == '{' || r == '[' || r == '(':
			depth++
		case r == '}' || r == ']' || r == ')':
			depth--
		case r == ',' && depth == 0:
			if part := strings.TrimSpace(string(runes[start:i])); part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	if part := strings.TrimSpace(string(runes[start:])); part != "" {
		parts = append(parts, part)
	}
	return parts
}

func parseBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	}
	return false, fmt.Errorf("%w: %q is not a bool", ErrType, raw)
}

// unquote removes quotes from a string written on the command line.
func unquote(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 {
		if (raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'') {
			unquoted, err := strconv.Unquote(`"` + raw[1:len(raw)-1] + `"`)
			if err == nil {
				return unquoted, nil
			}
			return raw[1 : len(raw)-1], nil
		}
	}
	return raw, nil
}

// InferType picks the TYCL type of a raw command-line value.
func InferType(raw string) string {
	if raw == "null" || raw == "" {
		return "string"
	}
	if _, err := strconv.Atoi(raw); err == nil {
		return "int"
	}
	if _, err := strconv.ParseFloat(raw, 64); err == nil {
		return "float"
	}
	switch strings.ToLower(raw) {
	case "true", "false":
		return "bool"
	}
	return "string"
}
