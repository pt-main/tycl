package generation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/tycl/format"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

func Tycl(conf *shared.Config) (string, error) {
	if conf == nil {
		return "{}", nil
	}
	raw, err := tyclRaw(conf)
	if err != nil {
		return "", err
	}

	return format.FormConfig(raw)
}

func tyclRaw(conf *shared.Config) (string, error) {
	if conf == nil {
		return "{}", nil
	}

	var b strings.Builder
	b.WriteString("{\n")

	addLine := func(line string) {
		b.WriteString(line)
		b.WriteString(",\n")
	}

	formatValue := func(vtype string, value interface{}) (string, error) {
		switch vtype {
		case "int":
			if v, ok := value.(int); ok {
				return strconv.Itoa(v), nil
			}
			return "", fmt.Errorf("invalid int value: %v", value)
		case "float":
			if v, ok := value.(float64); ok {
				return strconv.FormatFloat(v, 'f', -1, 64), nil
			}
			return "", fmt.Errorf("invalid float value: %v", value)
		case "bool":
			if v, ok := value.(bool); ok {
				if v {
					return "true", nil
				}
				return "false", nil
			}
			return "", fmt.Errorf("invalid bool value: %v", value)
		case "string":
			if v, ok := value.(string); ok {
				return lang.ReprStringValue(v), nil
			}
			return "", fmt.Errorf("invalid string value: %v", value)
		case "null":
			return "null", nil
		default:
			return "", fmt.Errorf("unsupported type: %s", vtype)
		}
	}

	for k, v := range conf.IntV {
		valStr, err := formatValue("int", v)
		if err != nil {
			return "", fmt.Errorf("key %q: %w", k, err)
		}
		addLine(fmt.Sprintf("%s: int = %s", k, valStr))
	}
	for k, v := range conf.FloatV {
		valStr, err := formatValue("float", v)
		if err != nil {
			return "", fmt.Errorf("key %q: %w", k, err)
		}
		addLine(fmt.Sprintf("%s: float = %s", k, valStr))
	}
	for k, v := range conf.BoolV {
		valStr, err := formatValue("bool", v)
		if err != nil {
			return "", fmt.Errorf("key %q: %w", k, err)
		}
		addLine(fmt.Sprintf("%s: bool = %s", k, valStr))
	}
	for k, v := range conf.StringV {
		valStr, err := formatValue("string", v)
		if err != nil {
			return "", fmt.Errorf("key %q: %w", k, err)
		}
		addLine(fmt.Sprintf("%s: string = %s", k, valStr))
	}
	for k, typ := range conf.NullV {
		addLine(fmt.Sprintf("%s: %s = null", k, typ))
	}

	for k, v := range conf.IntArrV {
		elements := make([]string, len(v))
		for i, val := range v {
			elements[i] = strconv.Itoa(val)
		}
		addLine(fmt.Sprintf("%s: ints = [%s]", k, strings.Join(elements, ", ")))
	}
	for k, v := range conf.FloatArrV {
		elements := make([]string, len(v))
		for i, val := range v {
			elements[i] = strconv.FormatFloat(val, 'f', -1, 64)
		}
		addLine(fmt.Sprintf("%s: floats = [%s]", k, strings.Join(elements, ", ")))
	}
	for k, v := range conf.BoolArrV {
		elements := make([]string, len(v))
		for i, val := range v {
			if val {
				elements[i] = "true"
			} else {
				elements[i] = "false"
			}
		}
		addLine(fmt.Sprintf("%s: bools = [%s]", k, strings.Join(elements, ", ")))
	}
	for k, v := range conf.StringArrV {
		elements := make([]string, len(v))
		for i, val := range v {
			escaped := strings.ReplaceAll(val, `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, `"`, `\"`)
			elements[i] = `"` + escaped + `"`
		}
		addLine(fmt.Sprintf("%s: strings = [%s]", k, strings.Join(elements, ", ")))
	}

	for k, sub := range conf.InnerV {
		subRaw, err := tyclRaw(sub)
		if err != nil {
			return "", fmt.Errorf("object %q: %w", k, err)
		}

		addLine(fmt.Sprintf("%s: object = %s", k, subRaw))
	}

	for k, arr := range conf.InnerArrV {
		if len(arr) == 0 {
			addLine(fmt.Sprintf("%s: objects = []", k))
			continue
		}
		elements := make([]string, len(arr))
		for i, sub := range arr {
			subRaw, err := tyclRaw(sub)
			if err != nil {
				return "", fmt.Errorf("object array %q index %d: %w", k, i, err)
			}
			elements[i] = subRaw
		}

		addLine(fmt.Sprintf("%s: objects = [%s]", k, strings.Join(elements, ", ")))
	}

	b.WriteString("}")
	return b.String(), nil
}
