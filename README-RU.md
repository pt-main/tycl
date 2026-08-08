# TYCL — Typed Config Language

[![Go Reference](https://pkg.go.dev/badge/github.com/pt-main/tycl.svg)](https://pkg.go.dev/github.com/pt-main/tycl)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-yellow.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/pt-main/tycl)](https://github.com/pt-main/tycl/releases)

```bash
go get github.com/pt-main/tycl
```

**TYCL** — это типизированный язык конфигурации для Go.  
Он даёт строгую типизацию, контракты (схемы) и читаемый синтаксис — без генерации кода и без `interface{}`.

---

## Зачем TYCL?

| Формат | Проблемы | TYCL решает |
|--------|----------|-------------|
| **JSON** | нет комментариев, всё через `interface{}`, нет валидации | строгая типизация, комментарии, контракты |
| **YAML** | чувствителен к пробелам, нет типизации | явные типы, детерминированный парсинг |
| **TOML** | ограничен, нет схем | контракты, гибкие структуры |

TYCL даёт **80% возможностей сложных языков** (CUE, Dhall) при **20% сложности**.  
Он подходит как **основной формат** для конфигов и как **промежуточное представление** — вы можете писать на TYCL, а затем генерировать JSON, YAML или TOML для интеграции с другими системами.

---

## Установка

### Как библиотека (Go‑модуль)

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

### CLI (утилита командной строки)

Скачайте бинарник из [релизов](https://github.com/pt-main/tycl/releases) или установите через `go install`:

```bash
go install github.com/pt-main/tycl/tycl@latest
```

Команды:

- `tycl valid <config> [contract]` — проверка конфига по контракту.
- `tycl syntax <file...>` — проверка синтаксиса и типов (без контракта).
- `tycl fmt <type> <file...>` — форматирование (`conf` / `contract`).
- `tycl gen <input> <output> <json|yaml|toml>` — генерация целевого формата.

---

## Синтаксис

TYCL близок к JSON, но каждое поле имеет явный тип.

### Базовые типы

```tycl
port: int = 8080
rate: float = 1.5
debug: bool = true
host: string = "localhost"
```

### Null‑значения

TYCL позволяет указать, что поле существует, но его значение отсутствует, при этом **тип поля известен**:

```tycl
timeout: int = null   /* поле timeout существует, но равно null, тип int */
```

**Правила работы с null:**

1. **Тип обязателен** — `null` всегда сопровождается типом (`int`, `string` и т.д.).
2. **Уникальность null** — для одного имени может быть **только одно** null‑значение.  
   Если вы объявили `timeout: int = null`, то нельзя добавить `timeout: string = null`, но можно добавить `timeout: string = "5s"` (не null).
3. **Null ≠ отсутствие поля** — `key: int = null` означает, что поле есть, но его значение не задано. Если поле вовсе не указано в конфиге, оно просто отсутствует (и контракт это заметит).

### Массивы

Массивы строго типизированы и обозначаются **типом во множественном числе**. Тип массива **обязателен**:

```tycl
ports: ints = [8080, 8081]
names: strings = ["dev", "prod"]
rates: floats = [1.1, 2.2]
flags: bools = [true, false]
servers: objects = [
    { host: string = "a", port: int = 80 },
    { host: string = "b", port: int = 443 }
]
```

Все элементы массива должны быть одного типа.

### Объекты (вложенные)

```tycl
server: object = {
    host: string = "127.0.0.1"
    port: int = 8080
    timeout: int = null
}
```

Объекты могут быть вложены произвольно:

```tycl
app: object = {
    name: string = "myapp"
    database: object = {
        host: string = "localhost"
        port: int = 5432
    }
}
```

### Комментарии

Поддерживаются **только блочные комментарии** `/* ... */`, и они могут располагаться строго **в начале или в конце блока** (объекта или массива). Это делает комментарии документацией к блоку.

```tycl
/* Этот объект описывает сервер */
server: object = {
    host: string = "127.0.0.1"
    port: int = 8080
} /* Конец описания сервера */
```

Однострочные комментарии (`//`) **не поддерживаются**.

### Важные правила

1. **Необязательность типов в ключах**  
   Если тип не указан, он выводится из значения:  
   `key = "text"` ->`string`, `key = 42` -> `int`.  
   **Исключение:** для массивов тип **обязателен** (`ints`, `strings` и т.д.).

2. **Дублирование ключей с разными типами**  
   Разрешено, но **не рекомендуется**, потому что при генерации в JSON/YAML/TOML конфликт приведёт к ошибке.  
   Пример:  
   ```tycl
   port: int = 8080
   port: string = "8080"   /* допустимо, но плохая практика */
   ```

3. **Null**  
   Может быть только один ключ с данным именем, если он равен `null` (независимо от типа).  
   ```tycl
   timeout: int = null     /* ок */
   timeout: string = null  /* ошибка: уже есть null для timeout */
   timeout: string = "5s"  /* ок, это не null */
   ```

---

## Структура Config (после парсинга)

Функция `tycl.Process` возвращает объект `*Config`, который содержит отдельные мапы для каждого типа данных. Это позволяет обращаться к значениям **без приведения типов**:

```go
type Config struct {
    IntV    map[string]int
    FloatV  map[string]float64
    BoolV   map[string]bool
    StringV map[string]string
    NullV   map[string]string   // ключ → тип null-значения

    IntArrV    map[string][]int
    FloatArrV  map[string][]float64
    BoolArrV   map[string][]bool
    StringArrV map[string][]string

    InnerV    map[string]*Config       // объекты
    InnerArrV map[string][]*Config     // массивы объектов
}
```

Пример доступа:

```go
cfg, _ := tycl.Process(`{ port: int = 8080, host: string = "localhost" }`, "")
port := cfg.IntV["port"]       // 8080 (int)
host := cfg.StringV["host"]    // "localhost" (string)
```

---

## Генерация других форматов из Go

Пакет `generation` позволяет экспортировать `*Config` в JSON, YAML, TOML и обратно в TYCL:

```go
import "github.com/pt-main/tycl/generation"

jsonStr, err := generation.Json(cfg)
yamlStr, err := generation.Yaml(cfg)
tomlStr, err := generation.Toml(cfg)
tyclStr, err := generation.Tycl(cfg)   // обратно в TYCL
```

Это полезно, если вы загрузили конфиг, изменили его в коде и хотите сохранить в другом формате.

---

## Контракты (схемы)

Контракт описывает ожидаемую структуру конфига. Он пишется на том же языке, но вместо значений указываются только типы.

```tycl
strict {
    port: int
    host: string
    debug: bool
    timeout: int         
    ports: ints
    server: object = strict {
        host: string
        port: int
    }
}
```

**Уровни строгости:**

- `dynamic` — проверка не выполняется (любая структура).
- `flexible` — все перечисленные поля должны присутствовать, лишние разрешены.
- `strict` — точное соответствие (нельзя добавлять лишние поля).

Контракты могут быть вложенными для объектов. Для массивов объектов контракт не поддерживается — проверяется только наличие ключа.

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

## Интеграция в Go (полный пример)

```go
package main

import (
    "fmt"
    "log"
    "github.com/pt-main/tycl"
    "github.com/pt-main/tycl/generation"
)

func main() {
    conf := `{
        port: int = 8080
        host: string = "localhost"
        timeout: int = null
    }`

    contract := `strict {
        port: int
        host: string
        timeout: int
    }`

    cfg, err := tycl.Process(conf, contract)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(cfg.IntV["port"])      // 8080
    fmt.Println(cfg.StringV["host"])   // "localhost"

    // Экспорт в JSON
    jsonData, _ := generation.Json(cfg)
    fmt.Println(jsonData)
}
```

Если контракт не нужен, передайте `""` или `"dynamic{}"` — проверка будет пропущена.

---

## Планы на будущее

- **Плагин для VS Code:** подсветка синтаксиса, автодополнение, форматирование, проверка контрактов на лету.

---

## Лицензия

Apache 2.0 — подробности в [LICENSE](LICENSE).

---

**TYCL** — создан для удобства, безопасности и простоты. Попробуйте — и вы не захотите возвращаться к JSON.