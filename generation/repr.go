package generation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/tycl/shared"
)

// GetRepr возвращает строковое представление значения из Config по типу и ключу
// в формате TYCL. Поддерживаются все типы: скаляры, объекты, массивы.
// Если ключ не найден или тип не совпадает, возвращает ошибку.
func GetRepr(cfg *shared.Config, typ, name string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
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

	// Массивы скаляров
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

	// Объект
	case "object":
		if v, ok := cfg.InnerV[name]; ok {
			return reprObject(v), nil
		}

	// Массив объектов
	case "objects":
		if v, ok := cfg.InnerArrV[name]; ok {
			return reprObjectArray(v), nil
		}

	default:
		return "", fmt.Errorf("unsupported type: %s", typ)
	}

	return "", fmt.Errorf("key %q not found for type %s", name, typ)
}

// Вспомогательные функции для форматирования в TYCL-стиле

func quoteString(s string) string {
	// Используем двойные кавычки и экранируем, как в TYCL
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
	// Используем существующую функцию Tycl для генерации TYCL-кода объекта
	// (она уже есть в пакете generation)
	code, err := Tycl(obj)
	if err != nil {
		// fallback: использовать упрощённое представление
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
