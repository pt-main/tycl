package generation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pt-main/tycl/shared"
)

// sortedKeys returns map keys in a stable order, so generated contracts and
// configs are byte-identical between runs.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func sortedStringKeys(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	sort.Strings(out)
	return out
}

// ContractFromConfig builds a contract that describes an existing config.
// defaultStrictness is the strictness applied to nested objects.
func ContractFromConfig(cfg *shared.Config, defaultStrictness shared.ContractType) (*shared.Contract, error) {
	if cfg == nil {
		return shared.NewNillContract(), nil
	}

	contract := shared.NewNillContract()
	contract.Type = defaultStrictness

	// Scalars
	for _, key := range sortedKeys(cfg.BoolV) {
		contract.BoolV = append(contract.BoolV, key)
	}
	for _, key := range sortedKeys(cfg.IntV) {
		contract.IntV = append(contract.IntV, key)
	}
	for _, key := range sortedKeys(cfg.FloatV) {
		contract.FloatV = append(contract.FloatV, key)
	}
	for _, key := range sortedKeys(cfg.StringV) {
		contract.StringV = append(contract.StringV, key)
	}

	// Scalar arrays
	for _, key := range sortedKeys(cfg.BoolArrV) {
		contract.BoolArrV = append(contract.BoolArrV, key)
	}
	for _, key := range sortedKeys(cfg.IntArrV) {
		contract.IntArrV = append(contract.IntArrV, key)
	}
	for _, key := range sortedKeys(cfg.FloatArrV) {
		contract.FloatArrV = append(contract.FloatArrV, key)
	}
	for _, key := range sortedKeys(cfg.StringArrV) {
		contract.StringArrV = append(contract.StringArrV, key)
	}

	// Typed nulls
	for _, key := range sortedKeys(cfg.NullV) {
		typ := cfg.NullV[key]
		switch typ {
		case "bool":
			contract.BoolV = append(contract.BoolV, key)
		case "int":
			contract.IntV = append(contract.IntV, key)
		case "float":
			contract.FloatV = append(contract.FloatV, key)
		case "string":
			contract.StringV = append(contract.StringV, key)
		case "bools":
			contract.BoolArrV = append(contract.BoolArrV, key)
		case "ints":
			contract.IntArrV = append(contract.IntArrV, key)
		case "floats":
			contract.FloatArrV = append(contract.FloatArrV, key)
		case "strings":
			contract.StringArrV = append(contract.StringArrV, key)
		default:
			return nil, fmt.Errorf("unsupported null type %q for key %q", typ, key)
		}
	}

	// Nested objects
	for _, key := range sortedKeys(cfg.InnerV) {
		subCfg := cfg.InnerV[key]
		subContract, err := ContractFromConfig(subCfg, defaultStrictness)
		if err != nil {
			return nil, fmt.Errorf("object %q: %w", key, err)
		}
		contract.Inner[key] = subContract
	}

	// Object arrays: an element contract is only usable when every element
	// shares the same shape
	for _, key := range sortedKeys(cfg.InnerArrV) {
		arr := cfg.InnerArrV[key]
		if len(arr) == 0 {
			continue // An empty array carries no shape to describe
		}

		// The first element defines the candidate contract
		firstContract, err := ContractFromConfig(arr[0], defaultStrictness)
		if err != nil {
			return nil, fmt.Errorf("object array %q (first element): %w", key, err)
		}

		// Every other element must match it
		allSame := true
		for i := 1; i < len(arr); i++ {
			otherContract, err := ContractFromConfig(arr[i], defaultStrictness)
			if err != nil {
				return nil, fmt.Errorf("object array %q (element %d): %w", key, i, err)
			}
			if !contractsEqual(firstContract, otherContract) {
				allSame = false
				break
			}
		}

		if allSame {
			// Shared shape: keep the contract
			contract.InnerArrV[key] = firstContract
		} else {
			// Mixed shapes: a nil contract renders as a bare 'objects'
			contract.InnerArrV[key] = nil
		}
	}

	return contract, nil
}

