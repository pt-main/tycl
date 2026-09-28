package lang

import (
	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/tycl/diag"
)

// diagList converts a parse error into flat diagnostics for assertions.
func diagList(code string, err core.ErrorInterface) []diag.Diagnostic {
	return diag.FromError(diag.NewSource("app.tycl", code), err)
}
