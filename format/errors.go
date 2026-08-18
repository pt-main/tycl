package format

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/lc/parsing/stringParsing/parser3"
	"github.com/pt-main/lc/public/errors"
	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl/shared"
)

var (
	whereSpace    = color.Set("  [?BE]|[?RT]")
	whereRedSpace = color.Set("  [?RD]|[?RT]")
	errSpace      = color.Set("[?RD]>[?RT]  ")
)

func GetErr(ei core.ErrorInterface) (res core.ErrorInterface) {
	inner := ei.Unwrap()
	res, ok := inner.(core.ErrorInterface)
	if !ok {
		if inner != nil {
			res = &core.Error{
				Code:  shared.WrappedError,
				Msg:   inner.Error(),
				Meta:  make(map[errors.ErrorMetaType]interface{}),
				Cause: nil,
			}
		}
	}
	return
}

func addSpace(code, space string, n int) (res string) {
	toAdd := strings.Repeat(space, n)
	res += toAdd
	res += strings.ReplaceAll(code, "\n", "\n"+toAdd)
	return
}

func FormatError(ei core.ErrorInterface) (res string) {
	if ei == nil {
		return ""
	}
	inner := GetErr(ei)
	meta := ei.GetMeta()

	switch errors.ErrorCodeType(ei.GetCode()) {
	case shared.ContextedError:
		errs := meta["errs"].([]core.ErrorInterface)
		res += FormatErrors(errs)
	case shared.ProcessingError:
		idx := meta["idx"].(int) + 1
		raw := meta["raw"].(string)
		idxStr := ""
		switch idx {
		case 1:
			idxStr = "1st"
		case 2:
			idxStr = "2nd"
		case 3:
			idxStr = "3rd"
		default:
			idxStr = strconv.Itoa(idx) + "th"
		}
		res += "Error at " + idxStr + " pair" + ":\n" +
			addSpace(raw, whereSpace, 1) + "\nError:"
	case errors.ParsingError:
		res += color.Set("[?YW]Parsing (1):[?RT]\n")
		res += addSpace(ei.GetMsg(), errSpace, 1)
	case parser3.ParseErrCode, parser3.GrammarErrCode, parser3.AdapterErrCode:
		text := ""
		switch ei.GetMeta()["Code"] {
		case "UnexpectedToken":
			text += fmt.Sprintf("Excepted '%v', got %v\n", ei.GetMeta()["Excepted"], ei.GetMeta()["Got"])
			text += addSpace(ei.GetMeta()["Raw"].(string), whereSpace, 1)
		default:
			text = ei.GetMsg()
		}
		if inner != nil || text != "" {
			res += color.Set("[?YW]Parsing (2):[?RT]")
			if text != "" {
				res += addSpace(text, errSpace, 1)
			}
		}
	case shared.RuntimeError:
		res += color.Set("[?YW]Runtime error:[?RT]\n")
		res += addSpace(ei.GetMsg(), errSpace, 1)
	default:
		res += color.Set("[?YW]Error:[?RT]\n")
		res += addSpace(ei.Error(), errSpace, 1)
	}
	if inner != nil {
		res += "\n"
		res += addSpace(FormatError(inner), whereRedSpace, 1)
	}
	return
}

func FormatErrors(ei []core.ErrorInterface) (res string) {
	for idx, err := range ei {
		res += "Error " + strconv.Itoa(idx+1) + ":\n"
		res += addSpace(FormatError(err), whereSpace, 1)
		if idx != len(ei)-1 {
			res += "\n"
		}
	}
	return res
}
