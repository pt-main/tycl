package lang

import (
	"slices"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	"github.com/pt-main/tycl/lang/lcproc"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func ParseConf(conf *shared.Config, code string, strictKeys bool) (*shared.Config, core.ErrorInterface) {
	return ParseConfAt(conf, code, strictKeys, 0)
}

// ParseConfAt parses config code that starts at a known offset in the
// original file, so errors inside it are reported at real file positions.
func ParseConfAt(conf *shared.Config, code string, strictKeys bool, base int) (*shared.Config, core.ErrorInterface) {
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
		Base:       base,
	}
	return cp.ParseBody()
}

type configParser struct {
	Code       string
	Conf       *shared.Config
	StrictKeys bool
	Node       *stringParsing.ParsedNode
	// Base is the rune offset of Code inside the original file, so that
	// errors raised while parsing a nested object still point at the right
	// place in the file the user actually wrote.
	Base int
}

func (cp *configParser) setValue(vtype, value, key string, valN *stringParsing.ParsedNode) (err core.ErrorInterface) {
	if value == "null" {
		switch vtype {
		case "object":
			err = shared.HintError(
				core.Err(shared.RuntimeError, "Object can't be null"),
				"use an empty object '{ }' instead of null",
			)
			return
		case "null":
			err = shared.HintError(
				core.Err(shared.RuntimeError, "Null value needs a type assertion"),
				"write 'key: int = null' so the type of the missing value stays known",
			)
			return
		default:
			if _, ok := cp.Conf.NullV[key]; ok {
				return shared.HintError(
					core.Err(shared.RuntimeError, "Key %q is already declared as null", key),
					"a key can be null only once, regardless of its type",
				)
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
				return duplicateKey(key, "string")
			}
			cp.Conf.StringV[key] = stringv
		case "int":
			if _, ok := cp.Conf.IntV[key]; ok {
				return duplicateKey(key, "int")
			}
			cp.Conf.IntV[key] = intv
		case "bool":
			if _, ok := cp.Conf.BoolV[key]; ok {
				return duplicateKey(key, "bool")
			}
			cp.Conf.BoolV[key] = boolv
		case "float":
			if _, ok := cp.Conf.FloatV[key]; ok {
				return duplicateKey(key, "float")
			}
			cp.Conf.FloatV[key] = floatv
		case "object":
			if _, ok := cp.Conf.InnerV[key]; ok {
				return duplicateKey(key, "object")
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

func (cp *configParser) setArray(vtype, key string, valN *stringParsing.ParsedNode) (err core.ErrorInterface) {
	startIdx := astools.FindChildIndex(valN, "LBRACK") + 1
	finalIdx := astools.FindChildIndex(valN, "RBRACK")
	eltype := vtype[:len(vtype)-1]
	switch eltype {
	case "string":
		cp.Conf.StringArrV[key] = make([]string, 0)
	case "int":
		cp.Conf.IntArrV[key] = make([]int, 0)
	case "float":
		cp.Conf.FloatArrV[key] = make([]float64, 0)
	case "bool":
		cp.Conf.BoolArrV[key] = make([]bool, 0)
	case "object":
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

// wrapPair builds the per-pair context error consumed by the diagnostics layer.
func wrapPair(idx int, raw string, errs []core.ErrorInterface) core.ErrorInterface {
	return core.Err(shared.ContextedError, "pair").
		WithMeta("idx", idx).
		WithMeta("raw", raw).
		WithMeta("errs", errs)
}

func pluralWord(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

func (cp *configParser) ParseBody() (conf *shared.Config, err core.ErrorInterface) {
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
	contexts := []core.ErrorInterface{}

	for idx, pair := range pairs {
		// anchor lets every error below point at the exact token that caused
		// it, so the rendered caret lands on the mistake.
		anchor := func(node *stringParsing.ParsedNode, e core.ErrorInterface) core.ErrorInterface {
			if node == nil {
				return e
			}
			start := cp.offsetOfFirstToken(node)
			return shared.StartError(e, start, start+len([]rune(node.Raw)))
		}

		pairErrs := []core.ErrorInterface{}
		fail := func(e core.ErrorInterface) {
			pairErrs = append(pairErrs, e)
		}

		key := astools.FindChild(&pair, "IDENT").Raw
		if cp.StrictKeys {
			if slices.Contains(keys, key) {
				fail(anchor(astools.FindChild(&pair, "IDENT"),
					shared.HintError(
						core.Err(shared.ProcessingError,
							"Duplicate key %q is not allowed with strict keys mode", key,
						).WithMeta("raw", pair.Raw).WithMeta("idx", idx),
						"remove the duplicate, or drop --strict-keys to allow it",
					)))
				contexts = append(contexts, wrapPair(idx, pair.Raw, pairErrs))
				continue
			}
			keys = append(keys, key)
		}

		valueNode := astools.GetChildAt(&pair, astools.FindChildIndex(&pair, "ASSIGN")+1)
		vtype := (valueNode.Switch)
		value := valueNode.Raw
		var typeNode *stringParsing.ParsedNode
		if colonAssign := astools.FindChildIndex(&pair, "COLON"); colonAssign != -1 {
			typeNode = astools.GetChildAt(&pair, colonAssign+1)
			if typeNode != nil {
				vtype = typeNode.Raw
			}
		}
		vtype = strings.ToLower(vtype)
		if !utils.IsTypeValid(vtype) && vtype != "action" {
			bad := typeNode
			if bad == nil {
				bad = valueNode
			}
			fail(anchor(bad, shared.HintError(
				core.Err(shared.ProcessingError, "Invalid value type: %v", vtype).
					WithMeta("raw", pair.Raw).WithMeta("idx", idx),
				typeSuggestion(vtype),
			)))
			contexts = append(contexts, wrapPair(idx, pair.Raw, pairErrs))
			continue
		}

		if setErr := cp.setValue(vtype, value, key, valueNode); setErr != nil {
			fail(anchor(valueNode, setErr))
			contexts = append(contexts, wrapPair(idx, pair.Raw, pairErrs))
		}
	}

	if len(contexts) > 0 {
		err = core.Err(shared.ContextedError, "config has %d invalid %s",
			len(contexts), pluralWord(len(contexts), "pair")).
			WithMeta("errs", contexts)
	}

	for _, comm := range comments {
		comment := comm.Metadata["value"].(string)
		for _, line := range strings.Split(comment, "\n") {
			cp.Conf.Comments = append(cp.Conf.Comments, line)
		}
	}
	return
}
