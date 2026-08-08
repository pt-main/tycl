package generation

import (
	"encoding/json"
	"fmt"

	"github.com/pt-main/tycl/shared"
)

func Json(conf *shared.Config) (string, error) {
	if conf == nil {
		return "null", nil
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
			return "", err
		}
	}
	for k, v := range conf.FloatV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}
	for k, v := range conf.BoolV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}
	for k, v := range conf.StringV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}

	for k := range conf.NullV {
		if err := addKey(k, nil); err != nil {
			return "", err
		}
	}

	for k, v := range conf.IntArrV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}
	for k, v := range conf.FloatArrV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}
	for k, v := range conf.BoolArrV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}
	for k, v := range conf.StringArrV {
		if err := addKey(k, v); err != nil {
			return "", err
		}
	}

	for k, v := range conf.InnerV {
		sub, err := Json(v)
		if err != nil {
			return "", fmt.Errorf("object %q: %w", k, err)
		}

		var subObj interface{}
		if err := json.Unmarshal([]byte(sub), &subObj); err != nil {
			return "", fmt.Errorf("object %q: %w", k, err)
		}
		if err := addKey(k, subObj); err != nil {
			return "", err
		}
	}

	for k, arr := range conf.InnerArrV {
		jsonArr := make([]interface{}, 0, len(arr))
		for i, item := range arr {
			sub, err := Json(item)
			if err != nil {
				return "", fmt.Errorf("object array %q index %d: %w", k, i, err)
			}
			var subObj interface{}
			if err := json.Unmarshal([]byte(sub), &subObj); err != nil {
				return "", fmt.Errorf("object array %q index %d: %w", k, i, err)
			}
			jsonArr = append(jsonArr, subObj)
		}
		if err := addKey(k, jsonArr); err != nil {
			return "", err
		}
	}

	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}
	return string(data), nil
}
