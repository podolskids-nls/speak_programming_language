// Пакет runner — общий пайплайн: исходник → лексер → парсер → интерпретатор.
package runner

import (
	"bytes"
	"io"
	"os"

	"speak/interpreter"
	"speak/lexer"
	"speak/parser"
)

// Run выполняет исходный код Speak и пишет print в w. При ошибке возвращает error.
func Run(source string, w io.Writer) error {
	tokens := lexer.New(source).Tokenize()
	prog, err := parser.New(tokens).ParseProgram()
	if err != nil {
		return err
	}

	interp := interpreter.New()
	if w != nil {
		interp.SetOutput(w)
	}
	return interp.Run(prog)
}

// RunCapture выполняет код и возвращает захваченный stdout и ошибку (для тестов).
func RunCapture(source string) (string, error) {
	var buf bytes.Buffer
	err := Run(source, &buf)
	return buf.String(), err
}

// RunFile читает файл и выполняет его, выводя в os.Stdout.
func RunFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return Run(string(data), os.Stdout)
}
