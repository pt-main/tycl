package cli

import (
	"fmt"
	"strings"

	"github.com/pt-main/tycl/contract"
	"github.com/pt-main/tycl/diag"
	"github.com/pt-main/tycl/format"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/lang"
	"github.com/pt-main/tycl/shared"
)

// validateInput parses a config and checks it against an optional contract.
func validateInput(ctx *Ctx, in *Input, contractText, contractName string) (*shared.Config, []diag.Diagnostic) {
	contr := shared.NewNillContract()
	if strings.TrimSpace(contractText) != "" {
		parsed, err := contract.ParseContract(contractText)
		if err != nil {
			src := diag.NewSource(contractName, contractText)
			return nil, diag.FromError(src, err)
		}
		contr = parsed
	}
	cfg, err := lang.ParseConf(shared.NewNilConfig(), in.Text, ctx.StrictKeys())
	if err != nil {
		return nil, diag.FromError(in.Source, err)
	}
	return cfg, diag.CheckContract(cfg, contr)
}

// validHandler implements `tycl valid`.
func validHandler(ctx *Ctx, args []string) *Result {
	in, err := ReadInput(args[0])
	if err != nil {
		return failFromError(err)
	}
	ctx.track(in)
	contractText := ""
	contractName := ""
	if len(args) > 1 {
		c, err := ReadInput(args[1])
		if err != nil {
			return failFromError(err)
		}
		ctx.track(c)
		contractText = c.Text
		contractName = c.Name
	}
	cfg, diags := validateInput(ctx, in, contractText, contractName)
	if len(diags) > 0 {
		return Fail(diags...)
	}
	if ctx.Format == FormatJSON {
		return Ok(newValidationPayload(in.Name, cfg, contractName), "")
	}
	subject := in.Name
	if contractName != "" {
		subject += " against " + contractName
	}
	return Ok(nil, subject+" is valid")
}

// syntaxHandler implements `tycl syntax`.
func syntaxHandler(ctx *Ctx, args []string) *Result {
	files := cleanOptional(args)
	if len(files) == 0 {
		return Usagef("no files to check: pass one or more config files, or '-' for stdin")
	}
	all := []diag.Diagnostic{}
	checked := []string{}
	results := make([]fileResult, 0, len(files))

	for _, path := range files {
		in, err := ReadInput(path)
		if err != nil {
			all = append(all, ioDiagnostic(path, err)...)
			results = append(results, fileResult{File: path, OK: false})
			continue
		}
		ctx.track(in)
		_, diags := validateInput(ctx, in, "", "")
		if len(diags) > 0 {
			all = append(all, diags...)
			results = append(results, fileResult{File: path, OK: false})
			continue
		}
		checked = append(checked, path)
		results = append(results, fileResult{File: path, OK: true})
	}
	return &Result{
		Data:        payloadFor(ctx, filesPayloadOrNil(checked, results)),
		Diagnostics: all,
		Message:     pluralFiles(len(checked)),
	}
}

// fmtHandler implements `tycl fmt`.
func fmtHandler(ctx *Ctx, args []string) *Result {
	kind := args[0]
	files := cleanOptional(args[1:])
	if len(files) == 0 {
		return Usagef("no files to format: pass one or more files, or '-' for stdin")
	}

	var formatter func(string) (string, error)
	switch kind {
	case "conf", "config":
		formatter = func(code string) (string, error) {
			res, err := format.FormConfig(code)
			if err != nil {
				return "", err
			}
			return res, nil
		}
	case "cont", "contract":
		formatter = func(code string) (string, error) {
			res, err := format.FormContract(code)
			if err != nil {
				return "", err
			}
			return res, nil
		}
	default:
		return Usagef(
			"unknown format kind %q: use 'conf' or 'contract'", kind)
	}

	all := []diag.Diagnostic{}
	changed := []string{}
	results := make([]fileResult, 0, len(files))

	for _, path := range files {
		in, err := ReadInput(path)
		if err != nil {
			all = append(all, ioDiagnostic(path, err)...)
			results = append(results, fileResult{File: path, OK: false})
			continue
		}
		ctx.track(in)
		formatted, err := formatter(in.Text)
		if err != nil {
			all = append(all, diag.FromError(in.Source, err)...)
			results = append(results, fileResult{File: path, OK: false})
			continue
		}
		if path == StdinPath {
			return Document(formatted)
		}
		if err := WriteOutput(path, formatted); err != nil {
			all = append(all, ioDiagnostic(path, err)...)
			results = append(results, fileResult{File: path, OK: false})
			continue
		}
		if formatted != in.Text {
			changed = append(changed, path)
		}
		results = append(results, fileResult{File: path, OK: true, Changed: formatted != in.Text})
	}
	return &Result{
		Data:        payloadFor(ctx, filesPayloadOrNil(changed, results)),
		Diagnostics: all,
		Message:     summaryFormatted(changed, results),
	}
}

