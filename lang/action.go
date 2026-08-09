package lang

import (
	"fmt"

	"github.com/pt-main/lc/parsing/stringParsing"
	"github.com/pt-main/lc/tooling/astools"
)

func GetArgs(actionNode *stringParsing.ParsedNode) []stringParsing.ParsedNode {
	if actionNode == nil || actionNode.Switch != "action" {
		return nil
	}

	children := astools.GetChildren(actionNode)
	if len(children) == 0 {
		return nil
	}

	lparenIdx := -1
	for i, child := range children {
		if child.Switch == "LPAREN" {
			lparenIdx = i
			break
		}
	}
	if lparenIdx == -1 {
		return nil
	}

	var args []stringParsing.ParsedNode
	for i := lparenIdx + 1; i < len(children); i++ {
		child := children[i]
		if child.Switch == "RPAREN" {
			break
		}
		if child.Switch == "SEPARATOR" || child.Switch == "COMMENT" {
			continue
		}
		args = append(args, child)
	}

	return args
}

func (cp *configParser) parseAction(pn *stringParsing.ParsedNode) (vtype string, val string, err error) {
	args := GetArgs(pn)
	action := astools.FindChild(pn, "IDENT").Raw
	fn, ok := Actions()[action]
	if !ok {
		err = fmt.Errorf("Unrecognized action: %v", action)
		return
	}
	vtype, val, err = fn(cp, pn, args)
	if err == nil {
		return
	}
	return "ERROR", "ERROR", fmt.Errorf("Can't parse action (for %v): %v", pn.Raw, err)
}
