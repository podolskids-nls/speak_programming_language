// Точка входа интерпретатора Speak.
// Запуск: go run main.go examples/hello.speak
package main

import (
	"fmt"
	"os"

	"speak/runner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Использование: go run main.go <файл.speak>")
		os.Exit(1)
	}

	path := os.Args[1]
	if err := runner.RunFile(path); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
