package diag

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

// configKeys lists every key of a config as a sorted set of type names.
func configKeys(cfg *shared.Config) map[string][]string {
	res := map[string][]string{}
	add := func(key, typ string) {
		res[key] = append(res[key], typ)
	}
	for k := range cfg.IntV {
		add(k, "int")
	}
	for k := range cfg.FloatV {
		add(k, "float")
	}
	for k := range cfg.BoolV {
		add(k, "bool")
	}
	for k := range cfg.StringV {
		add(k, "string")
	}
	for k := range cfg.IntArrV {
		add(k, "ints")
	}
	for k := range cfg.FloatArrV {
		add(k, "floats")
	}
	for k := range cfg.BoolArrV {
		add(k, "bools")
	}
	for k := range cfg.StringArrV {
		add(k, "strings")
	}
	for k := range cfg.InnerV {
		add(k, "object")
	}
	for k := range cfg.InnerArrV {
		add(k, "objects")
	}
	for k, typ := range cfg.NullV {
		add(k, typ)
	}
	for _, types := range res {
		sort.Strings(types)
	}
	return res
}

// contractKeys lists every key declared by a contract as a set of types.
func contractKeys(cont *shared.Contract) map[string]map[string]bool {
	res := map[string]map[string]bool{}
	add := func(key, typ string) {
		if res[key] == nil {
			res[key] = map[string]bool{}
		}
		res[key][typ] = true
	}
	for _, k := range cont.IntV {
		add(k, "int")
	}
	for _, k := range cont.FloatV {
		add(k, "float")
	}
	for _, k := range cont.BoolV {
		add(k, "bool")
	}
	for _, k := range cont.StringV {
		add(k, "string")
	}
	for _, k := range cont.IntArrV {
		add(k, "ints")
	}
	for _, k := range cont.FloatArrV {
		add(k, "floats")
	}
	for _, k := range cont.BoolArrV {
		add(k, "bools")
	}
	for _, k := range cont.StringArrV {
		add(k, "strings")
	}
	for k := range cont.Inner {
		add(k, "object")
	}
	for k := range cont.InnerArrV {
		add(k, "objects")
	}
	return res
}

// CheckContract validates a config against a contract and returns flat
// diagnostics. A dynamic contract accepts anything, a flexible contract
// requires every declared key, and a strict contract also forbids extras.
func CheckContract(cfg *shared.Config, cont *shared.Contract) []Diagnostic {
	if cfg == nil || cont == nil || cont.Type == shared.ContractDynamic {
		return nil
	}
	return checkLevel(cfg, cont, "", cont.Type)
}

func checkLevel(cfg *shared.Config, cont *shared.Contract, path string, level shared.ContractType) []Diagnostic {
	out := []Diagnostic{}

	contKeys := contractKeys(cont)
	cfgKeys := configKeys(cfg)

	declared := make([]string, 0, len(contKeys))
	for key := range contKeys {
		declared = append(declared, key)
	}
	sort.Strings(declared)

	for _, key := range declared {
		path := joinPath(path, key)
		types := contKeys[key]
		if !slicesContainsType(cfgKeys, key, types) {
			out = append(out, Diagnostic{
				Code:     CodeContract,
				Severity: SeverityError,
				Message:  missingKeyMessage(key, types),
				Path:     path,
				Hint:     "add the key to the config, or relax the contract to 'flexible'",
			})
			continue
		}
		if sub, ok := cont.Inner[key]; ok {
			if child, ok := cfg.InnerV[key]; ok {
				out = append(out, checkLevel(child, sub, path, sub.Type)...)
			}
		}
		if sub, ok := cont.InnerArrV[key]; ok && sub != nil {
			for idx, item := range cfg.InnerArrV[key] {
				itemDiags := checkLevel(item, sub, path, sub.Type)
				for i := range itemDiags {
					value := idx
					itemDiags[i].Index = &value
				}
				out = append(out, itemDiags...)
			}
		}
	}

	if level != shared.ContractStrict {
		return out
	}

	extra := make([]string, 0, len(cfgKeys))
	for key := range cfgKeys {
		extra = append(extra, key)
	}
	sort.Strings(extra)

	for _, key := range extra {
		path := joinPath(path, key)
		if _, ok := contKeys[key]; ok {
			continue
		}
		out = append(out, Diagnostic{
			Code:     CodeContract,
			Severity: SeverityError,
			Message:  fmt.Sprintf("key %q is not allowed by this strict contract", key),
			Path:     path,
			Hint:     fmt.Sprintf("declare it in the contract, or use a '%s' contract", shared.ContractFlexible),
		})
	}

	for key, typ := range cfg.NullV {
		if !utils.IsTypeValid(typ) {
			out = append(out, Diagnostic{
				Code:     CodeContract,
				Severity: SeverityError,
				Message:  fmt.Sprintf("null key %q has unknown type %q", key, typ),
				Path:     joinPath(path, key),
			})
			continue
		}
		allowed, ok := contKeys[key]
		if ok && allowed[typ] {
			continue
		}
		out = append(out, Diagnostic{
			Code:     CodeContract,
			Severity: SeverityError,
			Message:  fmt.Sprintf("null key %q is typed %q but the contract declares %s", key, typ, typeList(allowed)),
			Path:     joinPath(path, key),
			Hint:     "make the null type match the contract, or give the key a value",
		})
	}
	return out
}

func slicesContainsType(cfgKeys map[string][]string, key string, types map[string]bool) bool {
	for _, typ := range cfgKeys[key] {
		if types[typ] {
			return true
		}
	}
	return false
}

func missingKeyMessage(key string, types map[string]bool) string {
	switch len(types) {
	case 0:
		return fmt.Sprintf("required key %q is missing", key)
	case 1:
		for typ := range types {
			return fmt.Sprintf("required %s key %q is missing", typ, key)
		}
	}
	return fmt.Sprintf("required key %q is missing (contract allows %s)", key, typeList(types))
}

func typeList(types map[string]bool) string {
	if len(types) == 0 {
		return "nothing"
	}
	out := make([]string, 0, len(types))
	for typ := range types {
		out = append(out, typ)
	}
	sort.Strings(out)
	return strings.Join(out, " or ")
}
