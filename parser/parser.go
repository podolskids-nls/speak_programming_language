// Пакет parser — рекурсивный спуск: построение AST из токенов лексера.
package parser

import (
	"strconv"

	"speak/errors"
	"speak/lexer"
)

// Parser разбирает поток токенов в AST.
type Parser struct {
	tokens  []lexer.Token
	pos     int
	curTok  lexer.Token
	peekTok lexer.Token
}

// New создаёт парсер для готового списка токенов.
func New(tokens []lexer.Token) *Parser {
	p := &Parser{tokens: tokens}
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken продвигает текущий и «заглядывающий» токены.
func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	if p.pos < len(p.tokens) {
		p.peekTok = p.tokens[p.pos]
		p.pos++
	} else {
		p.peekTok = lexer.Token{Type: lexer.EOF, Line: p.curTok.Line}
	}
}

// ParseProgram разбирает всю программу.
func (p *Parser) ParseProgram() (*Program, error) {
	prog := &Program{Line: 1}
	p.skipNewlines()

	for p.curTok.Type != lexer.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
		p.skipNewlines()
	}
	return prog, nil
}

// skipNewlines пропускает пустые переводы строк.
func (p *Parser) skipNewlines() {
	for p.curTok.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

// parseStatement разбирает один оператор верхнего уровня.
func (p *Parser) parseStatement() (Node, error) {
	line := p.curTok.Line

	switch p.curTok.Type {
	case lexer.SET:
		return p.parseSetStatement()
	case lexer.PRINT:
		return p.parsePrintStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.REPEAT:
		return p.parseRepeatStatement()
	case lexer.DEFINE:
		return p.parseDefineStatement()
	case lexer.CALL:
		return p.parseCallStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.NEWLINE:
		p.nextToken()
		return nil, nil
	default:
		return nil, errors.NewParseError(line, "Unexpected token '"+string(p.curTok.Type)+"'")
	}
}

// parseSetStatement: set name to expression
func (p *Parser) parseSetStatement() (Node, error) {
	line := p.curTok.Line
	if p.curTok.Type != lexer.SET {
		return nil, errors.NewParseError(line, "Expected 'set'")
	}
	p.nextToken()

	if p.curTok.Type != lexer.IDENT {
		return nil, errors.NewParseError(line, "Expected variable name after 'set'")
	}
	name := p.curTok.Literal
	p.nextToken()

	if p.curTok.Type != lexer.TO {
		return nil, errors.NewParseError(line, "Expected 'to' after variable name")
	}
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errors.NewParseError(line, "Expected expression after 'to'")
	}

	p.skipStatementEnd()
	return &SetStatement{Line: line, Name: name, Value: value}, nil
}

// parsePrintStatement: print expression
func (p *Parser) parsePrintStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errors.NewParseError(line, "Expected expression after 'print'")
	}

	p.skipStatementEnd()
	return &PrintStatement{Line: line, Value: value}, nil
}

// parseIfStatement: if condition: body [else: body]
func (p *Parser) parseIfStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	stmt := &IfStatement{Line: line, Condition: cond, Body: body}

	p.skipNewlines()
	if p.curTok.Type == lexer.ELSE {
		p.nextToken()
		if p.curTok.Type != lexer.COLON {
			return nil, errors.NewParseError(p.curTok.Line, "Expected ':' after 'else'")
		}
		p.nextToken()
		elseBody, err := p.parseBlockBody()
		if err != nil {
			return nil, err
		}
		stmt.Else = elseBody
	}

	return stmt, nil
}

// parseWhileStatement: while condition: body
func (p *Parser) parseWhileStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	cond, err := p.parseCondition()
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &WhileStatement{Line: line, Condition: cond, Body: body}, nil
}

// parseRepeatStatement: repeat count times: body
func (p *Parser) parseRepeatStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	count, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != lexer.TIMES {
		return nil, errors.NewParseError(line, "Expected 'times' after repeat count")
	}
	p.nextToken()

	if p.curTok.Type != lexer.COLON {
		return nil, errors.NewParseError(line, "Expected ':' after 'times'")
	}
	p.nextToken()

	body, err := p.parseBlockBody()
	if err != nil {
		return nil, err
	}

	return &RepeatStatement{Line: line, Count: count, Body: body}, nil
}

