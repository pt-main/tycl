package lang

import (
	"strconv"
	"strings"

	"github.com/pt-main/tycl/shared"
)

// Entry is one value found at a config path.
type Entry struct {
	Type  string
	Value any
	Path  string
}

// Resolve walks a dot-path through nested objects and returns the value.
//
// It understands arrays with a numeric index, so "servers.0.host" works.
func Resolve(cfg *shared.Config, path string) (*Entry, bool) {
	if cfg == nil || path == "" {
		return nil, false
	}
	segments := strings.Split(path, ".")
	current := cfg

	for idx := 0; idx < len(segments); idx++ {
		key := segments[idx]
		isLast := idx == len(segments)-1

		if isLast {
			if e, ok := direct(current, key, path); ok {
				return e, true
			}
			return nil, false
		}

		if child, ok := current.InnerV[key]; ok {
			current = child
			continue
		}
		if items, ok := current.InnerArrV[key]; ok {
			next := segments[idx+1]
			position, err := strconv.Atoi(next)
			if err != nil {
				return nil, false
			}
			if position < 0 {
				position += len(items)
			}
			if position < 0 || position >= len(items) {
				return nil, false
			}
			current = items[position]
			idx++
			continue
		}
		return nil, false
	}
	return nil, false
}

// direct reads a key of the current object without descending further.
func direct(cfg *shared.Config, key, path string) (*Entry, bool) {
	if cfg == nil {
		return nil, false
	}
	if v, ok := cfg.NullV[key]; ok {
		return &Entry{Type: v, Value: nil, Path: path}, true
	}
	if v, ok := cfg.IntV[key]; ok {
		return &Entry{Type: "int", Value: v, Path: path}, true
	}
	if v, ok := cfg.FloatV[key]; ok {
		return &Entry{Type: "float", Value: v, Path: path}, true
	}
	if v, ok := cfg.BoolV[key]; ok {
		return &Entry{Type: "bool", Value: v, Path: path}, true
	}
	if v, ok := cfg.StringV[key]; ok {
		return &Entry{Type: "string", Value: v, Path: path}, true
	}
	if v, ok := cfg.IntArrV[key]; ok {
		return &Entry{Type: "ints", Value: v, Path: path}, true
	}
	if v, ok := cfg.FloatArrV[key]; ok {
		return &Entry{Type: "floats", Value: v, Path: path}, true
	}
	if v, ok := cfg.BoolArrV[key]; ok {
		return &Entry{Type: "bools", Value: v, Path: path}, true
	}
	if v, ok := cfg.StringArrV[key]; ok {
		return &Entry{Type: "strings", Value: v, Path: path}, true
	}
	if v, ok := cfg.InnerV[key]; ok {
		return &Entry{Type: "object", Value: v, Path: path}, true
	}
	if v, ok := cfg.InnerArrV[key]; ok {
		return &Entry{Type: "objects", Value: v, Path: path}, true
	}
	return nil, false
}

// Set assigns a value at a dot-path, creating nothing implicitly.
//
// The parent must already exist; only the final key is written. A numeric
// segment steps into an object array, so "servers.0.port" works.
func Set(cfg *shared.Config, path, vtype, raw string) error {
	segments := strings.Split(path, ".")
	if len(segments) == 0 {
		return errPath
	}
	current := cfg
	for idx := 0; idx < len(segments)-1; idx++ {
		key := segments[idx]
		if child, ok := current.InnerV[key]; ok {
			current = child
			continue
		}
		items, ok := current.InnerArrV[key]
		if !ok {
			return errPath
		}
		position, err := strconv.Atoi(segments[idx+1])
		if err != nil {
			return errPath
		}
		if position < 0 {
			position += len(items)
		}
		if position < 0 || position >= len(items) {
			return errPath
		}
		current = items[position]
		idx++
	}
	return setTyped(current, segments[len(segments)-1], vtype, raw)
}
