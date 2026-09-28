package format

import (
	"slices"
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	lcprocC "github.com/pt-main/tycl/contract/lcproc"
	lcprocL "github.com/pt-main/tycl/lang/lcproc"
	"github.com/pt-main/tycl/shared"
)

func FormContract(code string) (string, core.ErrorInterface) {
	p := lcprocC.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return "", err
	}
	return parseUniversal(&pn[0], FormContract)
}

func FormConfig(code string) (string, core.ErrorInterface) {
	p := lcprocL.NewParser()
	pn, err := p.Parse(code)
	if err != nil {
		return "", err
	}
	return parseUniversal(&pn[0], FormConfig)
}

// hasObjectChild reports whether an array contains nested objects.
func hasObjectChild(array *stringParsing.ParsedNode) bool {
	for _, child := range astools.GetChildren(array) {
		if child.Switch == "object" {
			return true
		}
	}
	return false
}

// spaceAfterCommas normalises an inline array so items read as `a, b, c`.
// Commas inside strings are left untouched.
func spaceAfterCommas(raw string) string {
	var b strings.Builder
	var quote rune
	runes := []rune(raw)
	for i, r := range runes {
		switch {
		case quote != 0:
			b.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '"' || r == '\'':
			quote = r
			b.WriteRune(r)
		case r == ',':
			b.WriteRune(r)
			// only pad when the next rune is not whitespace already
			if i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != '\n' {
				b.WriteRune(' ')
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func parseUniversal(pn *stringParsing.ParsedNode, form func(code string) (string, core.ErrorInterface)) (string, core.ErrorInterface) {
	object := astools.FindChild(
		astools.FindChild(
			pn, "config",
		), "object",
	)
	allChildren := astools.GetChildren(object)
	contract := astools.FindChild(object, "CONTRACT")
	isContract := false
	res := ""
	if contract != nil {
		isContract = true
		res += contract.Raw + " "
	}
	res += "{\n"
	tab := "    "
	addComment := func(comment *stringParsing.ParsedNode, tabs int) {
		res += strings.Repeat(tab, tabs) + "/*"
		startTabs := 0
		trimmed := []string{}
		value := comment.Metadata["value"].(string)
		valSplit := strings.Split(value, "\n")
		if len(valSplit) == 1 {
			res += value + "*/" + "\n"
			return
		}
		res += "\n"
		for idx, line := range valSplit {
			trimLine := strings.TrimSpace(line)
			if trimLine != "" {
				trimmed = append(trimmed, line)
			}
			if idx > 0 && idx < len(valSplit)-1 && trimLine == "" {
				trimmed = append(trimmed, "")
			}
		}
		// Re-indent the body relative to the comment's own first line.
		for idx, line := range trimmed {
			linetabs := strings.Count(line, tab) + strings.Count(line, "\t")
			if idx == 0 {
				startTabs = linetabs
			}
			addTabs := linetabs - startTabs
			if addTabs < 0 {
				addTabs = 0
			}
			trimLine := strings.TrimSpace(line)
			res += strings.Repeat(tab, addTabs+tabs+1) + trimLine + "\n"
		}
		res += strings.Repeat(tab, tabs) + "*/" + "\n"
	}

	for _, objChild := range allChildren {
		octype := objChild.Switch
		if octype == "pair" {
			res += tab
			children := astools.GetChildren(&objChild)
			for idx, child := range children {
				ctype := child.Switch
				nextNode := ""
				if idx < len(children)-1 {
					nextNode = children[idx+1].Switch
				}

				if ctype == "array" && isContract {
					return "", core.Err(shared.RuntimeError, "Invalid contract: array at: \n%v", objChild.Raw)
				}

				if ctype == "array" {
					// An array of objects is always expanded: a single line
					// of nested objects cannot be read or reviewed.
					if hasObjectChild(&child) {
						res += "[\n"
						for _, achild := range astools.GetChildren(&child) {
							switch {
							case achild.Switch == "COMMENT":
								addComment(&achild, 2)
							case achild.Raw == "[" || achild.Raw == "]":
								// brackets are emitted around the items
							case achild.Switch == "SEPARATOR":
								// separators inside an object are part of it
							case achild.Switch == "object":
								formed, err := form(achild.Raw)
								if err != nil {
									return "", core.Wrap(shared.WrappedError, err, "%v", err.GetMsg())
								}
								res += tab + tab + strings.ReplaceAll(formed, "\n", "\n"+tab+tab) + ",\n"
							default:
								res += tab + tab + achild.Raw + ",\n"
							}
						}
						res = strings.TrimSuffix(res, ",\n") + "\n"
						res += tab + "]"
						continue
					}
					if len(child.Raw) <= 50 {
						res += spaceAfterCommas(child.Raw)
					} else {
						children := astools.GetChildren(&child)
						for idx, achild := range children {
							if achild.Switch == "COMMENT" {
								addComment(&achild, 2)
								continue
							}
							if achild.Raw == "[" || achild.Switch == "SEPARATOR" && nextNode != "COMMENT" {
								res += achild.Raw + "\n"
								continue
							}
							if achild.Raw == "]" {
								res += tab + achild.Raw
								continue
							}
							child := achild.Raw
							var err core.ErrorInterface
							if achild.Switch == "object" {
								child, err = form(achild.Raw)
								if err != nil {
									return "", core.Wrap(shared.WrappedError, err, "%v", err.GetMsg())
								}
							}
							res += tab + tab + strings.Join(strings.Split(child, "\n"), "\n"+tab+tab)
							if idx == len(children)-2 {
								res += "\n"
							}
						}
					}
				} else if ctype == "object" {
					child, err := form(child.Raw)
					if err != nil {
						return "", err
					}
					if isContract {
						res += " "
					}
					res += strings.ReplaceAll(child, "\n", "\n    ")
				} else {
					res += child.Raw
				}

				if slices.Contains([]string{
					"COLON", "IDENT", "ASSIGN",
				}, ctype) {
					if nextNode != "COLON" && nextNode != "" {
						res += " "
					}
				}
			}
			res += ",\n"
		} else if octype == "COMMENT" {
			addComment(&objChild, 1)
		}
	}
	res += "}"
	return res, nil
}
