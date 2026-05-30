// Тесты парсера Speak.
package tests

import (
	"testing"

	"speak/lexer"
	"speak/parser"
)

func TestParseSetPrint(t *testing.T) {
	source := "set x to 10\nprint x\n"
	tokens := lexer.New(source).Tokenize()
	prog, err := parser.New(tokens).ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
}

func TestParseIf(t *testing.T) {
	source := "if x is greater than 5:\n    print x\n"
	tokens := lexer.New(source).Tokenize()
	_, err := parser.New(tokens).ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
}
