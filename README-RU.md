# TYCL - Typed Config Language

[![Go Reference](https://pkg.go.dev/badge/github.com/pt-main/tycl.svg)](https://pkg.go.dev/github.com/pt-main/tycl)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-yellow.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/pt-main/tycl)](https://github.com/pt-main/tycl/releases)

```bash
go get github.com/pt-main/tycl
```

**TYCL** - это типизированный язык конфигурации для Go.  
Он даёт строгую типизацию, контракты (схемы) и читаемый синтаксис - без генерации кода и без `interface{}`.

[Ченджлог](docs/changelog-ru.md) | [English version](docs/changelog.md)

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

cfg, err := tycl.Process(`{ port: int = 8080 }`, `strict { port: int }`, false)
if err != nil {
    log.Fatal(err)
}
port := cfg.IntV["port"] // 8080
```

### CLI

Скачайте бинарник из [releases](https://github.com/pt-main/tycl/releases) или установите через `go install`:

```bash
go install github.com/pt-main/tycl/tycl@latest
```

CLI — это полноценный интерфейс к языку: валидация, форматирование, конвертация, чтение и
правка конфигов без единой строки на Go.

#### Соглашения

**Потоки** — путь `-` означает stdin для ввода и stdout для вывода, поэтому любая команда
работает в пайпе:

```bash
cat app.tycl | tycl gen - - json | jq .port
tycl gen app.tycl - yaml > app.yaml
```

**Коды возврата** — предсказуемы для скриптов и CI:

| Код  | Значение                |
|------|-------------------------|
| `0`  | успех                   |
| `1`  | некорректные данные     |
| `2`  | ошибка использования    |
| `3`  | ошибка ввода/вывода     |

**JSON-вывод** — флаг `--json` заставляет любую команду печатать один стабильный конверт
в stdout и оставлять stderr пустым:

```json
{
  "ok": false,
  "command": "valid",
  "diagnostics": [
    {
      "code": "type.invalid",
      "severity": "error",
      "message": "Invalid value type: nope",
      "hint": "unknown type \"nope\", did you mean \"int\"?",
      "span": { "file": "app.tycl", "line": 2, "column": 11, "width": 4 }
    }
  ],
  "data": { }
}
```

**Глобальные флаги** — `--json`, `--no-color`, `--strict-keys`, `--verbose`, `--debug`.

#### Диагностика

Ошибки показываются с точной позицией, строкой исходника и подсказкой:

```console
$ tycl valid app.tycl
error 1st pair: Invalid value type: nope
  app.tycl:2:11
  hint: unknown type "nope", did you mean "int"?
     2 |     port: nope = 8080,
       |           ^^^^
1 error in app.tycl
```

Нарушения контракта отчитываются по каждому ключу и индексу массива, а `--json` отдаёт их
полями `code` / `message` / `hint` / `span` / `path` / `index` — готовыми для редактора или
language server.

#### Команды

**Проверка и форматирование**

```bash
tycl valid <config> [contract] [--strict-keys]   # валидация по контракту
tycl syntax <file...>                            # проверка many файлов сразу
tycl fmt <conf|contract> <file...>               # каноническое форматирование
```

**Конвертация**

```bash
tycl gen <input> <output> <json|yaml|toml|tycl>  # экспорт в другой формат
tycl contract <input> <output> <dynamic|flexible|strict>  # построить контракт
```

`gen` умеет проверять по контракту перед конвертацией: `tycl gen app.tycl app.json json --contract=schema.tycl`.

**Чтение и правка значений**

```bash
tycl get <config> <path>                         # прочитать значение
tycl query <config> [path...]                    # прочитать несколько или все пути
tycl set <config> <path> <type|auto> <value>     # записать значение
tycl remove <config> <path>                      # удалить ключ
tycl structure <config>                          # список всех доступных путей
```

Пути адресуют вложенные значения и элементы массивов:

```bash
tycl get app.tycl server.port         # вложенный объект
tycl get app.tycl servers.0.host      # элемент массива объектов
tycl set app.tycl port auto 9090      # тип определяется автоматически
tycl set app.tycl timeout int null    # типизированный null
tycl set app.tycl ports ints 80,443   # замена массива
tycl set app.tycl servers.0.port int 8081  # правка элемента массива
```

`set` перезаписывает файл в каноническом виде и умеет менять тип ключа, не оставляя старого
дубликата. Неизвестные пути отклоняются с подсказкой «did you mean».

**Комбинирование**

```bash
tycl merge <base> <override> [more...]
```

Поздние файлы побеждают. Вложенные объекты сливаются по ключам, поэтому override описывает
только то, что меняет.

**Инспекция**

```bash
tycl ast <file> [config|contract]   # дерево синтаксиса в JSON
tycl docs <config>                  # рендер документирующих комментариев
tycl types                          # система типов
```

`tycl ast` — машинный контракт языка: отдаёт каждый узел с типом, исходным текстом и позицией.
На этом строятся редакторы и сторонние инструменты.

**Полный пример**

```bash
# проверить, затем экспортировать
tycl valid app.tycl schema.tycl
tycl gen app.tycl app.json json --contract=schema.tycl

# прочитать значение в скрипте
port=$(tycl get app.tycl server.port)

# отредактировать без редактора
tycl set app.tycl server.host string 127.0.0.1
tycl fmt conf app.tycl

# объединить базу с override окружения
tycl merge base.tycl prod.tycl > merged.tycl
```

---

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
  Извлекает значение из любой части конфига (синтаксис пути - `[объект основного конфига].[его вложенный объект].[...]`, `[ключ основного конфига]`) по точечному пути (например, `"database.host"`) и приводит его к указанному типу. Путь может проходить через вложенные объекты. Это позволяет переиспользовать значения в разных частях конфига.

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

  ```tycl
    {
        ports: ints = [8080, 8081, 8082],
        firstPort: int = get("ports.0", "int")          // 8080
        lastPort: int = get("ports.-1", "int")          // 8082
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
cfg, _ := tycl.Process(`{ port: int = 8080, host: string = "localhost" }`, "", false)
port := cfg.IntV["port"]        // 8080 (int)
host := cfg.StringV["host"]     // "localhost" (string)
```

---

## Диагностика

Любая ошибка возвращается плоским списком диагностик: у каждой есть стабильный код,
сообщение, позиция в исходнике и подсказка, что делать.

```go
cfg, err := tycl.ProcessSource("app.tycl", code, contract, false)
if err != nil {
    for _, d := range tycl.Diagnostics("app.tycl", code, err) {
        fmt.Println(d.Code, d.Span, d.Message, d.Hint)
    }
}
```

Если нужен только список — берите `tycl.Validate`:

```go
problems := tycl.Validate("app.tycl", code, contract, false)
```

Диагностика выглядит так:

```go
type Diagnostic struct {
    Code     string  // "type.invalid", "contract.violation", "syntax.parse", ...
    Severity string  // "error" или "warning"
    Message  string
    Hint     string  // что делать
    Span     *Span   // файл, строка, колонка, ширина
    Path     string  // путь конфига, где проблема
    Index    *int    // индекс массива
}
```

`diag.Render` форматирует диагностики для терминала, вместе со строкой исходника и кареткой
под ошибочным токеном.

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

TYCL работает как **промежуточный язык**: вы пишете безопасные и читаемые конфиги, а затем
экспортируете их для интеграции с другими системами.

```bash
tycl gen config.tycl out.json json
tycl gen config.tycl out.yaml yaml
tycl gen config.tycl out.toml toml
cat config.tycl | tycl gen - - json | jq .port
```

Полный набор команд — в разделе [Команды](#команды).

---

## Генерация контрактов

TYCL умеет автоматически строить контракт по существующему конфигу:

```bash
tycl contract config.tycl contract.tycl strict
```

Это полезно, когда конфиг уже есть, а нужна схема для валидации будущих изменений. Та же
операция доступна из Go через `generation.ContractFromConfig`.

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