// parseDefineStatement: define name with param: body
func (p *Parser) parseDefineStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	if p.curTok.Type != lexer.IDENT {
		return nil, errors.NewParseError(line, "Expected function name after 'define'")
	}
	name := p.curTok.Literal
	p.nextToken()

	if p.curTok.Type != lexer.WITH {
		return nil, errors.NewParseError(line, "Expected 'with' after function name")
	}
	p.nextToken()

	if p.curTok.Type != lexer.IDENT {
		return nil, errors.NewParseError(line, "Expected parameter name after 'with'")
	}
	param := p.curTok.Literal
	p.nextToken()

	if p.curTok.Type != lexer.COLON {
		return nil, errors.NewParseError(line, "Expected ':' after parameter")
	}
	p.nextToken()

	body, err := p.parseBlockBody()
	if err != nil {
		return nil, err
	}

	return &DefineStatement{Line: line, Name: name, Param: param, Body: body}, nil
}

// parseCallStatement: call name with arg (как оператор)
func (p *Parser) parseCallStatement() (Node, error) {
	line := p.curTok.Line
	call, err := p.parseCallExpression()
	if err != nil {
		return nil, err
	}
	p.skipStatementEnd()
	return &CallStatement{Line: line, Call: call}, nil
}

// parseReturnStatement: return expression
func (p *Parser) parseReturnStatement() (Node, error) {
	line := p.curTok.Line
	p.nextToken()

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.skipStatementEnd()
	return &ReturnStatement{Line: line, Value: value}, nil
}

// parseBlock разбирает блок после условия (ожидает colon).
func (p *Parser) parseBlock() ([]Node, error) {
	if p.curTok.Type != lexer.COLON {
		return nil, errors.NewParseError(p.curTok.Line, "Expected ':' after condition")
	}
	p.nextToken()
	return p.parseBlockBody()
}

// parseBlockBody разбирает INDENT ... statements ... DEDENT.
func (p *Parser) parseBlockBody() ([]Node, error) {
	p.skipNewlines()

	if p.curTok.Type != lexer.INDENT {
		return nil, errors.NewParseError(p.curTok.Line, "Expected indented block")
	}
	p.nextToken()

	var statements []Node
	for p.curTok.Type != lexer.DEDENT && p.curTok.Type != lexer.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.skipNewlines()
	}

	if p.curTok.Type != lexer.DEDENT {
		return nil, errors.NewParseError(p.curTok.Line, "Expected end of block")
	}
	p.nextToken()

	return statements, nil
}

// skipStatementEnd пропускает NEWLINE в конце оператора.
func (p *Parser) skipStatementEnd() {
	if p.curTok.Type == lexer.NEWLINE {
		p.nextToken()
	}
}

// parseExpression — выражение с plus/minus (низший приоритет).
func (p *Parser) parseExpression() (Node, error) {
	return p.parsePlusMinus()
}

// parsePlusMinus — сложение, вычитание, конкатенация строк.
func (p *Parser) parsePlusMinus() (Node, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == lexer.PLUS || p.curTok.Type == lexer.MINUS {
		op := p.curTok.Literal
		line := p.curTok.Line
		p.nextToken()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Line: line, Op: op, Left: left, Right: right}
	}
	return left, nil
}

// parseTerm — умножение и деление (times, divided by).
func (p *Parser) parseTerm() (Node, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for {
		if p.curTok.Type == lexer.TIMES {
			// «times» перед «:» — это repeat N times:, а не умножение
			if p.peekTok.Type == lexer.COLON {
				break
			}
			line := p.curTok.Line
			p.nextToken()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = &BinaryExpr{Line: line, Op: "times", Left: left, Right: right}
		} else if p.curTok.Type == lexer.DIVIDED {
			line := p.curTok.Line
			p.nextToken()
			if p.curTok.Type != lexer.BY {
				return nil, errors.NewParseError(line, "Expected 'by' after 'divided'")
			}
			p.nextToken()
			right, err := p.parseFactor()
			if err != nil {
				return nil, err
			}
			left = &BinaryExpr{Line: line, Op: "divided", Left: left, Right: right}
		} else {
			break
		}
	}
	return left, nil
}

