package cli

import (
	"fmt"

	"github.com/pt-main/tap"
	"github.com/pt-main/tycl"
)

func NewCli() *tap.Parser {
	p := tap.NewParser(
		"tycl",
		fmt.Sprintf(`[?YW]╭───────[?BGN] Tycl - Typed config language
[?YW]⎬─ [?BBK]Version: %v
[?YW]│  [?BBK]Cli for validating and formatting.
[?YW]│  [?BBK][?BD]Humanmade[?RT][?BBK], By [?UE]Pt[?RT]
[?YW]╰───────[?RT]

[?BBK]╭───────Info:
⎬─ Flags:
│    [--strict-keys] - enable strict keys mode
╰───────[?RT]`, tycl.Version),
		[]string{"help"},
		tap.DefaultParserConfig(),
	)

	p.AddCommand("valid", ValidHandler,
		`[?GN]Validate a config file against a contract.[?RT]

[?BBE]Usage:[?RT]
    tycl valid <config-file> [contract-file] [--strict-keys]

[?BBE]Arguments:[?RT]
    config-file    Path to the config file (required)
    contract-file  Path to the contract file (optional, default: dynamic{})

[?BBE]Example:[?RT]
    tycl valid config.tycl contract.tycl
    tycl valid config.tycl`,
		[]string{"config-file", "contract-file"},
		nil,
		false,
	)

	p.AddCommand("syntax", SyntaxHandler,
		`[?GN]Check syntax and types for multiple files (no strict contract).[?RT]

[?BBE]Usage:[?RT]
    tycl syntax <file1> [file2] ... [--strict-keys]

[?BBE]Arguments:[?RT]
    files  List of config files to check (at least one)

[?BBE]Example:[?RT]
    tycl syntax config1.tycl config2.tycl`,
		nil,
		nil,
		true,
	)

	p.AddCommand("fmt", FormatHandler,
		`[?GN]Format config or contract files to a canonical style.[?RT]

[?BBE]Usage:[?RT]
    tycl fmt <type> <file1> [file2] ...

[?BBE]Type:[?RT]
    conf or config    Format config files
    cont or contract  Format contract files

[?BBE]Arguments:[?RT]
    type   Type of formatting (conf/config or cont/contract)
    files  One or more files to format

[?BBE]Examples:[?RT]
    tycl fmt conf config.tycl
    tycl fmt contract contract1.tycl contract2.tycl`,
		[]string{"type"},
		nil,
		true,
	)

	p.AddCommand("gen", GenerateHandler,
		`[?GN]Generate JSON, YAML, or TOML from a TYCL config file.[?RT]

[?BBE]Usage:[?RT]
    tycl gen <input-file> <output-file> <type> [--strict-keys]

[?BBE]Arguments:[?RT]
    input-file   Path to the TYCL config file (required)
    output-file  Path where the generated file will be saved (required)
    type         Output format: json, yaml, or toml (required)

[?BBE]Example:[?RT]
    tycl gen config.tycl config.json json
    tycl gen config.tycl config.yaml yaml
    tycl gen config.tycl config.toml toml`,
		[]string{"file-in", "file-out", "type"},
		nil,
		false,
	)

	p.AddCommand("contract", ContractHandler,
		`[?GN]Generate contract from tycl file.[?RT]

[?BBE]Usage:[?RT]
    tycl contact <input-config> <output-contract> <type> [--strict-keys]

[?BBE]Arguments:[?RT]
    input-file   Path to the TYCL config file (required)
    output-file  Path where the generated contract will be saved (required)
    type         Contract type: string, flexible, or dynamic (required)

[?BBE]Example:[?RT]
    tycl gen config.tycl config.json json
    tycl gen config.tycl config.yaml yaml
    tycl gen config.tycl config.toml toml`,
		[]string{"file-in", "file-out", "type"},
		nil,
		false,
	)

	p.AddCommand("file", FileHandler,
		`[?GN]Edit, view, or remove config values directly from CLI.[?RT]

[?BBE]Usage:[?RT]
    tycl file <subcommand> [args...] --path=<config-file> [--strict-keys]

[?BBE]Global flag for file command:[?RT]
    --path          Path to the TYCL config file (required)
    --strict-keys   Enable strict keys mode (duplicate keys forbidden)

[?BBE]Subcommands:[?RT]

    [?YW]get[?RT] <type> <key>                 [?BBK]# Get the value of a key[?RT]
    [?YW]set[?RT] <type> <key> <value>         [?BBK]# Set the value of a key[?RT]
    [?YW]remove[?RT] <type> <key>              [?BBK]# Remove a key[?RT]
    [?YW]structure[?RT]                        [?BBK]# Show the whole config structure[?RT]
    [?YW]help[?RT]

[?BBE]Examples:[?RT]
    [?BBK]# Get value[?RT]
    tycl file get --path=config.tycl int port

    [?BBK]# Set scalar[?RT]
    tycl file set --path=config.tycl int port 9090
    tycl file set --path=config.tycl string host "localhost"
    tycl file set --path=config.tycl null timeout int

    [?BBK]# Set array[?RT]
    tycl file set --path=config.tycl ints ports "8080,8081,8082"
    tycl file set --path=config.tycl strings names "dev,prod,stage"

    [?BBK]# Set object[?RT]
    tycl file set --path=config.tycl object server "{host: string = \"127.0.0.1\", port: int = 8080}"

    [?BBK]# Set array of objects[?RT]
    tycl file set --path=config.tycl objects servers "[{host: string = \"a\", port: int = 80}, {host: string = \"b\", port: int = 443}]"

    [?BBK]# Remove a key[?RT]
    tycl file remove --path=config.tycl int port
    tycl file remove --path=config.tycl object server

    [?BBK]# Show structure[?RT]
    tycl file structure --path=config.tycl`,
		nil, nil, true,
	)

	return p
}