// contractsEqual reports structural equality, ignoring strictness.
func contractsEqual(a, b *shared.Contract) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Key sets are compared as unordered collections
	if !stringSlicesEqual(a.BoolV, b.BoolV) {
		return false
	}
	if !stringSlicesEqual(a.IntV, b.IntV) {
		return false
	}
	if !stringSlicesEqual(a.FloatV, b.FloatV) {
		return false
	}
	if !stringSlicesEqual(a.StringV, b.StringV) {
		return false
	}
	if !stringSlicesEqual(a.BoolArrV, b.BoolArrV) {
		return false
	}
	if !stringSlicesEqual(a.IntArrV, b.IntArrV) {
		return false
	}
	if !stringSlicesEqual(a.FloatArrV, b.FloatArrV) {
		return false
	}
	if !stringSlicesEqual(a.StringArrV, b.StringArrV) {
		return false
	}

	// Nested objects
	if len(a.Inner) != len(b.Inner) {
		return false
	}
	for k, v := range a.Inner {
		if !contractsEqual(v, b.Inner[k]) {
			return false
		}
	}

	// Object array contracts
	if len(a.InnerArrV) != len(b.InnerArrV) {
		return false
	}
	for k, v := range a.InnerArrV {
		if !contractsEqual(v, b.InnerArrV[k]) {
			return false
		}
	}

	return true
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]bool, len(a))
	for _, s := range a {
		m[s] = true
	}
	for _, s := range b {
		if !m[s] {
			return false
		}
	}
	return true
}

// GenerateContractCode renders a Contract as TYCL source.
func GenerateContractCode(contract *shared.Contract) (string, error) {
	if contract == nil {
		return "dynamic {}", nil
	}

	var b strings.Builder

	var typeStr string
	switch contract.Type {
	case shared.ContractStrict:
		typeStr = "strict"
	case shared.ContractFlexible:
		typeStr = "flexible"
	default:
		typeStr = "dynamic"
	}
	b.WriteString(typeStr)
	b.WriteString(" {\n")

	// Scalars
	for _, key := range sortedStringKeys(contract.BoolV) {
		b.WriteString(fmt.Sprintf("    %s: bool,\n", key))
	}
	for _, key := range sortedStringKeys(contract.IntV) {
		b.WriteString(fmt.Sprintf("    %s: int,\n", key))
	}
	for _, key := range sortedStringKeys(contract.FloatV) {
		b.WriteString(fmt.Sprintf("    %s: float,\n", key))
	}
	for _, key := range sortedStringKeys(contract.StringV) {
		b.WriteString(fmt.Sprintf("    %s: string,\n", key))
	}

	// Scalar arrays
	for _, key := range sortedStringKeys(contract.BoolArrV) {
		b.WriteString(fmt.Sprintf("    %s: bools,\n", key))
	}
	for _, key := range sortedStringKeys(contract.IntArrV) {
		b.WriteString(fmt.Sprintf("    %s: ints,\n", key))
	}
	for _, key := range sortedStringKeys(contract.FloatArrV) {
		b.WriteString(fmt.Sprintf("    %s: floats,\n", key))
	}
	for _, key := range sortedStringKeys(contract.StringArrV) {
		b.WriteString(fmt.Sprintf("    %s: strings,\n", key))
	}

	// Nested objects
	for _, key := range sortedKeys(contract.Inner) {
		subContract := contract.Inner[key]
		subCode, err := GenerateContractCode(subContract)
		if err != nil {
			return "", fmt.Errorf("object %q: %w", key, err)
		}
		lines := strings.Split(subCode, "\n")
		for i, line := range lines {
			if i == 0 {
				// First line is the nested contract head, e.g. "strict {".
				b.WriteString(fmt.Sprintf("    %s: object = %s\n", key, line))
			} else if i == len(lines)-1 {
				// The closing brace needs the trailing comma.
				b.WriteString("    " + line + ",\n")
			} else if line != "" {
				b.WriteString("    " + line + "\n")
			}
		}
	}

	// Object arrays
	for _, key := range sortedKeys(contract.InnerArrV) {
		subContract := contract.InnerArrV[key]
		if subContract == nil {
			// No element contract available
			b.WriteString(fmt.Sprintf("    %s: objects,\n", key))
		} else {
			subCode, err := GenerateContractCode(subContract)
			if err != nil {
				return "", fmt.Errorf("object array %q: %w", key, err)
			}
			lines := strings.Split(subCode, "\n")
			for i, line := range lines {
				if i == 0 {
					b.WriteString(fmt.Sprintf("    %s: objects = %s\n", key, line))
				} else if i == len(lines)-1 {
					b.WriteString("    " + line + ",\n")
				} else if line != "" {
					b.WriteString("    " + line + "\n")
				}
			}
		}
	}

	b.WriteString("}")
	return b.String(), nil
}
