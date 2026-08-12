# TYCL — Typed Config Language

[![Go Reference](https://pkg.go.dev/badge/github.com/pt-main/tycl.svg)](https://pkg.go.dev/github.com/pt-main/tycl)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-yellow.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/pt-main/tycl)](https://github.com/pt-main/tycl/releases)

```bash
go get github.com/pt-main/tycl
```

**TYCL** is a typed configuration language for Go.  
It provides strong typing, contracts (schemas), and a readable syntax — with no code generation and no `interface{}`.

---

## Why TYCL?

| Format | Problems | TYCL solves |
|--------|----------|-------------|
| **JSON** | no comments, everything via `interface{}`, no validation | strong typing, comments, contracts |
| **YAML** | whitespace‑sensitive, no typing | explicit types, deterministic parsing |
| **TOML** | limited, no schemas | contracts, flexible structures |

TYCL gives you **70% of the power of complex configuration languages** at **30% of the complexity**.  
It works both as a **primary configuration format** and as an **intermediate representation** — you can write TYCL and then generate JSON, YAML, or TOML for integration with other systems.

---

## Installation

### As a library (Go module)

```bash
go get github.com/pt-main/tycl
```

Use it in your code:

```go
import "github.com/pt-main/tycl"

cfg, err := tycl.Process(`{ port: int = 8080 }`, `strict { port: int }`)
if err != nil {
    log.Fatal(err)
}
port := cfg.IntV["port"] // 8080
```

### CLI (command‑line tool)

Download a binary from the [releases](https://github.com/pt-main/tycl/releases) page, or install via `go install`:

```bash
go install github.com/pt-main/tycl/tycl@latest
```

Commands:

- `tycl valid <config> [contract]` — validate a config against a contract.
- `tycl syntax <file...>` — check syntax and types (without a contract).
- `tycl fmt <type> <file...>` — format (`conf` / `contract`).
- `tycl gen <input> <output> <json|yaml|toml>` — generate a target format.
- `tycl contract <input> <output> <type>` — generate a contract from a config.

All commands support the `--strict-keys` flag — it forbids duplicate keys (of any type) within the same object.

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

TYCL allows you to specify that a field exists but its value is absent, while **the field type is known**:

```tycl
timeout: int = null   /* field timeout exists but is null, type int */
```

**Rules for null:**

1. **Type is mandatory** — `null` always comes with a type (`int`, `string`, etc.).
2. **Uniqueness of null** — there can be **only one** null value for a given name.  
   If you declare `timeout: int = null`, you cannot add `timeout: string = null`, but you can add `timeout: string = "5s"` (not null).
3. **Null ≠ missing field** — `key: int = null` means the field exists but its value is not set. If the field is not mentioned in the config at all, it is simply absent (and the contract will notice that).

### Arrays

Arrays are strictly typed and denoted by the **type in plural form**. The array type is **mandatory**:

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

All elements of an array must have the same type.

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

**Block comments** `/* ... */` are supported as documentation, and they may be placed **only at the beginning or at the end of an object**.

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

TYCL supports calling functions directly in values. Actions allow you to read files, substitute environment variables, convert types, join strings, and reference other config values.

**Syntax:** `action_name(arguments)`

**Available actions:**

| Action | Description | Example |
|--------|-------------|---------|
| `file("path")` | Reads the contents of a file as a string | `data: string = file("config.json")` |
| `env("VAR", "default", "type")` | Gets an environment variable (with type) | `port: int = env("PORT", "8080", "int")` |
| `join(...)` | Concatenates strings | `name: string = join("auth", "-", "service")` |
| `asString(value)` | Converts a value to a string | `debug: string = asString({ debug: bool = true })` |
| `asObject(string)` | Converts a string (containing TYCL code) to an object | `db: object = asObject(file("db.tycl"))` |
| `get("path", "type")` | Gets a value from the config by dot‑notation path and type | `host: string = get("server.host", "string")` |

#### Detailed action descriptions

- **`file("path")`**  
  Reads the contents of the file at the given path and returns it as a string. Useful for embedding external configs or data.

  ```tycl
  config: string = file("settings.json")
  ```

- **`env("VAR", "default", "type")`**  
  Gets the value of an environment variable. If the variable is not set or empty, the default value is used. The third argument specifies the expected type (`string`, `int`, `bool`, `float`) — the result will be coerced to that type.

  ```tycl
  port: int = env("PORT", "8080", "int")
  host: string = env("HOST", "'localhost'", "string")
  debug: bool = env("DEBUG", "false", "bool")
  ```

- **`join(...)`**  
  Concatenates an arbitrary number of string arguments into a single string. Arguments can be literals or results of other actions.

  ```tycl
  fullName: string = join("Mr. ", "John", " ", "Doe")
  endpoint: string = join("https://", env("API_HOST", "'api'", "string"), ".example.com")
  ```

- **`asString(value)`**  
  Converts the given value (object, number, boolean) into a string. Usually used for debugging or creating string representations of complex structures.

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
  Extracts a value from any part of the config (path syntax: `[root config].[nested object].[...]`, arrays are not supported) using dot‑notation (e.g., `"database.host"`) and coerces it to the specified type. The path can traverse nested objects. This allows reusing values in different parts of the config.

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

1. **Types in keys are optional**  
   If the type is omitted, it is inferred from the value:  
   `key = "text"` → `string`, `key = 42` → `int`.  
   **Exception:** for arrays the type is **mandatory** (`ints`, `strings`, etc.).

2. **Duplicate keys with different types**  
   Allowed, but **not recommended**, because when generating JSON/YAML/TOML, a conflict will cause an error.  
   ```tycl
   port: int = 8080,
   port: string = "8080"   /* allowed, but bad practice */
   ```
   To forbid duplicate keys, use the `--strict-keys` flag in the CLI (the CLI documentation clearly states where this is allowed).

3. **Null**  
   There can be only one key with a given name if it equals `null` (regardless of type).
   ```tycl
   timeout: int = null,     /* ok */
   timeout: string = null,  /* error: null already exists for timeout */
   timeout: string = "5s"   /* ok, this is not null */
   ```

The strict keys feature (available in the CLI and `tycl.Process`) forbids rules 2 and 3, making all keys unique.

---

## Config structure (after parsing)

The `tycl.Process` function returns a `*Config` object that contains separate maps for each data type. This lets you access values **without type assertions**:

```go
type Config struct {
    IntV    map[string]int
    FloatV  map[string]float64
    BoolV   map[string]bool
    StringV map[string]string
    NullV   map[string]string           // key → type of null value

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

This is useful when you load a config, modify it in code, and want to save it in another format.

Important: for TOML generation, the config must have no null values (due to TOML's limitations).

---

## Contracts (schemas)

A contract describes the expected structure of a config. It is written in the same language, but instead of values, only types are given.

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

- `dynamic` — no validation (any structure allowed).
- `flexible` — all listed fields must be present, extra fields are allowed.
- `strict` — exact match (no extra fields allowed).

Contracts support nesting for objects and **arrays of objects** (as shown for `test1`). For arrays of objects, the contract applies to every element of the array.

---

## Generating other formats via CLI

The CLI tool allows exporting a config to JSON, YAML, or TOML without writing any code:

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

If you do not need a contract, pass `""` or `"dynamic{}"` — validation will be skipped.

## VSCode Plugin

Download the plugin from the releases page and install it.

Plugin features:

- Syntax highlighting for contracts and configs (file type is not checked)
- Autocompletion for syntax (actions, types, contract types are completed)

---

## License

Apache 2.0 — see [LICENSE](LICENSE) for details.

---

By Pt, 2026, written in Lc and uses Tap.
