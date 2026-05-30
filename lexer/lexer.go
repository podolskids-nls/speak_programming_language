// Пакет lexer — посимвольное чтение исходника и выдача потока токенов.
package lexer

import (
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Lexer разбирает строку исходного кода на токены.
type Lexer struct {
	input        string // весь исходный текст
	position     int    // текущая позиция (байт)
	readPosition int    // следующая позиция для чтения
	ch           byte   // текущий символ
	line         int    // текущая строка (с 1)
	column       int    // текущий столбец

	tokens []Token // накопленные токены
}

// New создаёт лексер для заданного исходного кода.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

// readChar продвигает указатель на один символ вперёд.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peek возвращает следующий символ, не двигая указатель.
func (l *Lexer) peek() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Tokenize разбирает весь вход и возвращает слайс токенов.
func (l *Lexer) Tokenize() []Token {
	l.processLines()
	// В конце потока — EOF
	l.tokens = append(l.tokens, Token{Type: EOF, Literal: "", Line: l.line})
	return l.tokens
}

// processLines обрабатывает файл построчно с учётом отступов (INDENT/DEDENT).
func (l *Lexer) processLines() {
	indentStack := []int{0} // стек уровней отступа; 0 — корневой уровень

	for l.ch != 0 {
		// Пропускаем пустые строки
		if l.ch == '\n' {
			l.readChar()
			continue
		}

		// Комментарий до конца строки
		if l.ch == '-' && l.peek() == '-' {
			l.skipComment()
			continue
		}

		// Начало строки с кодом — считаем отступ
		indent := l.measureIndent()
		currentIndent := indentStack[len(indentStack)-1]

		if indent > currentIndent {
			indentStack = append(indentStack, indent)
			l.tokens = append(l.tokens, Token{Type: INDENT, Literal: "", Line: l.line})
		} else if indent < currentIndent {
			for len(indentStack) > 1 && indent < indentStack[len(indentStack)-1] {
				indentStack = indentStack[:len(indentStack)-1]
				l.tokens = append(l.tokens, Token{Type: DEDENT, Literal: "", Line: l.line})
			}
		}

		// Токены текущей строки
		lineNum := l.line
		l.tokenizeLine()

		// Конец строки — NEWLINE (если не EOF)
		if l.ch == '\n' {
			l.tokens = append(l.tokens, Token{Type: NEWLINE, Literal: "\\n", Line: lineNum})
			l.readChar()
		}
	}

	// Закрываем все открытые блоки отступов
	for len(indentStack) > 1 {
		indentStack = indentStack[:len(indentStack)-1]
		l.tokens = append(l.tokens, Token{Type: DEDENT, Literal: "", Line: l.line})
	}
}

// measureIndent считает пробелы/табы в начале строки и продвигает указатель за ними.
func (l *Lexer) measureIndent() int {
	indent := 0
	for l.ch == ' ' || l.ch == '\t' {
		if l.ch == '\t' {
			indent += 4 // один таб = 4 пробела
		} else {
			indent++
		}
		l.readChar()
	}
	return indent
}

// skipComment пропускает строку комментария (-- ... до конца строки).
func (l *Lexer) skipComment() {
	for l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
}

// tokenizeLine разбирает одну логическую строку (без перевода строки в конце).
func (l *Lexer) tokenizeLine() {
	lineStart := l.line

	for l.ch != 0 && l.ch != '\n' {
		switch l.ch {
		case ' ', '\t':
			l.readChar()
		case ':':
			l.tokens = append(l.tokens, Token{Type: COLON, Literal: ":", Line: lineStart})
			l.readChar()
		case '"':
			l.readString()
		default:
			if isLetter(l.ch) {
				l.readIdentifier()
			} else if isDigit(l.ch) {
				l.readNumber()
			} else {
				l.readChar() // неизвестный символ — пропускаем
			}
		}
	}
}

// readIdentifier читает слово (идентификатор или ключевое слово).
func (l *Lexer) readIdentifier() {
	start := l.position
	line := l.line
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	word := l.input[start:l.position]
	tokType := LookupIdent(word)
	l.tokens = append(l.tokens, Token{Type: tokType, Literal: word, Line: line})
}

// readNumber читает целое или дробное число.
func (l *Lexer) readNumber() {
	start := l.position
	line := l.line
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peek()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	numStr := l.input[start:l.position]
	l.tokens = append(l.tokens, Token{Type: NUMBER, Literal: numStr, Line: line})
}

// readString читает строковый литерал в двойных кавычках.
func (l *Lexer) readString() {
	line := l.line
	l.readChar() // пропускаем открывающую "
	var result []byte
	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\\' && l.peek() == '"' {
			l.readChar()
			result = append(result, '"')
			l.readChar()
		} else if l.ch == '\n' {
			break // незакрытая строка
		} else {
			result = append(result, l.ch)
			l.readChar()
		}
	}
	if l.ch == '"' {
		l.readChar()
	}
	l.tokens = append(l.tokens, Token{Type: STRING, Literal: string(result), Line: line})
}

func isLetter(ch byte) bool {
	if ch >= utf8.RuneSelf {
		return false
	}
	r := rune(ch)
	return unicode.IsLetter(r) || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// ParseNumber преобразует текст числа в float64.
func ParseNumber(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
