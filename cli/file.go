package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pt-main/tap"
	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
	"github.com/pt-main/tycl/utils"
)

func FileHandler(p *tap.Parser, s []string) (err error) {
	defer func() {
		if err != nil {
			fmt.Println(err)
		}
	}()
	p2 := tap.NewParser("file", ``, []string{"help"}, tap.DefaultParserConfig())
	path, ok := p.Flags["path"]
	if !ok {
		return fmt.Errorf("Need --path value for file command using")
	}
	p2.Flags = p.Flags
	p2.Scope["src"], err = utils.OpenF(path)
	if err != nil {
		return err
	}
	_, strict := p.Flags["strict-keys"]
	p2.Scope["conf"], err = tycl.Process(p2.Scope["src"].(string), "", strict)
	if err != nil {
		return err
	}
	p2.Scope["path"] = path

	p2.AddCommand("get", FileGetHandler, `Get repr of key`, []string{"type", "key"}, nil, false)
	p2.AddCommand("set", FileSetHandler, `Set value of key`, []string{"type", "key", "value"}, nil, false)
	p2.AddCommand("remove", FileRemoveHandler, `Remove key`, []string{"type", "key"}, nil, false)
	p2.AddCommand("struct", FileStructureHandler, `Show structure`, nil, nil, false)

	err = p2.Parse(s)
	return
}

func FileGetHandler(p *tap.Parser, s []string) error {
	res, err := generation.GetRepr(p.Scope["conf"].(*shared.Config), s[0], s[1])
	fmt.Println(res)
	return err
}

func FileSetHandler(p *tap.Parser, s []string) error {
	vtype := s[0]
	key := s[1]
	value := s[2]
	conf := p.Scope["conf"].(*shared.Config)

	fmt.Println(vtype, key, value)

	delete(conf.StringV, key)
	delete(conf.IntV, key)
	delete(conf.FloatV, key)
	delete(conf.BoolV, key)
	delete(conf.NullV, key)
	delete(conf.IntArrV, key)
	delete(conf.FloatArrV, key)
	delete(conf.BoolArrV, key)
	delete(conf.StringArrV, key)
	delete(conf.InnerV, key)
	delete(conf.InnerArrV, key)

	if value == "null" {
		if !utils.IsTypeValid(vtype) {
			return fmt.Errorf("invalid null type: %s", vtype)
		}
		conf.NullV[key] = value
	} else {
		switch vtype {
		case "int":
			val, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid int: %w", err)
			}
			conf.IntV[key] = val

		case "float":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid float: %w", err)
			}
			conf.FloatV[key] = val

		case "bool":
			val, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid bool: %w", err)
			}
			conf.BoolV[key] = val

		case "string":
			conf.StringV[key] = value

		case "ints":
			parts := strings.Split(value, ",")
			arr := make([]int, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				val, err := strconv.Atoi(part)
				if err != nil {
					return fmt.Errorf("invalid int in array: %s", part)
				}
				arr = append(arr, val)
			}
			conf.IntArrV[key] = arr

		case "floats":
			parts := strings.Split(value, ",")
			arr := make([]float64, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				val, err := strconv.ParseFloat(part, 64)
				if err != nil {
					return fmt.Errorf("invalid float in array: %s", part)
				}
				arr = append(arr, val)
			}
			conf.FloatArrV[key] = arr

		case "bools":
			parts := strings.Split(value, ",")
			arr := make([]bool, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				val, err := strconv.ParseBool(part)
				if err != nil {
					return fmt.Errorf("invalid bool in array: %s", part)
				}
				arr = append(arr, val)
			}
			conf.BoolArrV[key] = arr

		case "strings":
			parts := strings.Split(value, ",")
			arr := make([]string, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				arr = append(arr, part)
			}
			conf.StringArrV[key] = arr

		case "object":
			objConf := shared.NewNilConfig()

			parsed, err := lang.ParseConf(objConf, value, false)
			if err != nil {
				return fmt.Errorf("failed to parse object: %w", err)
			}
			conf.InnerV[key] = parsed

		case "objects":
			tmp := fmt.Sprintf("{ temp: objects = %s }", value)
			tmpConf := shared.NewNilConfig()
			parsed, err := lang.ParseConf(tmpConf, tmp, false)
			if err != nil {
				return fmt.Errorf("failed to parse objects array: %w", err)
			}
			if arr, ok := parsed.InnerArrV["temp"]; ok {
				conf.InnerArrV[key] = arr
			} else {
				return fmt.Errorf("failed to extract objects array")
			}

		default:
			return fmt.Errorf("unsupported type: %s", vtype)
		}
	}
	code, err := generation.Tycl(conf)
	if err != nil {
		return err
	}
	path := p.Scope["path"].(string)
	return utils.WriteF(path, code)
}

func FileRemoveHandler(p *tap.Parser, s []string) error {
	vtype := s[0]
	key := s[1]
	conf := p.Scope["conf"].(*shared.Config)

	switch vtype {
	case "int":
		delete(conf.IntV, key)
	case "float":
		delete(conf.FloatV, key)
	case "bool":
		delete(conf.BoolV, key)
	case "string":
		delete(conf.StringV, key)
	case "null":
		delete(conf.NullV, key)
	case "ints":
		delete(conf.IntArrV, key)
	case "floats":
		delete(conf.FloatArrV, key)
	case "bools":
		delete(conf.BoolArrV, key)
	case "strings":
		delete(conf.StringArrV, key)
	case "object":
		delete(conf.InnerV, key)
	case "objects":
		delete(conf.InnerArrV, key)
	default:
		return fmt.Errorf("unsupported type: %s", vtype)
	}

	code, err := generation.Tycl(conf)
	if err != nil {
		return err
	}
	path := p.Scope["path"].(string)
	return utils.WriteF(path, code)
}

func FileStructureHandler(p *tap.Parser, s []string) error {
	conf := p.Scope["conf"].(*shared.Config)
	var b strings.Builder
	b.WriteString("Config structure:\n")
	if len(conf.IntV) > 0 {
		b.WriteString("  int:\n")
		for k := range conf.IntV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.FloatV) > 0 {
		b.WriteString("  float:\n")
		for k := range conf.FloatV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.BoolV) > 0 {
		b.WriteString("  bool:\n")
		for k := range conf.BoolV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.StringV) > 0 {
		b.WriteString("  string:\n")
		for k := range conf.StringV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.NullV) > 0 {
		b.WriteString("  null (type):\n")
		for k, v := range conf.NullV {
			b.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
	}
	if len(conf.IntArrV) > 0 {
		b.WriteString("  ints:\n")
		for k := range conf.IntArrV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.FloatArrV) > 0 {
		b.WriteString("  floats:\n")
		for k := range conf.FloatArrV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.BoolArrV) > 0 {
		b.WriteString("  bools:\n")
		for k := range conf.BoolArrV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.StringArrV) > 0 {
		b.WriteString("  strings:\n")
		for k := range conf.StringArrV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.InnerV) > 0 {
		b.WriteString("  objects:\n")
		for k := range conf.InnerV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	if len(conf.InnerArrV) > 0 {
		b.WriteString("  object arrays:\n")
		for k := range conf.InnerArrV {
			b.WriteString(fmt.Sprintf("    %s\n", k))
		}
	}
	fmt.Print(b.String())
	return nil
}
