package format

import (
	"strings"
	"testing"
)

const withObjectArray = `{
	servers: objects = [
		{
			host: string = "a",
			port: int = 80,
		},
		{
			host: string = "b",
			port: int = 443,
		}
	],
	tags: strings = ["x", "y"],
}`

func TestFormConfigIsIdempotent(t *testing.T) {
	first, err := FormConfig(withObjectArray)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	second, err := FormConfig(first)
	if err != nil {
		t.Fatalf("format twice: %v", err)
	}
	if first != second {
		t.Errorf("formatting is not idempotent:\n%s\n---\n%s", first, second)
	}
}

func TestFormConfigSeparatesArrayItems(t *testing.T) {
	out, err := FormConfig(withObjectArray)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if !strings.Contains(out, "},") {
		t.Errorf("array items are not comma separated:\n%s", out)
	}
	if !strings.Contains(out, "host: string = \"a\"") {
		t.Errorf("array item body lost:\n%s", out)
	}
}

func TestFormConfigKeepsShortScalarArraysInline(t *testing.T) {
	out, err := FormConfig(`{ tags: strings = ["x", "y"] }`)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if !strings.Contains(out, `tags: strings = ["x", "y"]`) {
		t.Errorf("short array was expanded:\n%s", out)
	}
}

func TestFormContractRejectsArrays(t *testing.T) {
	_, err := FormContract(`strict { tags: strings = ["x"] }`)
	if err == nil {
		t.Error("an array in a contract must be rejected")
	}
}
