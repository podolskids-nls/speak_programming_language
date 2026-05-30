// Пакет errors — типы ошибок интерпретатора Speak и форматирование сообщений для пользователя.
package errors

import (
	"fmt"
)

// SpeakError — базовый интерфейс ошибки с номером строки исходного кода.
type SpeakError interface {
	error
	Line() int
}

// RuntimeError — ошибка выполнения (неизвестная переменная, деление на ноль и т.д.).
type RuntimeError struct {
	line    int
	message string
}

func (e *RuntimeError) Error() string {
	return fmt.Sprintf("Error at line %d: %s", e.line, e.message)
}

func (e *RuntimeError) Line() int { return e.line }

// NewRuntimeError создаёт ошибку выполнения с указанной строкой.
func NewRuntimeError(line int, message string) *RuntimeError {
	return &RuntimeError{line: line, message: message}
}

// TypeError — ошибка несовместимости типов (например, times со строкой).
type TypeError struct {
	line    int
	message string
}

func (e *TypeError) Error() string {
	return fmt.Sprintf("TypeError at line %d: %s", e.line, e.message)
}

func (e *TypeError) Line() int { return e.line }

// NewTypeError создаёт ошибку типов.
func NewTypeError(line int, message string) *TypeError {
	return &TypeError{line: line, message: message}
}

// ParseError — синтаксическая ошибка при разборе программы.
type ParseError struct {
	line    int
	message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Error at line %d: %s", e.line, e.message)
}

func (e *ParseError) Line() int { return e.line }

// NewParseError создаёт ошибку парсера.
func NewParseError(line int, message string) *ParseError {
	return &ParseError{line: line, message: message}
}

// Recover перехватывает panic и превращает его в понятное сообщение об ошибке.
// Используется в интерпретаторе, чтобы Go-паники не «протекали» пользователю.
func Recover() (err error) {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case SpeakError:
			err = v
		case string:
			err = NewRuntimeError(0, v)
		case error:
			err = NewRuntimeError(0, v.Error())
		default:
			err = NewRuntimeError(0, fmt.Sprintf("internal error: %v", v))
		}
	}
	return err
}
