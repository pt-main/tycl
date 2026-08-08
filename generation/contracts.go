package generation

import (
	"fmt"
	"strings"

	"github.com/pt-main/tycl/shared"
)

// ContractFromConfig генерирует контракт из загруженного конфига.
// defaultStrictness — уровень строгости для вложенных объектов.
func ContractFromConfig(cfg *shared.Config, defaultStrictness shared.ContractType) (*shared.Contract, error) {
	if cfg == nil {
		return shared.NewNillContract(), nil
	}

	contract := shared.NewNillContract()
	contract.Type = defaultStrictness

	// Скаляры
	for key := range cfg.BoolV {
		contract.BoolV = append(contract.BoolV, key)
	}
	for key := range cfg.IntV {
		contract.IntV = append(contract.IntV, key)
	}
	for key := range cfg.FloatV {
		contract.FloatV = append(contract.FloatV, key)
	}
	for key := range cfg.StringV {
		contract.StringV = append(contract.StringV, key)
	}

	// Массивы скаляров
	for key := range cfg.BoolArrV {
		contract.BoolArrV = append(contract.BoolArrV, key)
	}
	for key := range cfg.IntArrV {
		contract.IntArrV = append(contract.IntArrV, key)
	}
	for key := range cfg.FloatArrV {
		contract.FloatArrV = append(contract.FloatArrV, key)
	}
	for key := range cfg.StringArrV {
		contract.StringArrV = append(contract.StringArrV, key)
	}

	// Null-значения
	for key, typ := range cfg.NullV {
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

	// Вложенные объекты
	for key, subCfg := range cfg.InnerV {
		subContract, err := ContractFromConfig(subCfg, defaultStrictness)
		if err != nil {
			return nil, fmt.Errorf("object %q: %w", key, err)
		}
		contract.Inner[key] = subContract
	}

	// Массивы объектов — определяем, одинаковы ли структуры элементов
	for key, arr := range cfg.InnerArrV {
		if len(arr) == 0 {
			continue // Пустой массив — не можем определить контракт
		}

		// Генерируем контракт для первого элемента
		firstContract, err := ContractFromConfig(arr[0], defaultStrictness)
		if err != nil {
			return nil, fmt.Errorf("object array %q (first element): %w", key, err)
		}

		// Проверяем, что все элементы имеют такую же структуру
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
			// Сохраняем контракт
			contract.InnerArrV[key] = firstContract
		} else {
			// Оставляем nil — при генерации будет выведено просто `objects`
			contract.InnerArrV[key] = nil
		}
	}

	return contract, nil
}

// contractsEqual сравнивает два контракта на структурное равенство (игнорируя тип строгости).
func contractsEqual(a, b *shared.Contract) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Сравниваем списки ключей (множества)
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

	// Рекурсивно сравниваем вложенные объекты (Inner)
	if len(a.Inner) != len(b.Inner) {
		return false
	}
	for k, v := range a.Inner {
		if !contractsEqual(v, b.Inner[k]) {
			return false
		}
	}

	// Рекурсивно сравниваем контракты для массивов объектов (InnerArrV)
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

// GenerateContractCode генерирует TYCL-код контракта из структуры Contract.
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

	// Скаляры
	for _, key := range contract.BoolV {
		b.WriteString(fmt.Sprintf("    %s: bool,\n", key))
	}
	for _, key := range contract.IntV {
		b.WriteString(fmt.Sprintf("    %s: int,\n", key))
	}
	for _, key := range contract.FloatV {
		b.WriteString(fmt.Sprintf("    %s: float,\n", key))
	}
	for _, key := range contract.StringV {
		b.WriteString(fmt.Sprintf("    %s: string,\n", key))
	}

	// Массивы скаляров
	for _, key := range contract.BoolArrV {
		b.WriteString(fmt.Sprintf("    %s: bools,\n", key))
	}
	for _, key := range contract.IntArrV {
		b.WriteString(fmt.Sprintf("    %s: ints,\n", key))
	}
	for _, key := range contract.FloatArrV {
		b.WriteString(fmt.Sprintf("    %s: floats,\n", key))
	}
	for _, key := range contract.StringArrV {
		b.WriteString(fmt.Sprintf("    %s: strings,\n", key))
	}

	// Вложенные объекты
	for key, subContract := range contract.Inner {
		subCode, err := GenerateContractCode(subContract)
		if err != nil {
			return "", fmt.Errorf("object %q: %w", key, err)
		}
		lines := strings.Split(subCode, "\n")
		for i, line := range lines {
			if i == 0 {
				// Первая строка: "strict {" или "flexible {"
				b.WriteString(fmt.Sprintf("    %s: object = %s\n", key, line))
			} else if i == len(lines)-1 {
				// Последняя строка: "}" — добавляем запятую
				b.WriteString("    " + line + ",\n")
			} else if line != "" {
				b.WriteString("    " + line + "\n")
			}
		}
	}

	// Массивы объектов
	for key, subContract := range contract.InnerArrV {
		if subContract == nil {
			// Без контракта
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
