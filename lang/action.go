package lang

import (
	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/lc/v2/parsing/stringParsing"
	"github.com/pt-main/lc/v2/tooling/astools"
	"github.com/pt-main/tycl/shared"
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

func (cp *configParser) parseAction(pn *stringParsing.ParsedNode) (vtype string, val string, err core.ErrorInterface) {
	args := GetArgs(pn)
	action := astools.FindChild(pn, "IDENT").Raw
	fn, ok := Actions()[action]
	if !ok {
		err = core.Err(shared.RuntimeError, "Unrecognized action: %v", action)
		return
	}
	var cause error
	vtype, val, cause = fn(cp, pn, args)
	if cause == nil {
		return
	}
	return "", "", core.Wrap(shared.RuntimeError, cause, "Can't parse action (for %v)", pn.Raw)
}
