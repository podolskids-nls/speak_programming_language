// Пакет interpreter — пошаговое выполнение узлов AST.
package interpreter

import (
	"fmt"
	"io"
	"math"
	"os"

	"speak/errors"
	"speak/parser"
)

// FunctionValue — значение пользовательской функции в среде выполнения.
type FunctionValue struct {
	Name    string
	Param   string
	Body    []parser.Node
	Closure *Environment // замыкание: среда на момент define
}

// returnSignal — служебный тип для прерывания тела функции при return (не panic пользователю).
type returnSignal struct {
	value interface{}
}

// Interpreter выполняет программу Speak.
type Interpreter struct {
	env    *Environment
	output io.Writer // куда писать print (по умолчанию os.Stdout)
}

// New создаёт интерпретатор с глобальной средой и stdout.
func New() *Interpreter {
	return &Interpreter{
		env:    NewEnvironment(),
		output: os.Stdout,
	}
}

// SetOutput задаёт writer для print (используется в тестах).
func (interp *Interpreter) SetOutput(w io.Writer) {
	interp.output = w
}

// Run выполняет программу целиком.
func (interp *Interpreter) Run(prog *parser.Program) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case returnSignal:
				// return на верхнем уровне — игнорируем
			case errors.SpeakError:
				err = v
			default:
				err = errors.NewRuntimeError(0, fmt.Sprintf("%v", v))
			}
		}
	}()

	for _, stmt := range prog.Statements {
		interp.exec(stmt)
	}
	return nil
}

// exec выполняет один оператор.
func (interp *Interpreter) exec(node parser.Node) interface{} {
	switch n := node.(type) {
	case *parser.SetStatement:
		val := interp.eval(n.Value)
		interp.env.Assign(n.Name, val)
		return val

	case *parser.PrintStatement:
		val := interp.eval(n.Value)
		fmt.Fprintln(interp.output, formatValue(val))
		return val

	case *parser.IfStatement:
		if interp.isTruthy(interp.eval(n.Condition)) {
			interp.execBlock(n.Body)
		} else if len(n.Else) > 0 {
			interp.execBlock(n.Else)
		}
		return nil

	case *parser.WhileStatement:
		for interp.isTruthy(interp.eval(n.Condition)) {
			interp.execBlock(n.Body)
		}
		return nil

	case *parser.RepeatStatement:
		countVal := interp.eval(n.Count)
		count := toInt(countVal, n.Line)
		for i := 0; i < count; i++ {
			interp.execBlock(n.Body)
		}
		return nil

	case *parser.DefineStatement:
		fn := &FunctionValue{
			Name:    n.Name,
			Param:   n.Param,
			Body:    n.Body,
			Closure: interp.env,
		}
		interp.env.Set(n.Name, fn)
		return fn

	case *parser.CallStatement:
		interp.eval(n.Call)
		return nil

	case *parser.ReturnStatement:
		val := interp.eval(n.Value)
		panic(returnSignal{value: val})

	default:
		return interp.eval(node)
	}
}

// execBlock выполняет список операторов (без новой области видимости — переменные «утекают»).
func (interp *Interpreter) execBlock(stmts []parser.Node) {
	for _, stmt := range stmts {
		interp.exec(stmt)
	}
}

// eval вычисляет значение выражения.
func (interp *Interpreter) eval(node parser.Node) interface{} {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *parser.NumberLiteral:
		return n.Value

	case *parser.StringLiteral:
		return n.Value

	case *parser.BoolLiteral:
		return n.Value

	case *parser.Identifier:
		val, ok := interp.env.Get(n.Name)
		if !ok {
			panic(errors.NewRuntimeError(n.Line, fmt.Sprintf("Unknown variable '%s'", n.Name)))
		}
		return val

	case *parser.BinaryExpr:
		left := interp.eval(n.Left)
		right := interp.eval(n.Right)
		return interp.evalBinary(n.Line, n.Op, left, right)

	case *parser.ComparisonExpr:
		return interp.evalComparison(n)

	case *parser.CallExpression:
		argVal := interp.eval(n.Arg)
		fnVal, _ := interp.lookupFunction(n.Name, n.Line)
		return interp.callFunction(fnVal, argVal, n.Line)

	default:
		panic(errors.NewRuntimeError(node.GetLine(), "Cannot evaluate node"))
	}
}

// lookupFunction находит функцию по имени.
func (interp *Interpreter) lookupFunction(name string, line int) (*FunctionValue, bool) {
	val, ok := interp.env.Get(name)
	if !ok {
		panic(errors.NewRuntimeError(line, fmt.Sprintf("Unknown function '%s'", name)))
	}
	fn, ok := val.(*FunctionValue)
	if !ok {
		panic(errors.NewRuntimeError(line, fmt.Sprintf("'%s' is not a function", name)))
	}
	return fn, true
}

