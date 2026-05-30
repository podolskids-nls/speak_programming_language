// Тесты лексера Speak.
package tests

import (
	"testing"

	"speak/lexer"
)

// tokenTypes возвращает слайс типов токенов (без EOF) для удобства проверок.
func tokenTypes(tokens []lexer.Token) []lexer.TokenType {
	var types []lexer.TokenType
	for _, t := range tokens {
		if t.Type == lexer.EOF {
			break
		}
		types = append(types, t.Type)
	}
	return types
}

func TestTokenizeSet(t *testing.T) {
	tokens := lexer.New("set x to 10").Tokenize()
	types := tokenTypes(tokens)

	expected := []lexer.TokenType{lexer.SET, lexer.IDENT, lexer.TO, lexer.NUMBER}
	if len(types) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(types), types)
	}
	for i, exp := range expected {
		if types[i] != exp {
			t.Errorf("token %d: expected %s, got %s", i, exp, types[i])
		}
	}
	if tokens[1].Literal != "x" {
		t.Errorf("expected ident 'x', got %s", tokens[1].Literal)
	}
	if tokens[3].Literal != "10" {
		t.Errorf("expected number '10', got %s", tokens[3].Literal)
	}
}

func TestTokenizeString(t *testing.T) {
	tokens := lexer.New(`print "hello"`).Tokenize()
	found := false
	for _, tok := range tokens {
		if tok.Type == lexer.STRING && tok.Literal == "hello" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected STRING token with literal 'hello'")
	}
}

func TestTokenizeComment(t *testing.T) {
	tokens := lexer.New("-- comment\nset x to 1").Tokenize()
	for _, tok := range tokens {
		if tok.Literal == "comment" || tok.Type == lexer.IDENT && tok.Literal == "comment" {
			t.Fatal("comment should be skipped")
		}
	}
	types := tokenTypes(tokens)
	if len(types) < 4 || types[0] != lexer.SET {
		t.Fatalf("expected set statement after comment, got %v", types)
	}
}

func TestTokenizeIndent(t *testing.T) {
	source := "if x is greater than 5:\n    print x\n"
	tokens := lexer.New(source).Tokenize()
	types := tokenTypes(tokens)

	hasIndent := false
	hasDedent := false
	for _, ty := range types {
		if ty == lexer.INDENT {
			hasIndent = true
		}
		if ty == lexer.DEDENT {
			hasDedent = true
		}
	}
	if !hasIndent {
		t.Error("expected INDENT token")
	}
	if !hasDedent {
		t.Error("expected DEDENT token")
	}
}
