package tycl

import (
	"fmt"
	"strings"

	"github.com/pt-main/tycl/shared"
)

func CheckContract(conf *shared.Config, cont *shared.Contract) error {
	if cont == nil {
		return nil
	}
	if cont.Type == shared.ContractDynamic {
		return nil
	}

	hasScalarKey := func(key, typ string) bool {
		switch typ {
		case "int":
			if _, ok := conf.IntV[key]; ok {
				return true
			}
			if t, ok := conf.NullV[key]; ok && t == "int" {
				return true
			}
		case "float":
			if _, ok := conf.FloatV[key]; ok {
				return true
			}
			if t, ok := conf.NullV[key]; ok && t == "float" {
				return true
			}
		case "string":
			if _, ok := conf.StringV[key]; ok {
				return true
			}
			if t, ok := conf.NullV[key]; ok && t == "string" {
				return true
			}
		case "bool":
			if _, ok := conf.BoolV[key]; ok {
				return true
			}
			if t, ok := conf.NullV[key]; ok && t == "bool" {
				return true
			}
		}
		return false
	}

	var errs []string

	for _, key := range cont.IntV {
		if !hasScalarKey(key, "int") {
			errs = append(errs, fmt.Sprintf("required int key %q not found", key))
		}
	}
	for _, key := range cont.FloatV {
		if !hasScalarKey(key, "float") {
			errs = append(errs, fmt.Sprintf("required float key %q not found", key))
		}
	}
	for _, key := range cont.BoolV {
		if !hasScalarKey(key, "bool") {
			errs = append(errs, fmt.Sprintf("required bool key %q not found", key))
		}
	}
	for _, key := range cont.StringV {
		if !hasScalarKey(key, "string") {
			errs = append(errs, fmt.Sprintf("required string key %q not found", key))
		}
	}

	for _, key := range cont.BoolArrV {
		if _, ok := conf.BoolArrV[key]; !ok {
			errs = append(errs, fmt.Sprintf("required bool array key %q not found", key))
		}
	}
	for _, key := range cont.IntArrV {
		if _, ok := conf.IntArrV[key]; !ok {
			errs = append(errs, fmt.Sprintf("required int array key %q not found", key))
		}
	}
	for _, key := range cont.FloatArrV {
		if _, ok := conf.FloatArrV[key]; !ok {
			errs = append(errs, fmt.Sprintf("required float array key %q not found", key))
		}
	}
	for _, key := range cont.StringArrV {
		if _, ok := conf.StringArrV[key]; !ok {
			errs = append(errs, fmt.Sprintf("required string array key %q not found", key))
		}
	}

	for key, subCont := range cont.Inner {
		subConf, ok := conf.InnerV[key]
		if !ok {
			errs = append(errs, fmt.Sprintf("required inner object key %q not found", key))
		} else {
			if err := CheckContract(subConf, subCont); err != nil {
				errs = append(errs, fmt.Sprintf("inner object %q: %v", key, err))
			}
		}
	}

	for key, subCont := range cont.InnerArrV {
		arr, ok := conf.InnerArrV[key]
		if !ok {
			errs = append(errs, fmt.Sprintf("required inner array key %q not found", key))
			continue
		}
		for i, item := range arr {
			if err := CheckContract(item, subCont); err != nil {
				errs = append(errs, fmt.Sprintf("inner array %q index %d: %v", key, i, err))
			}
		}
	}

	if cont.Type == shared.ContractFlexible {
		if len(errs) > 0 {
			return fmt.Errorf("contract validation failed:\n%s", strings.Join(errs, "\n"))
		}
		return nil
	}

	contBoolSet := make(map[string]bool)
	for _, k := range cont.BoolV {
		contBoolSet[k] = true
	}
	contIntSet := make(map[string]bool)
	for _, k := range cont.IntV {
		contIntSet[k] = true
	}
	contFloatSet := make(map[string]bool)
	for _, k := range cont.FloatV {
		contFloatSet[k] = true
	}
	contStringSet := make(map[string]bool)
	for _, k := range cont.StringV {
		contStringSet[k] = true
	}
	contBoolArrSet := make(map[string]bool)
	for _, k := range cont.BoolArrV {
		contBoolArrSet[k] = true
	}
	contIntArrSet := make(map[string]bool)
	for _, k := range cont.IntArrV {
		contIntArrSet[k] = true
	}
	contFloatArrSet := make(map[string]bool)
	for _, k := range cont.FloatArrV {
		contFloatArrSet[k] = true
	}
	contStringArrSet := make(map[string]bool)
	for _, k := range cont.StringArrV {
		contStringArrSet[k] = true
	}
	contInnerSet := make(map[string]bool)
	for k := range cont.Inner {
		contInnerSet[k] = true
	}
	contInnerArrSet := make(map[string]bool)
	for k := range cont.InnerArrV {
		contInnerArrSet[k] = true
	}

	for key := range conf.BoolV {
		if !contBoolSet[key] {
			errs = append(errs, fmt.Sprintf("extra bool key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.IntV {
		if !contIntSet[key] {
			errs = append(errs, fmt.Sprintf("extra int key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.FloatV {
		if !contFloatSet[key] {
			errs = append(errs, fmt.Sprintf("extra float key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.StringV {
		if !contStringSet[key] {
			errs = append(errs, fmt.Sprintf("extra string key %q not allowed in strict contract", key))
		}
	}

	for key := range conf.BoolArrV {
		if !contBoolArrSet[key] {
			errs = append(errs, fmt.Sprintf("extra bool array key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.IntArrV {
		if !contIntArrSet[key] {
			errs = append(errs, fmt.Sprintf("extra int array key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.FloatArrV {
		if !contFloatArrSet[key] {
			errs = append(errs, fmt.Sprintf("extra float array key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.StringArrV {
		if !contStringArrSet[key] {
			errs = append(errs, fmt.Sprintf("extra string array key %q not allowed in strict contract", key))
		}
	}

	for key := range conf.InnerV {
		if !contInnerSet[key] {
			errs = append(errs, fmt.Sprintf("extra inner object key %q not allowed in strict contract", key))
		}
	}
	for key := range conf.InnerArrV {
		if !contInnerArrSet[key] {
			errs = append(errs, fmt.Sprintf("extra inner array key %q not allowed in strict contract", key))
		}
	}

	for key, typ := range conf.NullV {
		if !shared.IsTypeValid(typ) {
			errs = append(errs, fmt.Sprintf("invalid type %q for null key %q", typ, key))
			continue
		}
		switch typ {
		case "bool":
			if !contBoolSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type bool not in contract bool list", key))
			}
		case "int":
			if !contIntSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type int not in contract int list", key))
			}
		case "float":
			if !contFloatSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type float not in contract float list", key))
			}
		case "string":
			if !contStringSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type string not in contract string list", key))
			}
		case "object":
			if !contInnerSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type object not in contract inner list", key))
			}
		case "bools":
			if !contBoolArrSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type bools not in contract bool array list", key))
			}
		case "ints":
			if !contIntArrSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type ints not in contract int array list", key))
			}
		case "floats":
			if !contFloatArrSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type floats not in contract float array list", key))
			}
		case "strings":
			if !contStringArrSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type strings not in contract string array list", key))
			}
		case "objects":
			if !contInnerArrSet[key] {
				errs = append(errs, fmt.Sprintf("null key %q with type objects not in contract inner array list", key))
			}
		default:
			errs = append(errs, fmt.Sprintf("unsupported type %q for null key %q", typ, key))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("strict contract validation failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}