// genHandler implements `tycl gen`.
func genHandler(ctx *Ctx, args []string) *Result {
	in, err := ReadInput(args[0])
	if err != nil {
		return failFromError(err)
	}
	ctx.track(in)
	target := args[2]
	outPath := args[1]

	contractText := ""
	contractName := ""
	if path, ok := ctx.FlagValue("contract"); ok {
		c, err := ReadInput(path)
		if err != nil {
			return failFromError(err)
		}
		ctx.track(c)
		contractText = c.Text
		contractName = c.Name
	}

	cfg, diags := validateInput(ctx, in, contractText, contractName)
	if len(diags) > 0 {
		return Fail(diags...)
	}

	rendered, err := renderTarget(cfg, target)
	if err != nil {
		return failFromError(err)
	}
	if outPath == StdinPath || outPath == "" {
		return Document(rendered)
	}
	if err := WriteOutput(outPath, rendered); err != nil {
		return failFromError(err)
	}
	return Ok(payloadFor(ctx, genPayload(target, outPath)), "generated "+target+" from "+in.Name)
}

// contractHandler implements `tycl contract`.
func contractHandler(ctx *Ctx, args []string) *Result {
	in, err := ReadInput(args[0])
	if err != nil {
		return failFromError(err)
	}
	ctx.track(in)
	contTypeName := args[2]
	outPath := args[1]

	contType, ok := parseContractType(contTypeName)
	if !ok {
		return Usagef(
			"unknown contract type %q: use 'dynamic', 'flexible' or 'strict'", contTypeName)
	}

	cfg, diags := validateInput(ctx, in, "", "")
	if len(diags) > 0 {
		return Fail(diags...)
	}

	cont, err := generation.ContractFromConfig(cfg, contType)
	if err != nil {
		return failFromError(err)
	}
	code, err := generation.GenerateContractCode(cont)
	if err != nil {
		return failFromError(err)
	}
	if outPath == StdinPath || outPath == "" {
		return Document(code)
	}
	if err := WriteOutput(outPath, code); err != nil {
		return failFromError(err)
	}
	return Ok(payloadFor(ctx, genPayload(contTypeName, outPath)), "generated "+contTypeName+" contract from "+in.Name)
}

func renderTarget(cfg *shared.Config, target string) (string, error) {
	switch target {
	case "json":
		return generation.Json(cfg)
	case "yaml", "yml":
		return generation.Yaml(cfg)
	case "toml":
		return generation.Toml(cfg)
	case "tycl":
		return generation.Tycl(cfg)
	default:
		return "", usage("unknown output format %q: use json, yaml, toml or tycl", target)
	}
}

func parseContractType(name string) (shared.ContractType, bool) {
	switch name {
	case "dynamic":
		return shared.ContractDynamic, true
	case "flexible":
		return shared.ContractFlexible, true
	case "strict":
		return shared.ContractStrict, true
	}
	return shared.ContractDynamic, false
}

type fileResult struct {
	File    string `json:"file"`
	OK      bool   `json:"ok"`
	Changed bool   `json:"changed,omitempty"`
}

type filesPayload struct {
	Files   []fileResult `json:"files"`
	Changed []string     `json:"changed,omitempty"`
}

func filesPayloadOrNil(changed []string, results []fileResult) filesPayload {
	payload := filesPayload{Files: results}
	if len(changed) > 0 {
		payload.Changed = changed
	}
	return payload
}

// payloadFor keeps structured data for JSON output only: in human mode a
// command prints its confirmation line instead of a machine payload.
func payloadFor(ctx *Ctx, data any) any {
	if ctx.Format == FormatJSON {
		return data
	}
	return nil
}

func pluralFiles(n int) string {
	if n == 1 {
		return "1 file checked"
	}
	return fmt.Sprintf("%d files checked", n)
}

func summaryFormatted(changed []string, results []fileResult) string {
	if len(changed) == 0 {
		return "already formatted"
	}
	return fmt.Sprintf("formatted %s", strings.Join(changed, ", "))
}

type validationPayload struct {
	File     string `json:"file"`
	Contract string `json:"contract,omitempty"`
	Keys     int    `json:"keys"`
	Valid    bool   `json:"valid"`
}

func newValidationPayload(file string, cfg *shared.Config, contractName string) validationPayload {
	payload := validationPayload{File: file, Keys: countKeys(cfg), Valid: true}
	if contractName != "" {
		payload.Contract = contractName
	}
	return payload
}

func genPayload(target, out string) map[string]string {
	return map[string]string{"format": target, "output": out}
}

func countKeys(cfg *shared.Config) int {
	if cfg == nil {
		return 0
	}
	return len(cfg.IntV) + len(cfg.FloatV) + len(cfg.BoolV) + len(cfg.StringV) +
		len(cfg.NullV) + len(cfg.IntArrV) + len(cfg.FloatArrV) + len(cfg.BoolArrV) +
		len(cfg.StringArrV) + len(cfg.InnerV) + len(cfg.InnerArrV)
}

func ioDiagnostic(path string, err error) []diag.Diagnostic {
	list := diag.FromError(nil, err)
	if len(list) == 0 {
		return []diag.Diagnostic{{
			Code:     diag.CodeIO,
			Severity: diag.SeverityError,
			Message:  err.Error(),
		}}
	}
	for i := range list {
		if list[i].Span == nil {
			list[i].Span = &diag.Span{File: path}
		}
	}
	return list
}

func failFromError(err error) *Result {
	res := Fail(ioDiagnostic("", err)...)
	if code := exitCode(err); code != CodeFail {
		res.Exit = code
	}
	return res
}