// parseFactor — литералы, идентификаторы, вызовы функций.
func (p *Parser) parseFactor() (Node, error) {
	line := p.curTok.Line

	switch p.curTok.Type {
	case lexer.NUMBER:
		val, err := strconv.ParseFloat(p.curTok.Literal, 64)
		if err != nil {
			return nil, errors.NewParseError(line, "Invalid number")
		}
		p.nextToken()
		return &NumberLiteral{Line: line, Value: val}, nil

	case lexer.STRING:
		val := p.curTok.Literal
		p.nextToken()
		return &StringLiteral{Line: line, Value: val}, nil

	case lexer.TRUE:
		p.nextToken()
		return &BoolLiteral{Line: line, Value: true}, nil

	case lexer.FALSE:
		p.nextToken()
		return &BoolLiteral{Line: line, Value: false}, nil

	case lexer.IDENT:
		name := p.curTok.Literal
		p.nextToken()
		return &Identifier{Line: line, Name: name}, nil

	case lexer.CALL:
		return p.parseCallExpression()

	default:
		return nil, nil
	}
}

// parseCallExpression: call name with arg
func (p *Parser) parseCallExpression() (*CallExpression, error) {
	line := p.curTok.Line
	if p.curTok.Type != lexer.CALL {
		return nil, errors.NewParseError(line, "Expected 'call'")
	}
	p.nextToken()

	if p.curTok.Type != lexer.IDENT {
		return nil, errors.NewParseError(line, "Expected function name after 'call'")
	}
	name := p.curTok.Literal
	p.nextToken()

	if p.curTok.Type != lexer.WITH {
		return nil, errors.NewParseError(line, "Expected 'with' after function name")
	}
	p.nextToken()

	arg, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if arg == nil {
		return nil, errors.NewParseError(line, "Expected argument after 'with'")
	}

	return &CallExpression{Line: line, Name: name, Arg: arg}, nil
}

// parseCondition — условие для if/while.
func (p *Parser) parseCondition() (Node, error) {
	left, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, errors.NewParseError(p.curTok.Line, "Expected condition expression")
	}

	if p.curTok.Type != lexer.IS {
		return nil, errors.NewParseError(p.curTok.Line, "Expected 'is' in condition")
	}
	line := p.curTok.Line
	p.nextToken()

	switch p.curTok.Type {
	case lexer.GREATER:
		p.nextToken()
		if p.curTok.Type != lexer.THAN {
			return nil, errors.NewParseError(line, "Expected 'than' after 'greater'")
		}
		p.nextToken()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ComparisonExpr{Line: line, Op: "greater", Left: left, Right: right}, nil

	case lexer.LESS:
		p.nextToken()
		if p.curTok.Type != lexer.THAN {
			return nil, errors.NewParseError(line, "Expected 'than' after 'less'")
		}
		p.nextToken()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ComparisonExpr{Line: line, Op: "less", Left: left, Right: right}, nil

	case lexer.EQUAL:
		p.nextToken()
		if p.curTok.Type != lexer.TO {
			return nil, errors.NewParseError(line, "Expected 'to' after 'equal'")
		}
		p.nextToken()
		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ComparisonExpr{Line: line, Op: "equal", Left: left, Right: right}, nil

	case lexer.TRUE:
		p.nextToken()
		return &ComparisonExpr{Line: line, Op: "istrue", Left: left, Right: nil}, nil

	case lexer.FALSE:
		p.nextToken()
		return &ComparisonExpr{Line: line, Op: "isfalse", Left: left, Right: nil}, nil

	default:
		return nil, errors.NewParseError(line, "Expected comparison after 'is'")
	}
}
