package tycl

import (
	"strings"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

var Version = "1.4.2"

// Process parses a config and validates it against a contract.
//
// Every failure is reported as a flat diagnostic list: the returned error is
// always a *diag.Error, so callers can render it for humans or as JSON.
func Process(conf, cont string, strictKeys bool) (*shared.Config, core.ErrorInterface) {
	return ProcessSource("", conf, cont, strictKeys)
}

// ProcessSource is Process with a file name, used to label diagnostics.
func ProcessSource(name, conf, cont string, strictKeys bool) (*shared.Config, core.ErrorInterface) {
	contr := shared.NewNillContract()
	if strings.TrimSpace(cont) != "" {
		parsed, err := contract.ParseContract(cont)
		if err != nil {
			return nil, core.Wrap(shared.ContractError, err, "Invalid contract: %v", err)
		}
		contr = parsed
	}
	cfg, err := lang.ParseConf(shared.NewNilConfig(), conf, strictKeys)
	if err != nil {
		return nil, err
	}
	return cfg, CheckContract(cfg, contr)
}

// CheckContract validates a config against a contract and returns a
// *diag.Error listing every violation, or nil when the config is valid.
func CheckContract(cfg *shared.Config, cont *shared.Contract) core.ErrorInterface {
	list := diag.CheckContract(cfg, cont)
	if len(list) == 0 {
		return nil
	}
	return &diag.Error{Diagnostics: list}
}

// Validate parses a config and validates it against a contract, returning
// structured diagnostics instead of an error chain.
func Validate(name, conf, cont string, strictKeys bool) []diag.Diagnostic {
	contr := shared.NewNillContract()
	if strings.TrimSpace(cont) != "" {
		parsed, err := contract.ParseContract(cont)
		if err != nil {
			return diag.FromError(diag.NewSource(name, cont), err)
		}
		contr = parsed
	}
	cfg, err := lang.ParseConf(shared.NewNilConfig(), conf, strictKeys)
	if err != nil {
		return diag.FromError(diag.NewSource(name, conf), err)
	}
	return diag.CheckContract(cfg, contr)
}

// Diagnostics converts any error returned by the API into flat diagnostics.
func Diagnostics(name, text string, err error) []diag.Diagnostic {
	return diag.FromError(diag.NewSource(name, text), err)
}
