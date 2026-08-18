# TYCL - Typed Config Language

[![Go Reference](https://pkg.go.dev/badge/github.com/pt-main/tycl.svg)](https://pkg.go.dev/github.com/pt-main/tycl)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-yellow.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/pt-main/tycl)](https://github.com/pt-main/tycl/releases)

```bash
go get github.com/pt-main/tycl
```

**TYCL** - это типизированный язык конфигурации для Go.  
Он даёт строгую типизацию, контракты (схемы) и читаемый синтаксис - без генерации кода и без `interface{}`.

---

## Зачем TYCL?

| Формат | Проблемы | TYCL решает |
|--------|----------|-------------|
| **JSON** | нет комментариев, всё через `interface{}`, нет валидации | строгая типизация, комментарии, контракты |
| **YAML** | чувствителен к пробелам, нет типизации | явные типы, детерминированный парсинг |
| **TOML** | ограничен, нет схем | контракты, гибкие структуры |

TYCL даёт **70% возможностей сложных языков конфигурации** при **30% сложности**.  
Он подходит как **основной формат** для конфигов и как **промежуточное представление** - вы можете писать на TYCL, а затем генерировать JSON, YAML или TOML для интеграции с другими системами.

---

## Установка

### Как библиотека

```bash
go get github.com/pt-main/tycl
```

Используйте в коде:

```go
import "github.com/pt-main/tycl"

cfg, err := tycl.Process(`{ port: int = 8080 }`, `strict { port: int }`)
if err != nil {
    log.Fatal(err)
}
port := cfg.IntV["port"] // 8080
```

### CLI

Скачайте бинарник из [релизов](https://github.com/pt-main/tycl/releases) или установите через `go install`:

```bash
go install github.com/pt-main/tycl/tycl@latest
```

Команды:

- `tycl valid <config> [contract]` - проверка конфига по контракту.
- `tycl syntax <file...>` - проверка синтаксиса и типов (без контракта).
- `tycl fmt <type> <file...>` - форматирование.
- `tycl gen <input> <output> <json|yaml|toml>` - генерация целевого формата.
- `tycl contract <input> <output> <type>` - генерация контракта из конфига.
- `tycl file --path=<path> <cmd> <args...>` - редактирование файла напрямую через cli (подробнее позже).

Некоторые команды поддерживают флаг `--strict-keys` - запрещает дублирование ключей (с любыми типами) в пределах одного объекта (см. в документации cli - `tycl help`).

### Редактирование конфигов через CLI

TYCL позволяет редактировать конфиги прямо из терминала без открытия текстового редактора. Это удобно для скриптов, быстрых правок и автоматизации, а так же для работы с tycl вне Go.

**Синтаксис:**
```bash
tycl file <subcommand> [args...] --path=<config-file> [--strict-keys]
```

**Глобальные флаги:**
- `--path` - путь к TYCL-конфигу (обязательно)
- `--strict-keys` - запрещает дублирование ключей (опционально)

**Доступные субкоманды:**

| Команда | Описание | Пример |
|---------|----------|--------|
| `get <type> <key>` | Вывести значение ключа | `tycl file get --path=config.tycl int port` |
| `set <type> <key> <value>` | Установить значение ключа | `tycl file set --path=config.tycl int port 9090` |
| `remove <type> <key>` | Удалить ключ | `tycl file remove --path=config.tycl int port` |
| `structure` | Показать структуру конфига | `tycl file structure --path=config.tycl` | 
| `help` | | `tycl help` |

#### Примеры

**Получение значения:**
```bash
tycl file get --path=config.tycl int port
```

**Установка скалярного значения:**
```bash
tycl file set --path=config.tycl int port 9090
tycl file set --path=config.tycl string host "localhost"
tycl file set --path=config.tycl null timeout int
```

**Установка массива:**
```bash
tycl file set --path=config.tycl ints ports "8080,8081,8082"
tycl file set --path=config.tycl strings names "dev,prod,stage"
```

**Установка объекта:**
```bash
tycl file set --path=config.tycl object server "{host: string = \"127.0.0.1\", port: int = 8080}"
```

**Установка массива объектов:**
```bash
tycl file set --path=config.tycl objects servers "[{host: string = \"a\", port: int = 80}, {host: string = \"b\", port: int = 443}]"
```

**Удаление ключа:**
```bash
tycl file remove --path=config.tycl int port
tycl file remove --path=config.tycl object server
```

**Просмотр структуры:**
```bash
tycl file structure --path=config.tycl
# Вывод:
# Config structure:
#   int:
#     port
#   string:
#     host
#   objects:
#     servers
```

---

## Синтаксис

TYCL близок к JSON, но каждое поле имеет явный тип.

### Базовые типы

```tycl
port: int = 8080,
rate: float = 1.5,
debug: bool = true,
host: string = "localhost",
```

### Null‑значения

TYCL позволяет указать, что поле существует, но его значение отсутствует, при этом **тип поля известен**:

```tycl
timeout: int = null   /* поле timeout существует, но равно null, тип int */
```

**Правила работы с null:**

1. **Тип обязателен** - `null` всегда сопровождается типом.
2. **Уникальность null** - для одного имени может быть **только одно** null‑значение.  
   Если вы объявили `timeout: int = null`, то нельзя добавить `timeout: string = null`, но можно добавить `timeout: string = "5s"` (не null).
3. **Null ≠ отсутствие поля** - `key: int = null` означает, что поле есть, но его значение не задано. Если поле вовсе не указано в конфиге, оно просто отсутствует (и контракт это заметит).

### Массивы

Массивы строго типизированы и обозначаются **типом во множественном числе**. Тип массива **обязателен**:

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

Все элементы массива должны быть одного типа.

### Объекты (вложенные)

```tycl
server: object = {
    host: string = "127.0.0.1",
    port: int = 8080,
    timeout: int = null,
}
```

Объекты могут быть вложены произвольно:

```tycl
app: object = {
    name: string = "myapp",
    database: object = {
        host: string = "localhost",
        port: int = 5432
    }
}
```

### Комментарии

Поддерживаются **блочные комментарии** `/* ... */` как документация, и они могут располагаться строго **в начале или в конце объекта**.

```tycl
server: object = {
    /* Этот объект описывает сервер */
    host: string = "127.0.0.1",
    port: int = 8080,
    /* Конец описания сервера */
}
```

Однострочные комментарии (`//`) игнорируются при трансляции и удаляются форматтером.

### Экшены (Actions)

TYCL поддерживает вызов функций прямо в значениях. Экшены позволяют читать файлы, подставлять переменные окружения, преобразовывать типы, объединять строки и ссылаться на другие значения конфига.

**Синтаксис:** `имя_экшена(аргументы)`

**Доступные экшены:**

| Экшен | Описание | Пример |
|-------|----------|--------|
| `file("path")` | Читает содержимое файла как строку | `data: string = file("config.json")` |
| `env("VAR", "default", "type")` | Получает переменную окружения (с типом) | `port: int = env("PORT", "8080", "int")` |
| `join(...)` | Объединяет строки | `name: string = join("auth", "-", "service")` |
| `asString(value)` | Преобразует значение в строку | `debug: string = asString({ debug: bool = true })` |
| `asObject(string)` | Преобразует строку (содержащую TYCL-код) в объект | `db: object = asObject(file("db.tycl"))` |
| `get("path", "type")` | Получает значение из конфига по точечному пути и типу | `host: string = get("server.host", "string")` |

#### Подробное описание экшенов

- **`file("path")`**  
  Считывает содержимое файла по указанному пути и возвращает его как строку. Полезно для встраивания внешних конфигов или данных.

  ```tycl
  config: string = file("settings.json")
  ```

- **`env("VAR", "default", "type")`**  
  Получает значение переменной окружения. Если переменная не задана или пуста, используется значение по умолчанию. Третий аргумент задаёт ожидаемый тип - результат будет приведён к этому типу.

  ```tycl
  port: int = env("PORT", "8080", "int")
  host: string = env("HOST", "'localhost'", "string")
  debug: bool = env("DEBUG", "false", "bool")
  ```

- **`join(...)`**  
  Объединяет произвольное количество строковых аргументов в одну строку. Аргументы могут быть как литералами, так и результатами других экшенов.

  ```tycl
  fullName: string = join("Mr. ", "John", " ", "Doe")
  endpoint: string = join("https://", env("API_HOST", "'api'", "string"), ".example.com")
  ```

- **`asString(value)`**  
  Преобразует переданное значение в строку. Обычно используется для отладки или для создания строковых представлений сложных структур.

  ```tycl
  configStr: string = asString({ debug: bool = true, level: int = 2 })
  ```

- **`asObject(string)`**  
  Принимает строку, содержащую TYCL-код, и парсит её в объект. Это позволяет динамически создавать объекты из строковых данных (например, из содержимого файла).

  ```tycl
  dynamicConfig: object = asObject(file("dynamic.tycl"))
  inlineConfig: object = asObject("{ enabled: bool = true }")
  ```

- **`get("path", "type")`**  
  Извлекает значение из любой части конфига (синтаксис пути - `[объект основного конфига].[его вложенный объект].[...]`, `[ключ основного конфига]`, массивы не поддерживаются) по точечному пути (например, `"database.host"`) и приводит его к указанному типу. Путь может проходить через вложенные объекты. Это позволяет переиспользовать значения в разных частях конфига.

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

  Если путь не существует или тип не совпадает, возникает ошибка валидации.

---

#### Пример с несколькими экшенами

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

### Важные правила

1. **Необязательность типов в ключах**  
   Если тип не указан, он выводится из значения:  
   `key = "text"` → `string`, `key = 42` → `int`.  
   **Исключение:** для массивов тип **обязателен**.

2. **Дублирование ключей с разными типами**  
   Разрешено, но **не рекомендуется**, потому что при генерации в JSON/YAML/TOML конфликт приведёт к ошибке.  
   ```tycl
   port: int = 8080,
   port: string = "8080"   /* допустимо, но плохая практика */
   ```
   Чтобы запретить дублирование ключей, используйте флаг `--strict-keys` в cli (в документации cli явно указано где это допустимо).

3. **Null**  
   Может быть только один ключ с данным именем, если он равен `null` (независимо от типа).
   ```tycl
   timeout: int = null,     /* ок */
   timeout: string = null,  /* ошибка: уже есть null для timeout */
   timeout: string = "5s"  /* ок, это не null */
   ```

Функция strict keys (работающая в cli/tycl.Process) запрещает правила 2 и 3, делая все ключи уникальными. 

---

## Структура Config (после парсинга)

Функция `tycl.Process` возвращает объект `*Config`, который содержит отдельные мапы для каждого типа данных. Это позволяет обращаться к значениям **без приведения типов**:

```go
type Config struct {
    IntV    map[string]int
    FloatV  map[string]float64
    BoolV   map[string]bool
    StringV map[string]string
    NullV   map[string]string           // ключ → тип null-значения

    IntArrV    map[string][]int
    FloatArrV  map[string][]float64
    BoolArrV   map[string][]bool
    StringArrV map[string][]string

    InnerV    map[string]*Config        // объекты
    InnerArrV map[string][]*Config      // массивы объектов
}
```

Пример доступа:

```go
cfg, _ := tycl.Process(`{ port: int = 8080, host: string = "localhost" }`, "")
port := cfg.IntV["port"]        // 8080 (int)
host := cfg.StringV["host"]     // "localhost" (string)
```

---

## Генерация других форматов из Go

Пакет `generation` позволяет экспортировать `*Config` в JSON, YAML, TOML и обратно в TYCL:

```go
import "github.com/pt-main/tycl/generation"

jsonStr, err := generation.Json(cfg)
yamlStr, err := generation.Yaml(cfg)
tomlStr, err := generation.Toml(cfg)
tyclStr, err := generation.Tycl(cfg)        // обратно в TYCL
```

Это полезно, если вы загрузили конфиг, изменили его в коде и хотите сохранить в другом формате.

Важно: для генерации TOML в конфиге обязаны отсутствовать null значения (из за ограничений TOML).

---

## Контракты (схемы)

Контракт описывает ожидаемую структуру конфига. Он пишется на том же языке, но вместо значений указываются только типы.

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

**Уровни строгости:**

- `dynamic` - проверка не выполняется (любая структура).
- `flexible` - все перечисленные поля должны присутствовать, лишние разрешены.
- `strict` - точное соответствие (нельзя добавлять лишние поля).

Контракты поддерживают вложенность для объектов и **массивов объектов** (как показано в примере для `test1`). Для массивов объектов контракт применяется к каждому элементу массива.

---

## Генерация других форматов через CLI

CLI-утилита позволяет экспортировать конфиг в JSON, YAML или TOML без написания кода:

```bash
tycl gen config.tycl out.json json
tycl gen config.tycl out.yaml yaml
tycl gen config.tycl out.toml toml
```

Это превращает TYCL в **промежуточный язык**: вы пишете безопасный и читаемый TYCL, а затем генерируете файлы для интеграции с другими системами.

---

## Генерация контрактов

TYCL умеет автоматически генерировать контракты из существующих конфигов:

```bash
tycl contract config.tycl contract.tycl strict
```

Это полезно, когда у вас уже есть конфиг, и вы хотите создать схему для валидации будущих изменений.

---

## Интеграция в Go (полный пример)

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

	// Экспорт в JSON
	fmt.Println(generation.Json(cfg))
	// Экспорт в TOML
	fmt.Println(generation.Toml(cfg))

	// Генерация контракта из конфига
	cont, _ := generation.ContractFromConfig(cfg, shared.ContractStrict)
	contCode, _ := generation.GenerateContractCode(cont)
	fmt.Println(contCode)
}
```

Если контракт не нужен, передайте `""` или `"dynamic{}"` - проверка будет пропущена.

## Vscode Plugin

Скачайте плагин из релиза и установите. 

Функции плагина - 

- Подсветка синтаксиса контрактов и конфига (тип файла не проверяется)
- Автодополнение синтаксиса (дополняются экшны, типы, типы контрактов)

---

## Лицензия

Apache 2.0 - подробности в [LICENSE](LICENSE).

---

By Pt, 2026, написано на Lc и использует Tap.