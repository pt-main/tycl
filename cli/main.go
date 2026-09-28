package cli

import (
	"fmt"
	"os"

	"github.com/pt-main/tap/go"
	"github.com/pt-main/tap/go/color"
	"github.com/pt-main/tycl"
)

// commandHandler adapts a typed handler to the signature tap expects.
type commandHandler func(*Ctx, []string) *Result

// exitCodeVar records the exit status chosen by the last handler.
//
// It cannot be recovered from the error chain: tap formats a handler failure
// into a flat message with %v, which drops the wrapped error, so the *ExitError
// carrying the code is no longer reachable once Parse returns.
var exitCodeVar = CodeOK

// adapt converts a command handler into a tap handler that owns the
// reporting, the JSON envelope and the exit code.
func adapt(name string, handler commandHandler) tap.HandlerFuncType {
	return func(p *tap.Parser, args []string) error {
		ctx := newContext(p, "")
		res := handler(ctx, args)
		if res.Sources == nil {
			res.Sources = ctx.Sources
		}
		code := Report(name, res, ctx.Format, ctx.Text)
		exitCodeVar = code
		if code != CodeOK {
			return &ExitError{Code: code}
		}
		return nil
	}
}

// NewCli builds the TYCL command line interface.
func NewCli() *tap.Parser {
	p := tap.NewParser(
		"tycl",
		fmt.Sprintf(`[?YW]╭───────[?BGN] Tycl - Typed config language
[?YW]⎬─ [?BBK]Version: %v
[?YW]│  [?BBK]Cli for validating, formatting and converting configs.
[?YW]│  [?BBK][?BD]Humanmade[?RT][?BBK], By [?UE]Pt[?RT]
[?YW]╰───────[?RT]

[?BBK]╭───────Global flags:
[?BBK]⎬─ [?BBK]--json[?RT]          machine-readable output (JSON envelope on stdout)
[?BBK]│    [?BBK]--no-color[?RT]    disable ANSI colors
[?BBK]│    [?BBK]--strict-keys[?RT] forbid duplicate keys in a config
[?BBK]│    [?BBK]--verbose[?RT]     trace command dispatch
[?BBK]│    [?BBK]--debug[?RT]       trace command dispatch and failures
[?BBK]╰───────[?RT]

[?BBK]╭───────Streams:
[?BBK]⎬─ [?BBK]'-'[?RT] as a file path means stdin (input) or stdout (output)
[?BBK]╰───────[?RT]

[?BBK]╭───────Exit codes:
[?BBK]⎬─ [?BBK]0[?RT] ok   [?BBK]1[?RT] invalid data   [?BBK]2[?RT] usage error   [?BBK]3[?RT] i/o error
[?BBK]╰───────[?RT]`, tycl.Version),
		[]string{"help", "h", "-h", "--help"},
		// Verbose and debug output from tap itself; no-color is also
		// handled by applyColorMode, which knows the --no-color spelling.
		tap.NewParserConfig("", "", "", "", "", "", "", true, true),
	)

	p.AddCommand("valid", adapt("valid", validHandler),
		`[?GN]Validate a config file against a contract.[?RT]

[?BBE]Usage:[?RT]
    tycl valid <config-file> [contract-file] [--strict-keys] [--json]

[?BBE]Arguments:[?RT]
    config-file    Path to the config, or '-' for stdin (required)
    contract-file  Path to the contract, or '-' for stdin (optional)

[?BBE]Example:[?RT]
    tycl valid config.tycl contract.tycl
    cat config.tycl | tycl valid - contract.tycl`,
		[]string{"config-file"}, []string{"contract-file"}, false)

	p.AddCommand("syntax", adapt("syntax", syntaxHandler),
		`[?GN]Check syntax and types of many files at once.[?RT]

[?BBE]Usage:[?RT]
    tycl syntax <file1> [file2] ... [--strict-keys] [--json]

[?BBE]Example:[?RT]
    tycl syntax configs/*.tycl
    cat app.tycl | tycl syntax -`,
		nil, nil, true)

	p.AddCommand("fmt", adapt("fmt", fmtHandler),
		`[?GN]Format configs or contracts to the canonical style.[?RT]

[?BBE]Usage:[?RT]
    tycl fmt <conf|contract> <file1> [file2] ... [--json]

[?BBE]Type:[?RT]
    conf, config       Format config files
    cont, contract     Format contract files

[?BBE]Example:[?RT]
    tycl fmt conf config.tycl
    tycl fmt contract schema.tycl
    cat config.tycl | tycl fmt conf -`,
		[]string{"type"}, nil, true)

	p.AddCommand("gen", adapt("gen", genHandler),
		`[?GN]Convert a config into json, yaml, toml or back to tycl.[?RT]

[?BBE]Usage:[?RT]
    tycl gen <input> <output> <json|yaml|toml|tycl> [--contract=<file>] [--json]

[?BBE]Example:[?RT]
    tycl gen app.tycl app.json json
    cat app.tycl | tycl gen - - json`,
		[]string{"file-in", "file-out", "type"}, nil, false)

	p.AddCommand("contract", adapt("contract", contractHandler),
		`[?GN]Generate a contract from an existing config.[?RT]

[?BBE]Usage:[?RT]
    tycl contract <input> <output> <dynamic|flexible|strict>

[?BBE]Example:[?RT]
    tycl contract app.tycl schema.tycl strict`,
		[]string{"file-in", "file-out", "type"}, nil, false)

	p.AddCommand("query", adapt("query", queryHandler),
		`[?GN]Read values from a config by dot-path.[?RT]

[?BBE]Usage:[?RT]
    tycl query <config> <path> [path...] [--json]
    tycl query <config> --json

[?BBE]Path syntax:[?RT]
    key                 a top-level key
    server.host         a key inside a nested object
    servers.0.port      an element of an object array

[?BBE]Example:[?RT]
    tycl query app.tycl server.port --json
    cat app.tycl | tycl query - --json`,
		[]string{"config"}, nil, true)

	p.AddCommand("get", adapt("get", getHandler),
		`[?GN]Read one value from a config.[?RT]

[?BBE]Usage:[?RT]
    tycl get <config> <path> [--json]`,
		[]string{"config", "path"}, nil, false)

	p.AddCommand("set", adapt("set", setHandler),
		`[?GN]Write a value into a config file.[?RT]

[?BBE]Usage:[?RT]
    tycl set <config> <path> <type|auto> <value> [--strict-keys] [--json]

[?BBE]Types:[?RT]
    auto, int, float, bool, string, object
    ints, floats, bools, strings, objects
    use 'null' as the value to declare a typed null

[?BBE]Example:[?RT]
    tycl set app.tycl port auto 8080
    tycl set app.tycl server.host string 127.0.0.1
    tycl set app.tycl timeout int null
    tycl set app.tycl ports ints 8080,8081,8082`,
		[]string{"config", "path", "type", "value"}, nil, false)

	p.AddCommand("remove", adapt("remove", removeHandler),
		`[?GN]Delete a key from a config file.[?RT]

[?BBE]Usage:[?RT]
    tycl remove <config> <path> [--json]`,
		[]string{"config", "path"}, nil, false)

	p.AddCommand("structure", adapt("structure", structureHandler),
		`[?GN]Show every readable path of a config.[?RT]

[?BBE]Usage:[?RT]
    tycl structure <config> [--json]`,
		[]string{"config"}, nil, false)

	p.AddCommand("ast", adapt("ast", astHandler),
		`[?GN]Dump the syntax tree of a file as JSON.[?RT]

[?BBE]Usage:[?RT]
    tycl ast <file> [config|contract] [--json]

[?BBE]Example:[?RT]
    tycl ast app.tycl --json
    cat app.tycl | tycl ast - config --json`,
		[]string{"file"}, []string{"kind"}, false)

	p.AddCommand("docs", adapt("docs", docsHandler),
		`[?GN]Render the documentation comments of a config.[?RT]

[?BBE]Usage:[?RT]
    tycl docs <config> [--json]`,
		[]string{"config"}, nil, false)

	p.AddCommand("types", adapt("types", typesHandler),
		`[?GN]List the type system of the language.[?RT]

[?BBE]Usage:[?RT]
    tycl types [--json]`,
		nil, nil, false)

	p.AddCommand("merge", adapt("merge", mergeHandler),
		`[?GN]Merge several configs; later files win.[?RT]

[?BBE]Usage:[?RT]
    tycl merge <base> <override> [more...] [--json]

[?BBE]Notes:[?RT]
    nested objects merge key by key, so an override file only
    needs to declare the keys it actually changes

[?BBE]Example:[?RT]
    tycl merge base.tycl prod.tycl > merged.tycl`,
		nil, nil, true)

	p.AddCommand("version", adapt("version", versionHandler),
		`[?GN]Print the tycl version.[?RT]`,
		nil, nil, false)

	return p
}

func versionHandler(ctx *Ctx, args []string) *Result {
	if ctx.Format == FormatJSON {
		return Ok(map[string]string{"version": tycl.Version}, "")
	}
	return Document("tycl " + tycl.Version)
}

// Main runs the CLI against os.Args and returns the process exit code.
func Main() int {
	return MainWithArgs(os.Args[1:])
}

// MainWithArgs runs the CLI against an explicit argument list.
func MainWithArgs(args []string) int {
	// Colour is a global switch, so it is restored for every run.
	color.ColorEnabled = true
	exitCodeVar = CodeOK

	p := NewCli()
	err := p.Parse(args)
	if err == nil {
		return CodeOK
	}
	// The handler already reported the problem; only the status is left.
	if exitCodeVar != CodeOK {
		return exitCodeVar
	}
	// Help and usage failures never reached a handler, so the code has to
	// come from the error itself.
	return exitCode(err)
}
