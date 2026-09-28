package lang

import (
	"slices"
	"strconv"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/tycl/shared"
)

func (cp *configParser) parseType(node *stringParsing.ParsedNode, vtype, value string) (
	restype string, boolv bool, intv int, floatv float64, stringv string, objectv *shared.Config, err core.ErrorInterface,
) {
	if vtype == "action" || node.Switch == "action" {
		vtype, value, err = cp.parseAction(node)
		if err != nil {
			return
		}
	}
	vtype = strings.ToLower(vtype)
	restype = vtype
	var cause error
	switch vtype {
	case "string":
		stringv, cause = parseStringValue(value)
		if cause == nil {
			return
		}
	case "int":
		intv, cause = strconv.Atoi(value)
		if cause == nil {
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
		floatv, cause = strconv.ParseFloat(value, 64)
		if cause == nil {
			return
		}
	case "object":
		var obj *shared.Config
		newconf := shared.NewNilConfig()
		newconf.MainConf = cp.Conf
		obj, err = ParseConfAt(newconf, value, cp.StrictKeys, cp.offsetOf(node))
		if err == nil {
			obj.Name = "inner"
			objectv = obj
			return
		}
	}
	if cause != nil {
		err = core.Wrap(shared.WrappedError, cause, "Parsing: %v", cause.Error())
	}
	err = core.Err(shared.RuntimeError, "Invalid value for %v: %v", vtype, describeValue(value)).
		WithMeta("hint", valueSuggestion(vtype, value))
	return
}

// describeValue renders a value for an error message without dumping a
// multi-line string into the terminal.
func describeValue(value string) string {
	if strings.ContainsAny(value, "\n") {
		return "a multi-line literal"
	}
	return "'" + value + "'"
}

func parseStringValue(value string) (stringv string, err core.ErrorInterface) {
	last := len(value) - 1
	if len(value) >= 2 && value[0] == '\'' && value[last] == '\'' {
		value = `"` + value[1:last] + `"`
		value = strings.ReplaceAll(value, "\\'", "'")
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		var cause error
		value, cause = strconv.Unquote(value)
		if cause == nil {
			stringv = value
			return
		}
		err = core.Wrap(shared.WrappedError, cause, "%v", cause)
		return
	}
	err = core.Err(shared.RuntimeError, "Invalid string format: a value must be wrapped in quotes")
	return
}
