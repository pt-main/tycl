package tycl

import (
	"fmt"
	"strings"

	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

var Version = "1.3.0"

func Process(conf, cont string, strictKeys bool) (cfgs *shared.Config, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(color.Set(err.Error()))
		}
		return
	}()
	contr := shared.NewNillContract()
	if strings.TrimSpace(cont) != "" {
		var err error
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
		return nil, fmt.Errorf("Invalid config: %w", err)
	}
	return cfg, nil
}