// callFunction вызывает пользовательскую функцию с одним аргументом.
func (interp *Interpreter) callFunction(fn *FunctionValue, arg interface{}, line int) interface{} {
	callEnv := NewEnclosed(fn.Closure)
	callEnv.Set(fn.Param, arg)

	prevEnv := interp.env
	interp.env = callEnv

	var result interface{} = nil
	func() {
		defer func() {
			if r := recover(); r != nil {
				if ret, ok := r.(returnSignal); ok {
					result = ret.value
				} else {
					panic(r)
				}
			}
		}()
		for _, stmt := range fn.Body {
			interp.exec(stmt)
		}
	}()

	interp.env = prevEnv
	return result
}

// evalBinary вычисляет бинарную операцию.
func (interp *Interpreter) evalBinary(line int, op string, left, right interface{}) interface{} {
	switch op {
	case "plus":
		return interp.add(line, left, right)
	case "minus":
		l, r := asNumbers(line, left, right, op)
		return l - r
	case "times":
		l, r := asNumbers(line, left, right, op)
		return l * r
	case "divided":
		l, r := asNumbers(line, left, right, op)
		if r == 0 {
			panic(errors.NewRuntimeError(line, "Cannot divide by zero"))
		}
		return l / r
	default:
		panic(errors.NewRuntimeError(line, "Unknown operator '"+op+"'"))
	}
}

// add — сложение чисел или конкатенация строк.
func (interp *Interpreter) add(line int, left, right interface{}) interface{} {
	ls, lok := left.(string)
	rs, rok := right.(string)
	if lok && rok {
		return ls + rs
	}
	if lok {
		return ls + formatValue(right)
	}
	if rok {
		return formatValue(left) + rs
	}
	l, r := asNumbers(line, left, right, "plus")
	return l + r
}

// evalComparison вычисляет условие сравнения.
func (interp *Interpreter) evalComparison(n *parser.ComparisonExpr) bool {
	left := interp.eval(n.Left)

	switch n.Op {
	case "istrue":
		b, ok := left.(bool)
		if !ok {
			panic(errors.NewTypeError(n.Line, "Expected boolean for 'is true'"))
		}
		return b

	case "isfalse":
		b, ok := left.(bool)
		if !ok {
			panic(errors.NewTypeError(n.Line, "Expected boolean for 'is false'"))
		}
		return !b

	case "greater", "less", "equal":
		right := interp.eval(n.Right)
		return compareValues(n.Line, n.Op, left, right)
	default:
		panic(errors.NewRuntimeError(n.Line, "Unknown comparison"))
	}
}

// compareValues сравнивает два значения совместимых типов.
func compareValues(line int, op string, left, right interface{}) bool {
	switch op {
	case "equal":
		ls, lok := left.(string)
		rs, rok := right.(string)
		if lok && rok {
			return ls == rs
		}
		ln, lnum := left.(float64)
		rn, rnum := right.(float64)
		if lnum && rnum {
			return ln == rn
		}
		if lb, lbok := left.(bool); lbok {
			if rb, rbok := right.(bool); rbok {
				return lb == rb
			}
		}
		panic(errors.NewTypeError(line, "Cannot compare values of different types"))

	case "greater", "less":
		ln, lok := left.(float64)
		rn, rok := right.(float64)
		if !lok || !rok {
			panic(errors.NewTypeError(line, "Comparison requires numbers"))
		}
		if op == "greater" {
			return ln > rn
		}
		return ln < rn
	}
	return false
}

// asNumbers приводит оба операнда к float64 или выдаёт TypeError.
func asNumbers(line int, left, right interface{}, op string) (float64, float64) {
	ln, lok := left.(float64)
	rn, rok := right.(float64)
	if !lok || !rok {
		panic(errors.NewTypeError(line, fmt.Sprintf("Cannot use '%s' with %s and %s",
			op, typeName(left), typeName(right))))
	}
	return ln, rn
}

// typeName — имя типа для сообщений об ошибках.
func typeName(v interface{}) string {
	switch v.(type) {
	case float64:
		return "number"
	case string:
		return "string"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "value"
	}
}

// isTruthy проверяет, истинно ли значение условия.
func (interp *Interpreter) isTruthy(val interface{}) bool {
	if b, ok := val.(bool); ok {
		return b
	}
	panic(errors.NewTypeError(0, "Condition must be boolean"))
}

// toInt преобразует число повторений в int.
func toInt(v interface{}, line int) int {
	n, ok := v.(float64)
	if !ok {
		panic(errors.NewTypeError(line, "Repeat count must be a number"))
	}
	return int(math.Round(n))
}

// formatValue преобразует значение в строку для print.
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case float64:
		if val == math.Trunc(val) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", val)
	}
}
