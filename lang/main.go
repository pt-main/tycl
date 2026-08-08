package lang

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl/format"
	"github.com/pt-main/tycl/lang/lcproc"
	"github.com/pt-main/tycl/shared"
)

func ParseConf(code string) (*shared.Config, error) {
	p := lcproc.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return nil, err
	}
	return ParseBody(&pn[0])
}

func parseType(vtype, value string) (
	boolv bool, intv int, floatv float64, stringv string, objectv *shared.Config, err error,
) {
	switch vtype {
	case "string":
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
		} else {
			err = fmt.Errorf("Invalid string format")
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
		obj, err = ParseConf(value)
		if err == nil {
			obj.Name = "inner"
			objectv = obj
			return
		}
	}
	err = fmt.Errorf("Invalid value (type of %v): %v", vtype, value)
	return
}

func setValue(vtype, value, key string, valN *stringParsing.ParsedNode, conf *shared.Config) (err error) {
	if value == "null" {
		switch vtype {
		case "object":
			err = fmt.Errorf("Object can't be null")
			return
		case "null":
			err = fmt.Errorf("Invalid syntax: need type assertion if value is null")
			return
		default:
			if _, ok := conf.NullV[key]; ok {
				return fmt.Errorf("Can't add null value: key duplicate")
			}
			conf.NullV[key] = vtype
			return
		}
	}
	switch vtype {
	case "string", "object", "int", "bool", "float":
		boolv, intv, floatv, stringv, objectv, err := parseType(vtype, value)
		if err != nil {
			return err
		}
		switch vtype {
		case "string":
			if _, ok := conf.StringV[key]; ok {
				return fmt.Errorf("Can't add string: key is already added")
			}
			conf.StringV[key] = stringv
		case "int":
			if _, ok := conf.IntV[key]; ok {
				return fmt.Errorf("Can't add int: key is already added")
			}
			conf.IntV[key] = intv
		case "bool":
			if _, ok := conf.BoolV[key]; ok {
				return fmt.Errorf("Can't add bool: key is already added")
			}
			conf.BoolV[key] = boolv
		case "float":
			if _, ok := conf.FloatV[key]; ok {
				return fmt.Errorf("Can't add float: key is already added")
			}
			conf.FloatV[key] = floatv
		case "object":
			if _, ok := conf.InnerV[key]; ok {
				return fmt.Errorf("Can't add object: key is already added")
			}
			conf.InnerV[key] = objectv
		}
	default:
		err = setArray(vtype, key, valN, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

func setArray(vtype, key string, valN *stringParsing.ParsedNode, conf *shared.Config) (err error) {
	startIdx := astools.FindChildIndex(valN, "LBRACK") + 1
	finalIdx := astools.FindChildIndex(valN, "RBRACK")
	eltype := vtype[:len(vtype)-1]
	switch eltype {
	case "strings":
		conf.StringArrV[key] = make([]string, 0)
	case "ints":
		conf.IntArrV[key] = make([]int, 0)
	case "floats":
		conf.FloatArrV[key] = make([]float64, 0)
	case "bools":
		conf.BoolArrV[key] = make([]bool, 0)
	case "objects":
		conf.InnerArrV[key] = make([]*shared.Config, 0)
	}
	for i := startIdx; i < finalIdx; i += 2 {
		val := astools.GetChildAt(valN, i)
		boolv, intv, floatv, stringv, objectv, err := parseType(eltype, val.Raw)
		if err != nil {
			return err
		}
		switch eltype {
		case "string":
			conf.StringArrV[key] = append(conf.StringArrV[key], stringv)
		case "int":
			conf.IntArrV[key] = append(conf.IntArrV[key], intv)
		case "float":
			conf.FloatArrV[key] = append(conf.FloatArrV[key], floatv)
		case "bool":
			conf.BoolArrV[key] = append(conf.BoolArrV[key], boolv)
		case "object":
			conf.InnerArrV[key] = append(conf.InnerArrV[key], objectv)
		}
	}
	return
}

func ParseBody(pn *stringParsing.ParsedNode) (conf *shared.Config, err error) {
	conf = shared.NewNilConfig()
	pairs := astools.FindChildren(
		astools.FindChild(
			astools.FindChild(
				pn, "config",
			), "object",
		), "pair",
	)
	for idx, pair := range pairs {
		defer func() {
			if err != nil {
				conf, err2 := format.FormConfig(pn.Raw)
				if err2 != nil {
					conf = pn.Raw
				}
				err = fmt.Errorf(color.Set(
					"[?RD]Code: [?RT]\n%v\n[?RD]Error in: [?RT]\n    %v: '%v': \n[?RD]%v[?RT]",
				), conf, idx, pair.Raw, err)
			}
			return
		}()
		key := astools.FindChild(&pair, "IDENT").Raw
		valueNode := astools.GetChildAt(&pair, astools.FindChildIndex(&pair, "ASSIGN")+1)
		vtype := (valueNode.Switch)
		value := valueNode.Raw
		colonAssign := astools.FindChildIndex(&pair, "COLON")
		if colonAssign != -1 {
			typeNode := astools.GetChildAt(&pair, colonAssign+1)
			if typeNode != nil {
				vtype = typeNode.Raw
			}
		}
		vtype = strings.ToLower(vtype)
		if !shared.IsTypeValid(vtype) {
			err = fmt.Errorf("Invalid value type: %v", vtype)
			return
		}
		if err = setValue(vtype, value, key, valueNode, conf); err != nil {
			return
		}
	}
	return
}
