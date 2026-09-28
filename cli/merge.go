package cli

import (
	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

// mergeHandler implements `tycl merge`: combine several configs.
//
// Later files win on conflicts. This is how a base config and an override
// are combined without writing any Go code.
func mergeHandler(ctx *Ctx, args []string) *Result {
	files := cleanOptional(args)
	if len(files) < 2 {
		return Usagef(
			"merging needs at least two files: tycl merge <base> <override> [more...]")
	}

	all := []diag.Diagnostic{}
	loaded := make([]*shared.Config, 0, len(files))

	for _, path := range files {
		in, err := ReadInput(path)
		if err != nil {
			all = append(all, ioDiagnostic(path, err)...)
			continue
		}
		ctx.track(in)
		cfg, diags := validateInput(ctx, in, "", "")
		if len(diags) > 0 {
			all = append(all, diags...)
			continue
		}
		loaded = append(loaded, cfg)
	}
	if len(loaded) < 2 {
		return &Result{Diagnostics: all}
	}

	merged := loaded[0]
	for _, cfg := range loaded[1:] {
		lang.MergeInto(merged, cfg)
	}

	code, err := generation.Tycl(merged)
	if err != nil {
		return failFromError(err)
	}
	if len(all) > 0 {
		return &Result{Data: code, Diagnostics: all, Stream: true}
	}
	return Document(code)
}
