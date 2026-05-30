// Тесты интерпретатора Speak (полный пайплайн).
package tests

import (
	"strings"
	"testing"

	"speak/runner"
)

// runProgram выполняет исходник и возвращает вывод stdout.
func runProgram(t *testing.T, source string) string {
	t.Helper()
	out, err := runner.RunCapture(source)
	if err != nil {
		t.Fatalf("execution error: %v", err)
	}
	return out
}

// runProgramExpectError выполняет код и ожидает ошибку.
func runProgramExpectError(t *testing.T, source string) error {
	t.Helper()
	_, err := runner.RunCapture(source)
	return err
}

func TestHelloWorld(t *testing.T) {
	out := runProgram(t, `print "Hello, World!"`)
	if strings.TrimSpace(out) != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", out)
	}
}

func TestArithmetic(t *testing.T) {
	source := "set x to 2 plus 3 times 4\nprint x\n"
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "14" {
		t.Errorf("expected 14, got %q", out)
	}
}

func TestVariables(t *testing.T) {
	source := "set x to 42\nprint x\n"
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "42" {
		t.Errorf("expected 42, got %q", out)
	}
}

func TestStringConcat(t *testing.T) {
	source := `set s to "Hello " plus "World"
print s
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", out)
	}
}

func TestIfTrue(t *testing.T) {
	source := `set x to 10
if x is greater than 5:
    print "yes"
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "yes" {
		t.Errorf("expected 'yes', got %q", out)
	}
}

func TestIfFalse(t *testing.T) {
	source := `set x to 3
if x is greater than 5:
    print "yes"
print "done"
`
	out := runProgram(t, source)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 || lines[0] != "done" {
		t.Errorf("expected only 'done', got %q", out)
	}
}

func TestIfElse(t *testing.T) {
	source := `set x to 3
if x is greater than 5:
    print "big"
else:
    print "small"
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "small" {
		t.Errorf("expected 'small', got %q", out)
	}
}

func TestRepeatLoop(t *testing.T) {
	source := `repeat 3 times:
    print "hi"
`
	out := runProgram(t, source)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), out)
	}
}

func TestWhileLoop(t *testing.T) {
	source := `set i to 0
while i is less than 3:
    print i
    set i to i plus 1
`
	out := runProgram(t, source)
	expected := "0\n1\n2\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestFunction(t *testing.T) {
	source := `define greet with name:
    print "Hi " plus name

call greet with "Ann"
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "Hi Ann" {
		t.Errorf("expected 'Hi Ann', got %q", out)
	}
}

func TestFunctionReturn(t *testing.T) {
	source := `define double with n:
    return n times 2

set r to call double with 5
print r
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "10" {
		t.Errorf("expected 10, got %q", out)
	}
}

func TestRecursion(t *testing.T) {
	source := `define factorial with n:
    if n is equal to 0:
        return 1
    return n times call factorial with n minus 1

set r to call factorial with 5
print r
`
	out := runProgram(t, source)
	if strings.TrimSpace(out) != "120" {
		t.Errorf("expected 120, got %q", out)
	}
}

func TestUnknownVariable(t *testing.T) {
	err := runProgramExpectError(t, "print x\n")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Unknown variable") {
		t.Errorf("expected 'Unknown variable' in error, got: %v", err)
	}
}

func TestDivisionByZero(t *testing.T) {
	err := runProgramExpectError(t, "set x to 10 divided by 0\n")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "divide by zero") {
		t.Errorf("expected 'divide by zero' in error, got: %v", err)
	}
}
