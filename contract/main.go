package contract

import (
	"slices"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	"github.com/pt-main/tycl/contract/lcproc"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func ParseContract(code string) (*shared.Contract, core.ErrorInterface) {
	p, err := lcproc.NewParser()
	if err != nil {
		return nil, err
	}
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
	contexts := []core.ErrorInterface{}
	for idx, pair := range pairs {
		problems := []core.ErrorInterface{}
		report := func(e core.ErrorInterface) {
			problems = append(problems, e)
			contexts = append(contexts,
				core.Err(shared.ContextedError, "pair").
					WithMeta("idx", idx).
					WithMeta("raw", pair.Raw).
					WithMeta("errs", problems))
		}

		key := astools.FindChild(&pair, "IDENT").Raw
		colonAssign := astools.FindChildIndex(&pair, "COLON")
		typeNode := astools.GetChildAt(&pair, colonAssign+1)
		if typeNode == nil {
			report(core.Err(shared.ProcessingError,
				"Contract key %q has no type", key).
				WithMeta("raw", pair.Raw).WithMeta("idx", idx))
			continue
		}
		vtype := strings.ToLower(typeNode.Raw)
		if !utils.IsTypeValid(vtype) {
			report(shared.HintError(
				core.Err(shared.ProcessingError, "Invalid contract type: %v", vtype).
					WithMeta("raw", pair.Raw).WithMeta("idx", idx),
				"known types: null, bool, int, float, string, object, bools, ints, floats, strings, objects",
			))
			continue
		}

		valueNode := astools.FindChild(&pair, "object")
		value := ""
		if valueNode != nil {
			value = valueNode.Raw
		}
		isObject := slices.Contains([]string{"object", "objects"}, vtype)
		if value == "" && isObject {
			report(shared.HintError(
				core.Err(shared.ProcessingError,
					"Contract type %q needs a body", vtype).
					WithMeta("raw", pair.Raw).WithMeta("idx", idx),
				"write '"+key+": object = strict { ... }'",
			))
			continue
		}
		if value != "" && !isObject {
			report(shared.HintError(
				core.Err(shared.ProcessingError, "Contract type %q cannot have a body", vtype).
					WithMeta("raw", pair.Raw).WithMeta("idx", idx),
				"only 'object' and 'objects' accept a body in a contract",
			))
			continue
		}
		if cause := processPair(vtype, key, value, con, valueNode); cause != nil {
			report(core.Wrap(shared.ProcessingError, cause, "%v", cause.GetMsg()).
				WithMeta("raw", pair.Raw).WithMeta("idx", idx))
		}
	}
	if len(contexts) > 0 {
		err = core.Err(shared.ContractError, "contract has %d invalid %s",
			len(contexts), map[bool]string{true: "pairs", false: "pair"}[len(contexts) != 1]).
			WithMeta("errs", contexts)
		return
	}

	for _, comm := range comments {
		comment := comm.Metadata["value"].(string)
		for _, line := range strings.Split(comment, "\n") {
			con.Comments = append(con.Comments, line)
		}
	}
	return
}
