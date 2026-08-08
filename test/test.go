package main

import (
	"fmt"

	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/generation"
)

func main() {
	confcode := `
{
    test = {
        key = "\"'test",
    },
    test2: object = {
        key = 'test',
        test = 0,
    },
    test1: objects = [
        {
            key = 'test',
        },
        {
            a = true,
            key = 'test2',
            test = {
                int = 0,
            },
        },
    ],
    test3: int = null,
}
` // formatted
	contcode := `
flexible {
    test: object = strict {
        key: string,
    },
}
` // formatted

	conf, err := tycl.Process(confcode, contcode)
	fmt.Println(err, conf)

	fmt.Println(confcode)
	fmt.Println(generation.Json(conf))
	fmt.Println(generation.Yaml(conf))
	fmt.Println(generation.Toml(conf))
	fmt.Println(generation.Tycl(conf))
}

/*
macbook@MacBook-Pro tycl % go run ./test
<nil> &{map[] map[] map[] map[] map[test3:int] map[] map[] map[] map[] map[test:0xc000330bd0 test2:0xc000345ce0] map[test1:[0xc000360e70 0xc000390000]] }

{
    test = {
        key = "\"'test",
    },
    test2: object = {
        key = 'test',
        test = 0,
    },
    test1: objects = [
        {
            key = 'test',
        },
        {
            a = true,
            key = 'test2',
        }
    ],
    test3: int = null,
}

{
  "test": {
    "key": "\"'test"
  },
  "test1": [
    {
      "key": "test"
    },
    {
      "a": true,
      "key": "test2"
    }
  ],
  "test2": {
    "key": "test",
    "test": 0
  },
  "test3": null
} <nil>
test:
    key: '"''test'
test1:
    - key: test
    - a: true
      key: test2
test2:
    key: test
    test: 0
test3: null
 <nil>
[test]
  key = "\"'test"

[[test1]]
  key = "test"

[[test1]]
  a = true
  key = "test2"

[test2]
  key = "test"
  test = 0
 <nil>
{
    test3: int = null,
    test: object = {
        key: string = "\"'test",
    },
    test2: object = {
        test: int = 0,
        key: string = "test",
    },
    test1: objects = [
        {
            key: string = "test",
        },
        {
            a: bool = true,
            key: string = "test2",
        }
    ],
} <nil>
macbook@MacBook-Pro tycl %
*/
