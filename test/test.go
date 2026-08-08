package main

import (
	"fmt"

	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/shared"
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
    test4: objects = [
        {
            key = 'test',
        },
        {
            key = 'test2',
        },
    ],
    test5: objects = [
        {
            key = 'test',
        },
        {
            key = 0,
        },
    ],
}
` // formatted
	contcode := `
flexible {
    test: object = strict {
        key: string,
    },
    test1: objects = flexible {
        key: string
    }
}
` // formatted

	conf, err := tycl.Process(confcode, contcode)
	fmt.Println(err, conf)

	fmt.Println(confcode)
	fmt.Println(generation.Json(conf))
	fmt.Println(generation.Yaml(conf))
	fmt.Println(generation.Toml(conf))
	fmt.Println(generation.Tycl(conf))

	contract, err := generation.ContractFromConfig(conf, shared.ContractStrict)
	fmt.Println(err)
	code, err := generation.GenerateContractCode(contract)
	fmt.Println(code, err)
}

/*
macbook@MacBook-Pro tycl % go run ./test
<nil> &{map[] map[] map[] map[] map[test3:int] map[] map[] map[] map[] map[test:0xc0000c1ab0 test2:0xc0002c8e00] map[test1:[0xc0002dff10 0xc0002f70a0] test4:[0xc000337340 0xc00034c4d0] test5:[0xc0003615e0 0xc000382770]] }

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
    test4: objects = [
        {
            key = 'test',
        },
        {
            key = 'test2',
        },
    ],
    test5: objects = [
        {
            key = 'test',
        },
        {
            key = 0,
        },
    ],
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
      "key": "test2",
      "test": {
        "int": 0
      }
    }
  ],
  "test2": {
    "key": "test",
    "test": 0
  },
  "test3": null,
  "test4": [
    {
      "key": "test"
    },
    {
      "key": "test2"
    }
  ],
  "test5": [
    {
      "key": "test"
    },
    {
      "key": 0
    }
  ]
} <nil>
test:
    key: '"''test'
test1:
    - key: test
    - a: true
      key: test2
      test:
        int: 0
test2:
    key: test
    test: 0
test3: null
test4:
    - key: test
    - key: test2
test5:
    - key: test
    - key: 0
 <nil>
[test]
  key = "\"'test"

[[test1]]
  key = "test"

[[test1]]
  a = true
  key = "test2"
  [test1.test]
    int = 0

[test2]
  key = "test"
  test = 0

[[test4]]
  key = "test"

[[test4]]
  key = "test2"

[[test5]]
  key = "test"

[[test5]]
  key = 0
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
            test: object = {
                int: int = 0,
            },
        }
    ],
    test4: objects = [
        {
            key: string = "test",
        },
        {
            key: string = "test2",
        }
    ],
    test5: objects = [
        {
            key: string = "test",
        },
        {
            key: int = 0,
        }
    ],
} <nil>
<nil>
strict {
    test3: int,
    test2: object = strict {
        test: int,
        key: string,
    },
    test: object = strict {
        key: string,
    },
    test1: objects,
    test4: objects = strict {
        key: string,
    },
    test5: objects,
} <nil>
macbook@MacBook-Pro tycl %
*/
