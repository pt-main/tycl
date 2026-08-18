package generation

import (
	"strconv"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/tycl/shared"
)

func GetRepr(cfg *shared.Config, typ, name string) (string, core.ErrorInterface) {
	if cfg == nil {
		return "", core.Err(shared.RuntimeError, "Config is nil")
	}

	switch typ {
	case "int":
		if v, ok := cfg.IntV[name]; ok {
			return strconv.Itoa(v), nil
		}
	case "float":
		if v, ok := cfg.FloatV[name]; ok {
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		}
	case "bool":
		if v, ok := cfg.BoolV[name]; ok {
			return strconv.FormatBool(v), nil
		}
	case "string":
		if v, ok := cfg.StringV[name]; ok {
			return quoteString(v), nil
		}
	case "null":
		if _, ok := cfg.NullV[name]; ok {
			return "null", nil
		}

	case "ints":
		if v, ok := cfg.IntArrV[name]; ok {
			return reprIntArray(v), nil
		}
	case "floats":
		if v, ok := cfg.FloatArrV[name]; ok {
			return reprFloatArray(v), nil
		}
	case "bools":
		if v, ok := cfg.BoolArrV[name]; ok {
			return reprBoolArray(v), nil
		}
	case "strings":
		if v, ok := cfg.StringArrV[name]; ok {
			return reprStringArray(v), nil
		}

	case "object":
		if v, ok := cfg.InnerV[name]; ok {
			return reprObject(v), nil
		}

	case "objects":
		if v, ok := cfg.InnerArrV[name]; ok {
			return reprObjectArray(v), nil
		}

	default:
		return "", core.Err(shared.RuntimeError, "Unsupported type: %s", typ)
	}

	return "", core.Err(shared.RuntimeError, "Key %q not found for type %s", name, typ)
}

func quoteString(s string) string {
	return strconv.Quote(s)
}

func reprIntArray(arr []int) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = strconv.Itoa(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func reprFloatArray(arr []float64) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func reprBoolArray(arr []bool) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = strconv.FormatBool(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func reprStringArray(arr []string) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = quoteString(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func reprObject(obj *shared.Config) string {
	if obj == nil {
		return "{}"
	}
	code, err := Tycl(obj)
	if err != nil {

		return "{}"
	}
	return code
}

func reprObjectArray(arr []*shared.Config) string {
	if len(arr) == 0 {
		return "[]"
	}
	parts := make([]string, len(arr))
	for i, obj := range arr {
		parts[i] = reprObject(obj)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
