package shared

import "slices"

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
