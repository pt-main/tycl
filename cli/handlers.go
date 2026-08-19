package cli

import (
	"fmt"

	"github.com/pt-main/lc/engine/core"
	"github.com/pt-main/tap"
	"github.com/pt-main/tap/color"
	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/format"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func ValidHandler(p *tap.Parser, s []string) error {
	config, err := utils.OpenF(s[0])
	if err != nil {
		return err
	}
	contract := "dynamic{}"
	if len(s) > 1 {
		ctr, err := utils.OpenF(s[1])
		if err != nil {
			return err
		}
		contract = ctr
	}
	_, strict := p.Flags["--strict-keys"]
	_, err_ := tycl.Process(config, contract, strict)
	if err_ != nil {
		fmt.Printf(color.Set("[?RD]Validation failed: %v[?RT]\n"), format.FormatError(err_))
		return err_
	}
	color.PrintlnColored("[?GN]Validated[?RT]")
	return nil
}

func SyntaxHandler(p *tap.Parser, s []string) error {
	validated := true
	if len(s) == 0 {
		color.PrintlnColored("[?RD]Has no files to check syntax[?RT]")
		return nil
	}
	for _, arg := range s {
		config, err := utils.OpenF(arg)
		if err != nil {
			color.PrintlnColored("[?RD]Can't open %v:[?RT]\n%v", arg, err)
			validated = false
			continue
		}
		contract := "dynamic{}"
		_, strict := p.Flags["--strict-keys"]
		_, err = tycl.Process(config, contract, strict)
		if err != nil {
			color.PrintlnColored("[?RD]%v validation failed:[?RT]\n%v", arg, err)
			validated = false
			continue
		}
		color.PrintlnColored("[?GN]%v validated[?RT]", arg)
	}
	if !validated {
		color.PrintlnColored("[?RD]Validation failed[?RT]")
	}
	return nil
}

func FormatHandler(p *tap.Parser, s []string) error {
	ftype := s[0]
	files := s[1:]
	if len(files) == 0 {
		color.PrintlnColored("[?RD]Has no files to format[?RT]")
		return nil
	}
	var formfunc func(string) (string, core.ErrorInterface)
	switch ftype {
	case "cont", "contract":
		formfunc = format.FormContract
	case "conf", "config":
		formfunc = format.FormConfig
	default:
		return fmt.Errorf("Can't format: invalid format type")
	}
	for _, file := range files {
		config, err := utils.OpenF(file)
		if err != nil {
			color.PrintlnColored("[?RD]Can't open %v:[?RT]\n%v", file, err)
			continue
		}
		res, err_ := formfunc(config)
		if err_ != nil {
			fmt.Printf(color.Set("[?RD]%v formatting failed:[?RT]\n%v\n"), file, format.FormatError(err_))
			continue
		}
		err = utils.WriteF(file, res)
		if err != nil {
			color.PrintlnColored("[?RD]Can't write %v:[?RT]\n%v", file, err)
			continue
		}
		color.PrintlnColored("[?GN]%v formatted[?RT]", file)
	}
	return nil
}

func GenerateHandler(p *tap.Parser, s []string) error {
	config, err := utils.OpenF(s[0])
	if err != nil {
		return err
	}
	contract := "dynamic{}"
	_, strict := p.Flags["--strict-keys"]
	conf, err_ := tycl.Process(config, contract, strict)
	if err_ != nil {
		fmt.Println(color.Set("[?RD]Error:[?RT]\n"), format.FormatError(err_))
		return err_
	}
	var form func(*shared.Config) (string, error)
	format := s[2]
	switch format {
	case "json":
		form = generation.Json
	case "yaml":
		form = generation.Yaml
	case "toml":
		form = generation.Toml
	default:
		return fmt.Errorf("Invalid format: %v", format)
	}
	res, err := form(conf)
	if err != nil {
		return err
	}
	err = utils.WriteF(s[1], res)
	if err != nil {
		return err
	}
	return nil
}

func ContractHandler(p *tap.Parser, s []string) error {
	config, err := utils.OpenF(s[0])
	if err != nil {
		return err
	}
	contract := "dynamic{}"
	_, strict := p.Flags["--strict-keys"]
	conf, err_ := tycl.Process(config, contract, strict)
	if err_ != nil {
		fmt.Println(color.Set("[?RD]Error:[?RT]\n"), format.FormatError(err_))
		return err_
	}
	var contT shared.ContractType
	ctype := s[2]
	switch ctype {
	case "dynamic":
		contT = shared.ContractDynamic
	case "flexible":
		contT = shared.ContractFlexible
	case "strict":
		contT = shared.ContractStrict
	default:
		return fmt.Errorf("Invalid contract type: %v", ctype)
	}
	cont, err := generation.ContractFromConfig(conf, contT)
	if err != nil {
		return err
	}
	res, err := generation.GenerateContractCode(cont)
	err = utils.WriteF(s[1], res)
	if err != nil {
		return err
	}
	return nil
}
