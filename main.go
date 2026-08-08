package tycl

import (
	"fmt"
	"strings"

	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

var Version = "1.1.0"

func Process(conf, cont string) (*shared.Config, error) {
	contr := shared.NewNillContract()
	if strings.TrimSpace(cont) != "" {
		var err error
		contr, err = contract.ParseContract(cont)
		if err != nil {
			return nil, err
		}
	}
	cfg, err := lang.ParseConf(conf)
	if err != nil {
		return nil, err
	}
	if err := CheckContract(cfg, contr); err != nil {
		return nil, fmt.Errorf("Invalid config: %w", err)
	}
	return cfg, nil
}
