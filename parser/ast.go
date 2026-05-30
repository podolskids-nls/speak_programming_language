// Пакет parser — синтаксический анализ: построение AST из потока токенов.
package parser

// Node — общий интерфейс для всех узлов абстрактного синтаксического дерева.
type Node interface {
	nodeType() string
	GetLine() int
}

// Program — корень AST: последовательность операторов верхнего уровня.
type Program struct {
	Line       int
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }
func (p *Program) GetLine() int     { return p.Line }

// SetStatement — присваивание: set x to expression
type SetStatement struct {
	Line  int
	Name  string
	Value Node
}

func (s *SetStatement) nodeType() string { return "SetStatement" }
func (s *SetStatement) GetLine() int     { return s.Line }

// PrintStatement — вывод: print expression
type PrintStatement struct {
	Line  int
	Value Node
}

func (s *PrintStatement) nodeType() string { return "PrintStatement" }
func (s *PrintStatement) GetLine() int     { return s.Line }

// IfStatement — условие с необязательной веткой else
type IfStatement struct {
	Line      int
	Condition Node
	Body      []Node
	Else      []Node
}

func (s *IfStatement) nodeType() string { return "IfStatement" }
func (s *IfStatement) GetLine() int     { return s.Line }

// WhileStatement — цикл while condition: body
type WhileStatement struct {
	Line      int
	Condition Node
	Body      []Node
}

func (s *WhileStatement) nodeType() string { return "WhileStatement" }
func (s *WhileStatement) GetLine() int     { return s.Line }

// RepeatStatement — цикл repeat N times: body
type RepeatStatement struct {
	Line  int
	Count Node
	Body  []Node
}

func (s *RepeatStatement) nodeType() string { return "RepeatStatement" }
func (s *RepeatStatement) GetLine() int   { return s.Line }

// DefineStatement — объявление функции: define name with param: body
type DefineStatement struct {
	Line  int
	Name  string
	Param string
	Body  []Node
}

func (s *DefineStatement) nodeType() string { return "DefineStatement" }
func (s *DefineStatement) GetLine() int     { return s.Line }

// CallStatement — вызов функции как оператор: call name with arg
type CallStatement struct {
	Line int
	Call *CallExpression
}

func (s *CallStatement) nodeType() string { return "CallStatement" }
func (s *CallStatement) GetLine() int     { return s.Line }

// CallExpression — вызов функции как выражение: call name with arg
type CallExpression struct {
	Line int
	Name string
	Arg  Node
}

func (e *CallExpression) nodeType() string { return "CallExpression" }
func (e *CallExpression) GetLine() int     { return e.Line }

// ReturnStatement — return expression
type ReturnStatement struct {
	Line  int
	Value Node
}

func (s *ReturnStatement) nodeType() string { return "ReturnStatement" }
func (s *ReturnStatement) GetLine() int     { return s.Line }

// BinaryExpr — бинарная операция: left op right
type BinaryExpr struct {
	Line  int
	Op    string // plus, minus, times, divided
	Left  Node
	Right Node
}

func (e *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (e *BinaryExpr) GetLine() int     { return e.Line }

// ComparisonExpr — сравнение: left is greater/less/equal ... right или left is true/false
type ComparisonExpr struct {
	Line  int
	Op    string // greater, less, equal, istrue, isfalse
	Left  Node
	Right Node // nil для is true / is false
}

func (e *ComparisonExpr) nodeType() string { return "ComparisonExpr" }
func (e *ComparisonExpr) GetLine() int     { return e.Line }

// NumberLiteral — числовой литерал
type NumberLiteral struct {
	Line  int
	Value float64
}

func (n *NumberLiteral) nodeType() string { return "NumberLiteral" }
func (n *NumberLiteral) GetLine() int     { return n.Line }

// StringLiteral — строковый литерал
type StringLiteral struct {
	Line  int
	Value string
}

func (s *StringLiteral) nodeType() string { return "StringLiteral" }
func (s *StringLiteral) GetLine() int     { return s.Line }

// BoolLiteral — булев литерал true/false
type BoolLiteral struct {
	Line  int
	Value bool
}

func (b *BoolLiteral) nodeType() string { return "BoolLiteral" }
func (b *BoolLiteral) GetLine() int     { return b.Line }

// Identifier — ссылка на переменную
type Identifier struct {
	Line int
	Name string
}

func (i *Identifier) nodeType() string { return "Identifier" }
func (i *Identifier) GetLine() int     { return i.Line }
