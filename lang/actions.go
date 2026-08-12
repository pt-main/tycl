package lang

import (
	"fmt"
	"os"
	"strings"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/utils"
)

type actionMap map[string]func(
	cp *configParser, pn *stringParsing.ParsedNode, args []stringParsing.ParsedNode,
) (vtype string, val string, err error)

func Actions() map[string]func(cp *configParser, pn *stringParsing.ParsedNode, args []stringParsing.ParsedNode) (
	vtype, val string, err error,
) {
	return actionMap{
		"file": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype, val string, err error) {
			if len(args) != 1 {
				err = fmt.Errorf("Invalid arguments length: must be 1 (now %v)", len(args))
				return
			}
			sw := args[0].Switch
			if sw != "STRING" {
				err = fmt.Errorf("Invalid argument: need string, got %v", sw)
				return
			}
			var strv string
			strv, err = parseStringValue(args[0].Raw)
			if err != nil {
				return
			}
			val, err = utils.OpenF(strv)
			if err != nil {
				return
			}
			val = utils.ReprStringValue(val)
			vtype = "string"
			return
		},
		"asObject": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err error) {
			var a0t string
			a0t, _, _, _, val, _, err = cp.parseType(&args[0], "string", args[0].Raw)
			if a0t != "string" {
				err = fmt.Errorf("Invalid argument: need string, got %v", a0t)
				return
			}
			vtype = "object"
			return
		},
		"asString": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err error) {
			if len(args) != 1 {
				err = fmt.Errorf("Invalid arguments length: must be 1 (now %v)", len(args))
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
			args []stringParsing.ParsedNode) (vtype string, val string, err error) {
			join := ""
			for _, arg := range args {
				var aT string
				var strv string
				aT, _, _, _, strv, _, err = cp.parseType(&arg, "string", arg.Raw)
				if aT != "string" {
					err = fmt.Errorf("Invalid argument: need string, got %v", aT)
					return
				}
				join += strv
			}
			vtype = "string"
			val = utils.ReprStringValue(join)
			return
		},
		"env": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err error) {
			if len(args) != 3 {
				err = fmt.Errorf("env: need 3 args, got %d", len(args))
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
		"get": func(cp *configParser, pn *stringParsing.ParsedNode,
			args []stringParsing.ParsedNode) (vtype string, val string, err error) {
			if len(args) != 2 {
				err = fmt.Errorf("got: need 2 args, got %d", len(args))
				return
			}

			var name string
			name, err = parseStringValue(args[0].Raw)
			if err != nil {
				return
			}

			vtype, err = parseStringValue(args[1].Raw)
			if err != nil {
				return
			}

			obj := cp.Conf.MainConf
			getName := ""
			path := strings.Split(name, ".")
			for idx, to := range path {
				if idx != len(path)-1 {
					var ok bool
					obj, ok = obj.InnerV[to]
					if !ok {
						err = fmt.Errorf("Can't get: %v", to)
						return
					}
				} else {
					getName = to
				}
			}

			val, err = generation.GetRepr(obj, vtype, getName)
			if err != nil {
				return
			}

			return
		},
	}
}
