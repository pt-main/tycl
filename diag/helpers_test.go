package diag

import (
	"encoding/json"
	"testing"

	"github.com/pt-main/lc/v2/engine/core"
	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

func parseError(t *testing.T, code string) core.ErrorInterface {
	t.Helper()
	_, err := lang.ParseConf(shared.NewNilConfig(), code, false)
	if err == nil {
		t.Fatal("expected a parse error")
	}
	return err
}

func parseConfig(t *testing.T, code string) (*shared.Config, error) {
	t.Helper()
	cfg, err := lang.ParseConf(shared.NewNilConfig(), code, false)
	if err != nil {
		t.Fatalf("parse %q: %v", code, err)
	}
	return cfg, nil
}

func parseContract(t *testing.T, code string) *shared.Contract {
	t.Helper()
	cont, err := contract.ParseContract(code)
	if err != nil {
		t.Fatalf("parse contract %q: %v", code, err)
	}
	return cont
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(data)
}

func hasMessage(list []Diagnostic, want string) bool {
	for _, d := range list {
		if d.Message == want {
			return true
		}
	}
	return false
}
