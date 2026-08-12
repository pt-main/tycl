package lang

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl/lang/lcproc"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func ParseConf(conf *shared.Config, code string, strictKeys bool) (*shared.Config, error) {
	p := lcproc.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return nil, err
	}
	cp := configParser{
		Code:       code,
		StrictKeys: strictKeys,
		Node:       &pn[0],
		Conf:       conf,
	}
	return cp.ParseBody()
}

type configParser struct {
	Code       string
	Conf       *shared.Config
	StrictKeys bool
	Node       *stringParsing.ParsedNode
}

func (cp *configParser) setValue(vtype, value, key string, valN *stringParsing.ParsedNode) (err error) {
	if value == "null" {
		switch vtype {
		case "object":
			err = fmt.Errorf("Object can't be null")
			return
		case "null":
			err = fmt.Errorf("Invalid syntax: need type assertion if value is null")
			return
		default:
			if _, ok := cp.Conf.NullV[key]; ok {
				return fmt.Errorf("Can't add null value: key duplicate")
			}
			cp.Conf.NullV[key] = vtype
			return
		}
	}
	switch vtype {
	case "string", "object", "int", "bool", "float", "action":
		vtype, boolv, intv, floatv, stringv, objectv, err := cp.parseType(valN, vtype, value)
		if err != nil {
			return err
		}
		switch vtype {
		case "string":
			if _, ok := cp.Conf.StringV[key]; ok {
				return fmt.Errorf("Can't add string: key is already added")
			}
			cp.Conf.StringV[key] = stringv
		case "int":
			if _, ok := cp.Conf.IntV[key]; ok {
				return fmt.Errorf("Can't add int: key is already added")
			}
			cp.Conf.IntV[key] = intv
		case "bool":
			if _, ok := cp.Conf.BoolV[key]; ok {
				return fmt.Errorf("Can't add bool: key is already added")
			}
			cp.Conf.BoolV[key] = boolv
		case "float":
			if _, ok := cp.Conf.FloatV[key]; ok {
				return fmt.Errorf("Can't add float: key is already added")
			}
			cp.Conf.FloatV[key] = floatv
		case "object":
			if _, ok := cp.Conf.InnerV[key]; ok {
				return fmt.Errorf("Can't add object: key is already added")
			}
			cp.Conf.InnerV[key] = objectv
		}
	default:
		err = cp.setArray(vtype, key, valN)
		if err != nil {
			return err
		}
	}
	return nil
}

func (cp *configParser) setArray(vtype, key string, valN *stringParsing.ParsedNode) (err error) {
	startIdx := astools.FindChildIndex(valN, "LBRACK") + 1
	finalIdx := astools.FindChildIndex(valN, "RBRACK")
	eltype := vtype[:len(vtype)-1]
	switch eltype {
	case "strings":
		cp.Conf.StringArrV[key] = make([]string, 0)
	case "ints":
		cp.Conf.IntArrV[key] = make([]int, 0)
	case "floats":
		cp.Conf.FloatArrV[key] = make([]float64, 0)
	case "bools":
		cp.Conf.BoolArrV[key] = make([]bool, 0)
	case "objects":
		cp.Conf.InnerArrV[key] = make([]*shared.Config, 0)
	}
	for i := startIdx; i < finalIdx; i += 2 {
		val := astools.GetChildAt(valN, i)
		eltype, boolv, intv, floatv, stringv, objectv, err := cp.parseType(val, eltype, val.Raw)
		if err != nil {
			return err
		}
		switch eltype {
		case "string":
			cp.Conf.StringArrV[key] = append(cp.Conf.StringArrV[key], stringv)
		case "int":
			cp.Conf.IntArrV[key] = append(cp.Conf.IntArrV[key], intv)
		case "float":
			cp.Conf.FloatArrV[key] = append(cp.Conf.FloatArrV[key], floatv)
		case "bool":
			cp.Conf.BoolArrV[key] = append(cp.Conf.BoolArrV[key], boolv)
		case "object":
			cp.Conf.InnerArrV[key] = append(cp.Conf.InnerArrV[key], objectv)
		}
	}
	return
}

func (cp *configParser) ParseBody() (conf *shared.Config, err error) {
	defer func() {
		conf = cp.Conf
	}()

	object := astools.FindChild(
		astools.FindChild(
			cp.Node, "config",
		), "object",
	)
	pairs := astools.FindChildren(object, "pair")
	comments := astools.FindChildren(object, "COMMENT")

	keys := []string{}
	for idx, pair := range pairs {
		defer func() {
			if err != nil {
				err = fmt.Errorf(color.Set(
					"[?RD]Error in: [?RT]\n    %v: '%v': \n[?RD]%v[?RT]",
				), idx, pair.Raw, err)
			}
			return
		}()
		key := astools.FindChild(&pair, "IDENT").Raw
		if cp.StrictKeys {
			if slices.Contains(keys, key) {
				err = fmt.Errorf(
					"Can't add key duplicate (with same or not same type): stcict keys mode enabled",
				)
				return
			}
			keys = append(keys, key)
		}

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
		if !utils.IsTypeValid(vtype) && vtype != "action" {
			err = fmt.Errorf("Invalid value type: %v", vtype)
			return
		}
		if err = cp.setValue(vtype, value, key, valueNode); err != nil {
			return
		}
	}

	for _, comm := range comments {
		comment := comm.Metadata["value"].(string)
		for _, line := range strings.Split(comment, "\n") {
			cp.Conf.Comments = append(cp.Conf.Comments, line)
		}
	}
	return
}
