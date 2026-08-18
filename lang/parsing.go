package lang

import (
	"slices"
	"strconv"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/lc/parsing/stringParsing"
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
	var err_ error
	switch vtype {
	case "string":
		stringv, err_ = parseStringValue(value)
		if err == nil {
			return
		}
	case "int":
		intv, err_ = strconv.Atoi(value)
		if err_ == nil {
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
		floatv, err_ = strconv.ParseFloat(value, 64)
		if err_ == nil {
			return
		}
	case "object":
		var obj *shared.Config
		newconf := shared.NewNilConfig()
		newconf.MainConf = cp.Conf
		obj, err = ParseConf(newconf, value, cp.StrictKeys)
		if err == nil {
			obj.Name = "inner"
			objectv = obj
			return
		}
	}
	if err_ != nil {
		err = core.Wrap(shared.WrappedError, err_, "Parsing: %v", err_.Error())
	}
	err = core.Wrap(shared.RuntimeError, err, "Invalid value (type of %v): '%v'", vtype, value)
	return
}

func parseStringValue(value string) (stringv string, err core.ErrorInterface) {
	last := len(value) - 1
	if len(value) >= 2 && value[0] == '\'' && value[last] == '\'' {
		value = `"` + value[1:last] + `"`
		value = strings.ReplaceAll(value, "\\'", "'")
	}
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		var err_ error
		value, err_ = strconv.Unquote(value)
		if err_ == nil {
			stringv = value
			return
		}
		err = core.Wrap(shared.WrappedError, err_, err_.Error())
	}
	err = core.Err(shared.RuntimeError, "Invalid string format")
	return
}
