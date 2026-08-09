package generation

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/pt-main/tycl/shared"
)

func Toml(conf *shared.Config) (string, error) {
	if conf == nil {
		return "", nil
	}

	obj, err := toTOMLObject(conf)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := toml.NewEncoder(&buf).Encode(obj); err != nil {
		return "", fmt.Errorf("marshal toml: %w", err)
	}
	return buf.String(), nil
}

func toTOMLObject(conf *shared.Config) (map[string]interface{}, error) {
	if conf == nil {
		return nil, nil
	}
	obj := make(map[string]interface{})

	addKey := func(key string, value interface{}) error {
		if _, exists := obj[key]; exists {
			return fmt.Errorf("conflicting key %q (multiple types)", key)
		}
		obj[key] = value
		return nil
	}

	for k, v := range conf.IntV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.FloatV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.BoolV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.StringV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}

	for k, v := range conf.IntArrV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.FloatArrV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.BoolArrV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}
	for k, v := range conf.StringArrV {
		if err := addKey(k, v); err != nil {
			return nil, err
		}
	}

	for k, v := range conf.InnerV {
		subObj, err := toTOMLObject(v)
		if err != nil {
			return nil, fmt.Errorf("object %q: %w", k, err)
		}
		if err := addKey(k, subObj); err != nil {
			return nil, err
		}
	}

	for k, arr := range conf.InnerArrV {
		tomlArr := make([]map[string]interface{}, 0, len(arr))
		for i, item := range arr {
			subObj, err := toTOMLObject(item)
			if err != nil {
				return nil, fmt.Errorf("object array %q index %d: %w", k, i, err)
			}
			tomlArr = append(tomlArr, subObj)
		}
		if err := addKey(k, tomlArr); err != nil {
			return nil, err
		}
	}

	errs := []string{}
	for key, value := range conf.NullV {
		errs = append(errs, fmt.Sprintf("Can't add '%v' (type of %v): toml has no null values support", key, value))
	}
	if len(errs) > 0 {
		return obj, fmt.Errorf("Toml generating: \n%v", strings.Join(errs, "\n- "))
	}

	return obj, nil
}
