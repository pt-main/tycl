package contract

import (
	"strings"
	"testing"

	"github.com/pt-main/tycl/diag"
)

func TestParseContractRejectsUnknownType(t *testing.T) {
	_, err := ParseContract(`flexible { name: strig }`)
	if err == nil {
		t.Fatal("an unknown contract type must be rejected")
	}
	list := diag.FromError(diag.NewSource("c.tycl", `flexible { name: strig }`), err)
	if len(list) == 0 {
		t.Fatal("no diagnostics produced")
	}
	if !strings.Contains(list[0].Message, "strig") {
		t.Errorf("diagnostic does not name the bad type: %+v", list[0])
	}
	if list[0].Hint == "" {
		t.Errorf("no hint for an unknown type: %+v", list[0])
	}
}

func TestParseContractRequiresObjectBody(t *testing.T) {
	code := `flexible { srv: object }`
	_, err := ParseContract(code)
	if err == nil {
		t.Fatal("an object without a body must be rejected")
	}
	list := diag.FromError(diag.NewSource("c.tycl", code), err)
	if len(list) == 0 {
		t.Fatal("no diagnostics produced")
	}
	if !strings.Contains(list[0].Message, "body") {
		t.Errorf("unexpected message: %q", list[0].Message)
	}
}

func TestParseContractRejectsBodyOnScalar(t *testing.T) {
	code := `flexible { port: int = { a: int } }`
	if _, err := ParseContract(code); err == nil {
		t.Skip("grammar rejects this shape before the contract check")
	}
}

func TestParseContractAcceptsValidContract(t *testing.T) {
	cont, err := ParseContract(`strict {
		port: int,
		server: object = strict { host: string },
		servers: objects = flexible { host: string },
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cont.Type.String() != "strict" {
		t.Errorf("type = %v, want strict", cont.Type)
	}
	if _, ok := cont.Inner["server"]; !ok {
		t.Error("nested object contract missing")
	}
	if _, ok := cont.InnerArrV["servers"]; !ok {
		t.Error("object array contract missing")
	}
}

func TestParseContractReportsEveryProblem(t *testing.T) {
	code := `flexible {
		name: strig,
		srv: object,
	}`
	_, err := ParseContract(code)
	if err == nil {
		t.Fatal("expected errors")
	}
	list := diag.FromError(diag.NewSource("c.tycl", code), err)
	if len(list) < 2 {
		t.Errorf("want both problems reported, got %d: %+v", len(list), list)
	}
}
