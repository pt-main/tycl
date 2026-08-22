package lang

import (
	"os"
	"strconv"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

type actionMap map[string]func(
	cp *configParser, pn *stringParsing.ParsedNode, args []stringParsing.ParsedNode,
) (vtype string, val string, err core.ErrorInterface)

func Actions() map[string]func(cp *configParser, pn *stringParsing.ParsedNode, args []stringParsing.ParsedNode) (
	vtype, val string, err core.ErrorInterface,
) {
	return actionMap{
		"file": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype, val string, err core.ErrorInterface) {
			if len(args) != 1 {
				err = core.Err(shared.RuntimeError, "Invalid arguments length: must be 1 (now %v)", len(args))
				return
			}
			sw := args[0].Switch
			if sw != "STRING" {
				err = core.Err(shared.RuntimeError, "Invalid argument: need string, got %v", sw)
				return
			}
			var strv string
			strv, err = parseStringValue(args[0].Raw)
			if err != nil {
				return
			}
			var err_ error
			val, err_ = utils.OpenF(strv)
			if err_ != nil {
				err = core.Wrap(shared.WrappedError, err_, err_.Error())
				return
			}
			val = utils.ReprStringValue(val)
			vtype = "string"
			return
		},
		"asObject": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
			var a0t string
			a0t, _, _, _, val, _, err = cp.parseType(&args[0], "string", args[0].Raw)
			if a0t != "string" {
				err = core.Err(shared.RuntimeError, "Invalid argument: need string, got %v", a0t)
				return
			}
			vtype = "object"
			return
		},
		"asString": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
			if len(args) != 1 {
				err = core.Err(shared.RuntimeError, "Invalid arguments length: must be 1 (now %v)", len(args))
				return
			}
			val = args[0].Raw
			if args[0].Switch == "action" {
				_, val, err = cp.parseAction(&args[0])
			}
			val = utils.ReprStringValue(val)
			vtype = "string"
			return
		},
		"join": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
			join := ""
			for _, arg := range args {
				var aT string
				var strv string
				aT, _, _, _, strv, _, err = cp.parseType(&arg, "string", arg.Raw)
				if aT != "string" {
					err = core.Err(shared.RuntimeError, "Invalid argument: need string, got %v", aT)
					return
				}
				join += strv
			}
			vtype = "string"
			val = utils.ReprStringValue(join)
			return
		},
		"env": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
			if len(args) != 3 {
				err = core.Err(shared.RuntimeError, "env: need 3 args, got %d", len(args))
				return
			}

			var name string
			name, err = parseStringValue(args[0].Raw)
			if err != nil {
				return
			}

			var defaultStr string
			defaultStr, err = parseStringValue(args[1].Raw)
			if err != nil {
				return
			}

			vtype, err = parseStringValue(args[2].Raw)
			if err != nil {
				return
			}

			envVal, exists := os.LookupEnv(name)
			if !exists || envVal == "" {
				envVal = defaultStr
			}

			val = envVal
			return
		},
		"get": func(cp *configParser, pn *stringParsing.ParsedNode, args []stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
			if len(args) != 2 {
				err = core.Err(shared.RuntimeError, "Get: need 2 args (path, type), got %d", len(args))
				return
			}

			pathNode := args[0]
			if pathNode.Switch != "STRING" {
				err = core.Err(shared.RuntimeError, "Get: first arg must be string (path)")
				return
			}
			path, err_ := parseStringValue(pathNode.Raw)
			if err_ != nil {
				err = core.Wrap(shared.WrappedError, err_, "Get: invalid path")
				return
			}

			typeNode := args[1]
			if typeNode.Switch != "STRING" {
				err = core.Err(shared.RuntimeError, "Get: second arg must be string (type)")
				return
			}
			expectedType, err_ := parseStringValue(typeNode.Raw)
			if err_ != nil {
				err = core.Wrap(shared.WrappedError, err_, "Get: invalid type")
				return
			}
			expectedType = strings.ToLower(expectedType)
			if !utils.IsTypeValid(expectedType) && expectedType != "object" {
				err = core.Err(shared.RuntimeError, "Get: unsupported type %s", expectedType)
				return
			}

			obj := cp.Conf.MainConf
			segments := strings.Split(path, ".")
			if len(segments) == 0 {
				err = core.Err(shared.RuntimeError, "Get: empty path")
				return
			}

			lastSeg := segments[len(segments)-1]
			isIndex := false
			var index int
			if i, err_ := strconv.Atoi(lastSeg); err_ == nil {
				isIndex = true
				index = i
			}

			pathSegs := segments
			if isIndex {
				pathSegs = segments[:len(segments)-1]
			}

			for _, seg := range pathSegs {
				var ok bool
				obj, ok = obj.InnerV[seg]
				if !ok {
					err = core.Err(shared.RuntimeError, "Get: path %q not found", path)
					return
				}
			}

			if isIndex {

				if len(segments) < 2 {
					err = core.Err(shared.RuntimeError, "Get: need array name before index")
					return
				}
				arrayName := segments[len(segments)-2]
				var found bool
				var resultStr string

				switch expectedType {
				case "int":
					arr, ok := obj.IntArrV[arrayName]
					if !ok {
						break
					}
					found = true
					if index < 0 {
						index = len(arr) + index
					}
					if index < 0 || index >= len(arr) {
						err = core.Err(shared.RuntimeError, "Get: index %d out of range for array %s", index, arrayName)
						return
					}
					resultStr = strconv.Itoa(arr[index])
					vtype = "int"
				case "float":
					arr, ok := obj.FloatArrV[arrayName]
					if !ok {
						break
					}
					found = true
					if index < 0 {
						index = len(arr) + index
					}
					if index < 0 || index >= len(arr) {
						err = core.Err(shared.RuntimeError, "Get: index %d out of range for array %s", index, arrayName)
						return
					}
					resultStr = strconv.FormatFloat(arr[index], 'f', -1, 64)
					vtype = "float"
				case "bool":
					arr, ok := obj.BoolArrV[arrayName]
					if !ok {
						break
					}
					found = true
					if index < 0 {
						index = len(arr) + index
					}
					if index < 0 || index >= len(arr) {
						err = core.Err(shared.RuntimeError, "Get: index %d out of range for array %s", index, arrayName)
						return
					}
					resultStr = strconv.FormatBool(arr[index])
					vtype = "bool"
				case "string":
					arr, ok := obj.StringArrV[arrayName]
					if !ok {
						break
					}
					found = true
					if index < 0 {
						index = len(arr) + index
					}
					if index < 0 || index >= len(arr) {
						err = core.Err(shared.RuntimeError, "Get: index %d out of range for array %s", index, arrayName)
						return
					}
					resultStr = arr[index]
					vtype = "string"
				case "object":
					arr, ok := obj.InnerArrV[arrayName]
					if !ok {
						break
					}
					found = true
					if index < 0 {
						index = len(arr) + index
					}
					if index < 0 || index >= len(arr) {
						err = core.Err(shared.RuntimeError, "Get: index %d out of range for array %s", index, arrayName)
						return
					}
					objCode, err_ := generation.Tycl(arr[index])
					if err_ != nil {
						err = core.Wrap(shared.WrappedError, err_, "Get: cannot generate object representation")
						return
					}
					resultStr = objCode
					vtype = "object"
				default:
					err = core.Err(shared.RuntimeError, "Get: unsupported type %s", expectedType)
					return
				}

				if !found {
					err = core.Err(shared.RuntimeError, "Get: array %q not found in %q", arrayName, path)
					return
				}
				val = resultStr
				return
			} else {
				key := lastSeg
				var found bool
				var resultStr string

				switch expectedType {
				case "int":
					if v, ok := obj.IntV[key]; ok {
						found = true
						resultStr = strconv.Itoa(v)
						vtype = "int"
					}
				case "float":
					if v, ok := obj.FloatV[key]; ok {
						found = true
						resultStr = strconv.FormatFloat(v, 'f', -1, 64)
						vtype = "float"
					}
				case "bool":
					if v, ok := obj.BoolV[key]; ok {
						found = true
						resultStr = strconv.FormatBool(v)
						vtype = "bool"
					}
				case "string":
					if v, ok := obj.StringV[key]; ok {
						found = true
						resultStr = v
						vtype = "string"
					}
				case "object":
					if v, ok := obj.InnerV[key]; ok {
						found = true
						objCode, err_ := generation.Tycl(v)
						if err_ != nil {
							err = core.Wrap(shared.WrappedError, err_, "Get: cannot generate object representation")
							return
						}
						resultStr = objCode
						vtype = "object"
					}
				default:
					err = core.Err(shared.RuntimeError, "Get: unsupported type %s", expectedType)
					return
				}

				if !found {
					err = core.Err(shared.RuntimeError, "Get: key %q not found in path %q", key, path)
					return
				}
				val = resultStr
				return
			}
		},
	}

}
