package tycl

import (
	"strings"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

var Version = "1.3.8"

func Process(conf, cont string, strictKeys bool) (cfgs *shared.Config, err core.ErrorInterface) {
	contr := shared.NewNillContract()
	if strings.TrimSpace(cont) != "" {
		contr, err = contract.ParseContract(cont)
		if err != nil {
			return nil, err
		}
	}
	mainconf := shared.NewNilConfig()
	cfg, err := lang.ParseConf(mainconf, conf, strictKeys)
	if err != nil {
		return nil, err
	}
	if err := CheckContract(cfg, contr); err != nil {
		return nil, core.Wrap(shared.WrappedError, err, "Invalid config: %v", err)
	}
	return cfg, nil
}
