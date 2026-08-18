# TYCL - Typed Config Language

[![Go Reference](https://pkg.go.dev/badge/github.com/pt-main/tycl.svg)](https://pkg.go.dev/github.com/pt-main/tycl)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-yellow.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/pt-main/tycl)](https://github.com/pt-main/tycl/releases)

```bash
go get github.com/pt-main/tycl
```

**TYCL** is a typed configuration language for Go.  
It provides strong typing, contracts (schemas), and a readable syntax – without code generation and without `interface{}`.

---

## Why TYCL?

| Format  | Problems                                                                 | TYCL solves                                |
|---------|--------------------------------------------------------------------------|--------------------------------------------|
| **JSON**| no comments, everything via `interface{}`, no validation                 | strong typing, comments, contracts         |
| **YAML**| whitespace‑sensitive, no typing                                         | explicit types, deterministic parsing      |
| **TOML**| limited, no schemas                                                     | contracts, flexible structures             |

TYCL gives you **70% of the power** of complex configuration languages with **30% of the complexity**.  
It works both as a **primary format** for configs and as an **intermediate representation** – you can write in TYCL and then generate JSON, YAML, or TOML for integration with other systems.

---

## Installation

### As a library

```bash
go get github.com/pt-main/tycl
```

Use in code:

```go
import "github.com/pt-main/tycl"

cfg, err := tycl.Process(`{ port: int = 8080 }`, `strict { port: int }`)
if err != nil {
    log.Fatal(err)
}
port := cfg.IntV["port"] // 8080
```

### CLI

Download the binary from [releases](https://github.com/pt-main/tycl/releases) or install via `go install`:

```bash
go install github.com/pt-main/tycl/tycl@latest
```

Commands:

- `tycl valid <config> [contract]` – validate a config against a contract.
- `tycl syntax <file...>` – check syntax and types (without a contract).
- `tycl fmt <type> <file...>` – formatting.
- `tycl gen <input> <output> <json|yaml|toml>` – generate target format.
- `tycl contract <input> <output> <type>` – generate a contract from a config.
- `tycl file --path=<path> <cmd> <args...>` – edit a file directly via CLI (more details later).

Some commands support the `--strict-keys` flag – it forbids duplicate keys (of any type) within the same object (see CLI docs – `tycl help`).

### Editing configs via CLI

TYCL allows you to edit configs directly from the terminal without opening a text editor. This is convenient for scripts, quick fixes, automation, and working with TYCL outside Go.

**Syntax:**
```bash
tycl file <subcommand> [args...] --path=<config-file> [--strict-keys]
```

**Global flags:**
- `--path` – path to the TYCL config (required)
- `--strict-keys` – forbids duplicate keys (optional)

**Available subcommands:**

| Command     | Description                              | Example                                               |
|-------------|------------------------------------------|-------------------------------------------------------|
| `get <type> <key>` | Print the value of a key          | `tycl file get --path=config.tycl int port`          |
| `set <type> <key> <value>` | Set a key to a value      | `tycl file set --path=config.tycl int port 9090`     |
| `remove <type> <key>` | Remove a key                    | `tycl file remove --path=config.tycl int port`       |
| `structure`  | Show the structure of the config        | `tycl file structure --path=config.tycl`              |
| `help`       |                                        | `tycl help`                                           |

#### Examples

**Getting a value:**
```bash
tycl file get --path=config.tycl int port
```

**Setting a scalar value:**
```bash
tycl file set --path=config.tycl int port 9090
tycl file set --path=config.tycl string host "localhost"
tycl file set --path=config.tycl null timeout int
```

**Setting an array:**
```bash
tycl file set --path=config.tycl ints ports "8080,8081,8082"
tycl file set --path=config.tycl strings names "dev,prod,stage"
```

**Setting an object:**
```bash
tycl file set --path=config.tycl object server "{host: string = \"127.0.0.1\", port: int = 8080}"
```

**Setting an array of objects:**
```bash
tycl file set --path=config.tycl objects servers "[{host: string = \"a\", port: int = 80}, {host: string = \"b\", port: int = 443}]"
```

**Removing a key:**
```bash
tycl file remove --path=config.tycl int port
tycl file remove --path=config.tycl object server
```

**Viewing structure:**
```bash
tycl file structure --path=config.tycl
# Output:
# Config structure:
#   int:
#     port
#   string:
#     host
#   objects:
#     servers
```

---

## Syntax

TYCL is close to JSON, but every field has an explicit type.

### Basic types

```tycl
port: int = 8080,
rate: float = 1.5,
debug: bool = true,
host: string = "localhost",
```

### Null values

TYCL allows a field to exist but have no value, while **the type of the field is still known**:

```tycl
timeout: int = null   /* field timeout exists but is null, type int */
```

**Rules for null:**

1. **Type is required** – `null` is always accompanied by a type.
2. **Uniqueness of null** – for a given name there can be **only one** null‑valued entry.  
   If you declare `timeout: int = null`, you cannot add `timeout: string = null`, but you can add `timeout: string = "5s"` (not null).
3. **Null ≠ missing field** – `key: int = null` means the field is present but unset. If the field is not mentioned in the config at all, it is simply absent (and the contract will notice that).

### Arrays

Arrays are strictly typed and denoted by the **plural form of the type**. The array type is **required**:

```tycl
ports: ints = [8080, 8081],
names: strings = ["dev", "prod"],
rates: floats = [1.1, 2.2],
flags: bools = [true, false],
servers: objects = [
    { host: string = "a", port: int = 80 },
    { host: string = "b", port: int = 443 }
]
```

All elements of an array must be of the same type.

### Objects (nested)

```tycl
server: object = {
    host: string = "127.0.0.1",
    port: int = 8080,
    timeout: int = null,
}
```

Objects can be nested arbitrarily:

```tycl
app: object = {
    name: string = "myapp",
    database: object = {
        host: string = "localhost",
        port: int = 5432
    }
}
```

### Comments

**Block comments** `/* ... */` are supported as documentation and may be placed strictly **at the beginning or at the end of an object**.

```tycl
server: object = {
    /* This object describes the server */
    host: string = "127.0.0.1",
    port: int = 8080,
    /* End of server description */
}
```

Line comments (`//`) are ignored during translation and removed by the formatter.

### Actions

TYCL supports calling functions directly in values. Actions allow reading files, substituting environment variables, transforming types, joining strings, and referencing other config values.

**Syntax:** `action_name(arguments)`

**Available actions:**

| Action         | Description                                                      | Example                                                           |
|----------------|------------------------------------------------------------------|-------------------------------------------------------------------|
| `file("path")` | Reads the content of a file as a string                         | `data: string = file("config.json")`                             |
| `env("VAR", "default", "type")` | Gets an environment variable (with type conversion) | `port: int = env("PORT", "8080", "int")`                         |
| `join(...)`    | Concatenates strings                                            | `name: string = join("auth", "-", "service")`                    |
| `asString(value)` | Converts a value to a string                                 | `debug: string = asString({ debug: bool = true })`               |
| `asObject(string)` | Parses a string (containing TYCL code) into an object        | `db: object = asObject(file("db.tycl"))`                         |
| `get("path", "type")` | Gets a value from the config by dot‑path and type            | `host: string = get("server.host", "string")`                    |

#### Detailed action descriptions

- **`file("path")`**  
  Reads the file at the given path and returns its content as a string. Useful for embedding external configs or data.

  ```tycl
  config: string = file("settings.json")
  ```

- **`env("VAR", "default", "type")`**  
  Gets the value of an environment variable. If the variable is not set or empty, the default is used. The third argument specifies the expected type – the result will be converted to that type.

  ```tycl
  port: int = env("PORT", "8080", "int")
  host: string = env("HOST", "'localhost'", "string")
  debug: bool = env("DEBUG", "false", "bool")
  ```

- **`join(...)`**  
  Concatenates any number of string arguments into a single string. Arguments can be literals or results of other actions.

  ```tycl
  fullName: string = join("Mr. ", "John", " ", "Doe")
  endpoint: string = join("https://", env("API_HOST", "'api'", "string"), ".example.com")
  ```

- **`asString(value)`**  
  Converts the given value to a string. Typically used for debugging or creating string representations of complex structures.

  ```tycl
  configStr: string = asString({ debug: bool = true, level: int = 2 })
  ```

- **`asObject(string)`**  
  Takes a string containing TYCL code and parses it into an object. This allows dynamic creation of objects from string data (e.g., file contents).

  ```tycl
  dynamicConfig: object = asObject(file("dynamic.tycl"))
  inlineConfig: object = asObject("{ enabled: bool = true }")
  ```

- **`get("path", "type")`**  
  Extracts a value from any part of the config (path syntax: `[main object].[nested object].[...]`, `[main key]`, arrays not supported) by a dot‑separated path (e.g., `"database.host"`) and converts it to the specified type. The path can go through nested objects. This allows reusing values across the config.

  ```tycl
  {
      server: object = {
          host: string = "localhost"
          port: int = 8080
      }
      mainHost: string = get("server.host", "string")   // "localhost"
      mainPort: int = get("server.port", "int")         // 8080
  }
  ```

  If the path does not exist or the type does not match, a validation error occurs.

---

#### Example with multiple actions

```tycl
database: object = asObject(
    file("database.tycl")
),

server: object = {
    host: string = env("SERVER_HOST", "'localhost'", "string"),
    port: int = env("SERVER_PORT", "8080", "int")
},

log: object = {
    level: string = env("LOG_LEVEL", "'info'", "string"),
    file: string = env("LOG_FILE", "'app.log'", "string")
},

modules: strings = [
    join("auth", "-", "service"),
    join("user", "-", "api"),
    join("admin", "-", "ui")
],

debug: string = asString(
    { debug_mode: bool = true }
),

mainHost: string = get("server.host", "string")
```

### Important rules

1. **Type optionality in keys**  
   If a type is omitted, it is inferred from the value:  
   `key = "text"` → `string`, `key = 42` → `int`.  
   **Exception:** for arrays, the type is **required**.

2. **Duplicate keys with different types**  
   Allowed, but **not recommended**, because when generating JSON/YAML/TOML a conflict will cause an error.  
   ```tycl
   port: int = 8080,
   port: string = "8080"   /* allowed, but bad practice */
   ```
   To forbid duplicate keys, use the `--strict-keys` flag in the CLI (the CLI docs clearly state where it is allowed).

3. **Null**  
   There can be only one key with a given name if it is equal to `null` (regardless of type).
   ```tycl
   timeout: int = null,     /* ok */
   timeout: string = null,  /* error: already have null for timeout */
   timeout: string = "5s"   /* ok, this is not null */
   ```

The strict keys feature (available in the CLI / `tycl.Process`) forbids rules 2 and 3, making all keys unique.

---

## Config structure (after parsing)

The `tycl.Process` function returns a `*Config` object, which contains separate maps for each data type. This allows accessing values **without type assertions**:

```go
type Config struct {
    IntV    map[string]int
    FloatV  map[string]float64
    BoolV   map[string]bool
    StringV map[string]string
    NullV   map[string]string           // key → type of null‑value

    IntArrV    map[string][]int
    FloatArrV  map[string][]float64
    BoolArrV   map[string][]bool
    StringArrV map[string][]string

    InnerV    map[string]*Config        // objects
    InnerArrV map[string][]*Config      // arrays of objects
}
```

Example access:

```go
cfg, _ := tycl.Process(`{ port: int = 8080, host: string = "localhost" }`, "")
port := cfg.IntV["port"]        // 8080 (int)
host := cfg.StringV["host"]     // "localhost" (string)
```

---

## Generating other formats from Go

The `generation` package allows exporting a `*Config` to JSON, YAML, TOML, and back to TYCL:

```go
import "github.com/pt-main/tycl/generation"

jsonStr, err := generation.Json(cfg)
yamlStr, err := generation.Yaml(cfg)
tomlStr, err := generation.Toml(cfg)
tyclStr, err := generation.Tycl(cfg)        // back to TYCL
```

This is useful if you load a config, modify it in code, and want to save it in another format.

Important: for TOML generation, null values must be absent from the config (due to TOML limitations).

---

## Contracts (schemas)

A contract describes the expected structure of a config. It is written in the same language, but only types (no values) are given.

```tycl
strict {
    port: int,
    host: string,
    debug: bool,
    timeout: int,
    ports: ints,
    server: object = strict {
        host: string,
        port: int
    },
    test1: objects = flexible {
        key: string
    }
}
```

**Strictness levels:**

- `dynamic` – no validation (any structure is allowed).
- `flexible` – all listed fields must be present, extra fields are allowed.
- `strict` – exact match (no extra fields allowed).

Contracts support nesting for objects and **arrays of objects** (as shown in the example for `test1`). For object arrays, the contract applies to every element of the array.

---

## Generating other formats via CLI

The CLI tool allows exporting a config to JSON, YAML, or TOML without writing code:

```bash
tycl gen config.tycl out.json json
tycl gen config.tycl out.yaml yaml
tycl gen config.tycl out.toml toml
```

This turns TYCL into an **intermediate language**: you write safe and readable TYCL, then generate files for integration with other systems.

---

## Generating contracts

TYCL can automatically generate contracts from existing configs:

```bash
tycl contract config.tycl contract.tycl strict
```

This is useful when you already have a config and want to create a schema for validating future changes.

---

## Integration in Go (full example)

```go
package main

import (
	"fmt"
	"log"

	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/shared"
)

func main() {
	conf := `
{
    port: int = 8080,
    host: string = "localhost",
    timeout: int = -1,
    test1: objects = [
        { key: string = "a" },
        { key: string = "b" }
    ]
}`

	contract := `
strict {
    port: int,
    host: string,
    timeout: int,
    test1: objects = flexible {
        key: string
    }
}`

	cfg, err := tycl.Process(conf, contract, false) // strictKeys=false
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.IntV["port"])       // 8080
	fmt.Println(cfg.StringV["host"])    // "localhost"

	// Export to JSON
	fmt.Println(generation.Json(cfg))
	// Export to TOML
	fmt.Println(generation.Toml(cfg))

	// Generate a contract from the config
	cont, _ := generation.ContractFromConfig(cfg, shared.ContractStrict)
	contCode, _ := generation.GenerateContractCode(cont)
	fmt.Println(contCode)
}
```

If you don't need a contract, pass `""` or `"dynamic{}"` – validation will be skipped.

---

## Vscode Plugin

Download the plugin from the release and install it.

Plugin features:

- Syntax highlighting for contracts and configs (file type is not checked)
- Autocompletion for syntax (actions, types, contract types)

---

## License

Apache 2.0 – see [LICENSE](LICENSE) for details.

---

By Pt, 2026, written in Lc and using Tap.