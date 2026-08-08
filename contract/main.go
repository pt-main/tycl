package contract

import (
	"fmt"
	"strings"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl/contract/lcproc"
	"github.com/pt-main/tycl/shared"
)

func ParseContract(code string) (*shared.Contract, error) {
	p := lcproc.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return nil, err
	}
	return ParseBody(&pn[0])
}

func processPair(vtype, key, value string, con *shared.Contract, valueNode *stringParsing.ParsedNode) (err error) {
	switch vtype {
	case "null":
		err = fmt.Errorf("Can't contract null values")
	case "bool":
		con.BoolV = append(con.BoolV, key)
	case "int":
		con.IntV = append(con.IntV, key)
	case "float":
		con.FloatV = append(con.FloatV, key)
	case "string":
		con.StringV = append(con.StringV, key)
	case "object":
		ctr := astools.FindChild(valueNode, "CONTRACT").Raw
		var contractType shared.ContractType
		switch ctr {
		case "strict":
			contractType = shared.ContractStrict
		case "flexible":
			contractType = shared.ContractFlexible
		case "dynamic":
			contractType = shared.ContractDynamic
		}
		if value == "" {
			err = fmt.Errorf("Can't contract object: invalid value: '%v'", value)
			return
		}
		var inner *shared.Contract
		inner, err = ParseContract(value)
		if err != nil {
			return
		}
		inner.Type = contractType
		con.Inner[key] = inner
	case "strings":
		con.StringArrV = append(con.StringArrV, key)
	case "ints":
		con.IntArrV = append(con.IntArrV, key)
	case "bools":
		con.BoolArrV = append(con.BoolArrV, key)
	case "floats":
		con.FloatArrV = append(con.FloatArrV, key)
	case "objects":
		con.InnerArrV = append(con.InnerArrV, key)
	}
	return
}

func ParseBody(pn *stringParsing.ParsedNode) (con *shared.Contract, err error) {
	con = shared.NewNillContract()
	pairs := astools.FindChildren(
		astools.FindChild(
			astools.FindChild(
				pn, "config",
			), "object",
		), "pair",
	)
	obj := astools.FindChild(
		astools.FindChild(
			pn, "config",
		), "object",
	)
	ctr := astools.FindChild(obj, "CONTRACT").Raw
	switch ctr {
	case "strict":
		con.Type = shared.ContractStrict
	case "flexible":
		con.Type = shared.ContractFlexible
	case "dynamic":
		con.Type = shared.ContractDynamic
	}
	for _, pair := range pairs {
		defer func() {
			if err != nil {
				err = fmt.Errorf(color.Set(
					"[?RD]Code: [?RT]\n%v\n[?RD]Error in: [?RT]'%v': \n[?RD]%v[?RT]",
				), pn.Raw, pair, err)
			}
			return
		}()
		key := astools.FindChild(&pair, "IDENT").Raw
		colonAssign := astools.FindChildIndex(&pair, "COLON")
		typeNode := astools.GetChildAt(&pair, colonAssign+1)
		vtype := typeNode.Raw
		vtype = strings.ToLower(vtype)
		if !shared.IsTypeValid(vtype) {
			err = fmt.Errorf("Invalid type: %v", vtype)
			return
		}
		valueNode := astools.FindChild(&pair, "object")
		value := ""
		if valueNode != nil {
			value = valueNode.Raw
		}
		if value == "" && vtype == "object" {
			err = fmt.Errorf("Can't contract: assertion value is not object (type '%v', must be object)", vtype)
			return
		}
		if err = processPair(vtype, key, value, con, valueNode); err != nil {
			return
		}
	}
	return
}
