package contract

import (
	"slices"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
	"github.com/pt-main/tycl/contract/lcproc"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func ParseContract(code string) (*shared.Contract, core.ErrorInterface) {
	p := lcproc.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return nil, err
	}
	return ParseBody(&pn[0])
}

func processPair(vtype, key, value string, con *shared.Contract, valueNode *stringParsing.ParsedNode) (err core.ErrorInterface) {
	switch vtype {
	case "null":
		err = core.Err(shared.RuntimeError, "Can't contract null values")
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
			err = core.Err(shared.RuntimeError, "Can't contract object: invalid value: '%v'", value)
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
		contr, err := ParseContract(value)
		if err != nil {
			return err
		}
		con.InnerArrV[key] = contr
	}
	return
}

func ParseBody(pn *stringParsing.ParsedNode) (con *shared.Contract, err core.ErrorInterface) {
	con = shared.NewNillContract()
	obj := astools.FindChild(
		astools.FindChild(
			pn, "config",
		), "object",
	)
	pairs := astools.FindChildren(
		obj, "pair",
	)
	comments := astools.FindChildren(
		obj, "COMMENT",
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
	errs := []core.ErrorInterface{}
	for idx, pair := range pairs {
		defer func() {
			if len(errs) > 0 {
				err = core.Err(shared.ContextedError, "ERR").
					WithMeta("idx", idx).
					WithMeta("raw", pair.Raw).
					WithMeta("errs", errs)
			}
			return
		}()
		key := astools.FindChild(&pair, "IDENT").Raw
		colonAssign := astools.FindChildIndex(&pair, "COLON")
		typeNode := astools.GetChildAt(&pair, colonAssign+1)
		vtype := typeNode.Raw
		vtype = strings.ToLower(vtype)
		if !utils.IsTypeValid(vtype) {
			err = core.Err(shared.ProcessingError, "Invalid type: %v", vtype).
				WithMeta("raw", pair.Raw).WithMeta("idx", idx)
			return
		}
		valueNode := astools.FindChild(&pair, "object")
		value := ""
		if valueNode != nil {
			value = valueNode.Raw
		}
		isObject := slices.Contains([]string{"object", "objects"}, vtype)
		if value == "" && isObject {
			err = core.Err(shared.ProcessingError,
				"Can't contract: assertion value is not object (type '%v', must be object or objects)",
				vtype).WithMeta("raw", pair.Raw).WithMeta("idx", idx)
			return
		}
		if value != "" && !isObject {
			err = core.Err(shared.ProcessingError, "Can't contract: invalid type assertion").
				WithMeta("raw", pair.Raw).WithMeta("idx", idx)
		}
		err_ := processPair(vtype, key, value, con, valueNode)
		if err_ != nil {
			err = core.Wrap(shared.ProcessingError, err_, core.GetRealError(err_)).
				WithMeta("raw", pair.Raw).WithMeta("idx", idx)
			return
		}
	}

	for _, comm := range comments {
		comment := comm.Metadata["value"].(string)
		for _, line := range strings.Split(comment, "\n") {
			con.Comments = append(con.Comments, line)
		}
	}
	return
}
