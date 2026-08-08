package generation

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/pt-main/tycl/shared"
)

func Yaml(conf *shared.Config) (string, error) {
	if conf == nil {
		return "null\n", nil
	}

	obj, err := toYAMLObject(conf)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("marshal yaml: %w", err)
	}
	return string(data), nil
}

func toYAMLObject(conf *shared.Config) (map[string]interface{}, error) {
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
	for k := range conf.NullV {
		if err := addKey(k, nil); err != nil {
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
		subObj, err := toYAMLObject(v)
		if err != nil {
			return nil, fmt.Errorf("object %q: %w", k, err)
		}
		if err := addKey(k, subObj); err != nil {
			return nil, err
		}
	}

	for k, arr := range conf.InnerArrV {
		yamlArr := make([]interface{}, 0, len(arr))
		for i, item := range arr {
			subObj, err := toYAMLObject(item)
			if err != nil {
				return nil, fmt.Errorf("object array %q index %d: %w", k, i, err)
			}
			yamlArr = append(yamlArr, subObj)
		}
		if err := addKey(k, yamlArr); err != nil {
			return nil, err
		}
	}

	return obj, nil
}
