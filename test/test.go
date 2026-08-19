package main

import (
	"fmt"

	"github.com/pt-main/tycl"
	"github.com/pt-main/tycl/format"
	"github.com/pt-main/tycl/generation"
	"github.com/pt-main/tycl/shared"
)

func main() {
	conf := `
{
	/* doc 1 */ 
	test: ints = [],
	/* doc 2 */
}`

	contract := ``

	cfg, err := tycl.Process(conf, contract, false)
	if err != nil {
		fmt.Println(format.FormatError(err))
	}

	fmt.Println(cfg.IntV["port"])    // 8080
	fmt.Println(cfg.StringV["host"]) // "localhost"

	// Экспорт в JSON
	fmt.Println(generation.Json(cfg))
	// Экспорт в TOML
	fmt.Println(generation.Toml(cfg))

	// Генерация контракта из конфига
	cont, _ := generation.ContractFromConfig(cfg, shared.ContractStrict)
	contCode, _ := generation.GenerateContractCode(cont)
	fmt.Println(contCode)
}
