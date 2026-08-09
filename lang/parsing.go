package lang

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/tycl/shared"
)

func (cp *configParser) parseType(node *stringParsing.ParsedNode, vtype, value string) (
	restype string, boolv bool, intv int, floatv float64, stringv string, objectv *shared.Config, err error,
) {
	if vtype == "action" || node.Switch == "action" {
		vtype, value, err = cp.parseAction(node)
		if err != nil {
			return
		}
	}
	vtype = strings.ToLower(vtype)
	restype = vtype
	switch vtype {
	case "string":
		stringv, err = parseStringValue(value)
		if err == nil {
			return
		}
	case "int":
		intv, err = strconv.Atoi(value)
		if err == nil {
			return
		}
	case "bool":
		if slices.Contains([]string{"true", "false"}, value) {
			boolv = false
			if value == "true" {
				boolv = true
			}
			return
		}
	case "float":
		floatv, err = strconv.ParseFloat(value, 64)
		if err == nil {
			return
		}
	case "object":
		var obj *shared.Config
		obj, err = ParseConf(value, cp.StrictKeys)
		if err == nil {
			obj.Name = "inner"
			objectv = obj
			return
		}
	}
	err = fmt.Errorf("Invalid value (type of %v): '%v'. %v", vtype, value, err)
	return
}

func parseStringValue(value string) (stringv string, err error) {
	last := len(value) - 1
	if len(value) >= 2 && value[0] == '\'' && value[last] == '\'' {
		value = `"` + value[1:last] + `"`
		value = strings.ReplaceAll(value, "\\'", "'")
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value, err = strconv.Unquote(value)
		if err == nil {
			stringv = value
			return
		}
	}
	err = fmt.Errorf("Invalid string format")
	return
}

func ReprStringValue(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `"`, `\"`), "\n", `\n`) + `"`
}
