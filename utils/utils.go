package utils

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func OpenF(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("Open: %v", err)
	}
	return string(data), nil
}

func WriteF(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Write: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(data)
	if err != nil {
		return fmt.Errorf("Write: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("Write: %v", err)
	}
	return nil
}

func ReprStringValue(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `"`, `\"`), "\n", `\n`) + `"`
}

func IsTypeValid(vtype string) bool {
	return slices.Contains([]string{
		"null",
		"bool",
		"int",
		"float",
		"string",
		"object",

		"bools",
		"ints",
		"floats",
		"strings",
		"objects",
	}, vtype)
}
