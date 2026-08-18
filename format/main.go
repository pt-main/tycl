package format

import (
	"slices"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
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

func parseUniversal(pn *stringParsing.ParsedNode, form func(code string) (string, core.ErrorInterface)) (string, core.ErrorInterface) {
	// find pairs
	// declarate res variable with start of tycl object
	// if object is contract - add contract type to res and set IS_CONTRACT = true
	// (if IS_CONTRACT==true adding contract type, spaces, and block arrays)
	// add every child to res
	// add final bracket to res
	object := astools.FindChild(
		astools.FindChild(
			pn, "config",
		), "object",
	)
	allChildren := astools.GetChildren(object)
	contract := astools.FindChild(object, "CONTRACT")
	IS_CONTRACT := false
	res := ""
	if contract != nil {
		IS_CONTRACT = true
		res += contract.Raw + " "
	}
	res += "{\n"
	tab := "    "
	addComment := func(comment *stringParsing.ParsedNode, tabs int) {
		res += strings.Repeat(tab, tabs) + "/*"
		startTabs := 0
		trimmed := []string{}
		value := comment.Metadata["value"].(string)
		// split comment value and cut first and last spaces
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
		// add lines with formatted tabs
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
				// find child type
				// find next child type
				// if child is object/array add generated code of object/array to res
				// else just add raw child
				// add space if child is color, ident or assign and next node is not colon

				ctype := child.Switch
				nextNode := ""
				if idx < len(children)-1 {
					nextNode = children[idx+1].Switch
				}

				if ctype == "array" && IS_CONTRACT {
					return "", core.Err(shared.RuntimeError, "Invalid contract: array at: \n%v", objChild.Raw)
				}

				if ctype == "array" {
					if len(child.Raw) <= 50 {
						res += child.Raw
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
							// form object value, objects will be formed automaticly
							child := achild.Raw
							var err error
							if achild.Switch == "object" {
								child, err = form(achild.Raw)
								if err != nil {
									return "", core.Wrap(shared.WrappedError, err, err.Error())
								}
							}
							// add tabs to lines
							res += tab + tab + strings.Join(strings.Split(child, "\n"), "\n"+tab+tab)
							// add newline if line is last
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
					if IS_CONTRACT {
						res += " "
					}
					res += strings.ReplaceAll(child, "\n", "\n    ")
				} else {
					res += child.Raw
				}

				if slices.Contains([]string{
					"COLON", "IDENT", "ASSIGN",
				}, ctype) {
					// colon is next only after first ident
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
