// Пакет lexer — лексический анализ исходного кода Speak (разбиение на токены).
package lexer

// TokenType — тип лексемы в языке Speak.
type TokenType string

// Все типы токенов согласно спецификации языка.
const (
	// Ключевые слова — команды и управляющие конструкции
	SET     TokenType = "SET"
	TO      TokenType = "TO"
	PRINT   TokenType = "PRINT"
	IF      TokenType = "IF"
	ELSE    TokenType = "ELSE"
	REPEAT  TokenType = "REPEAT"
	TIMES   TokenType = "TIMES"
	WHILE   TokenType = "WHILE"
	DEFINE  TokenType = "DEFINE"
	WITH    TokenType = "WITH"
	CALL    TokenType = "CALL"
	RETURN  TokenType = "RETURN"
	IS      TokenType = "IS"
	GREATER TokenType = "GREATER"
	LESS    TokenType = "LESS"
	EQUAL   TokenType = "EQUAL"
	THAN    TokenType = "THAN"
	TRUE    TokenType = "TRUE"
	FALSE   TokenType = "FALSE"
	PLUS    TokenType = "PLUS"
	MINUS   TokenType = "MINUS"
	DIVIDED TokenType = "DIVIDED"
	BY      TokenType = "BY"
	AND     TokenType = "AND"
	OR      TokenType = "OR"
	NOT     TokenType = "NOT"

	// Литералы и идентификаторы
	IDENT  TokenType = "IDENT"
	NUMBER TokenType = "NUMBER"
	STRING TokenType = "STRING"
	BOOL   TokenType = "BOOL"

	// Служебные токены: двоеточие, перевод строки, отступы
	COLON   TokenType = "COLON"
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"
	EOF     TokenType = "EOF"
)

// Token — одна лексема с типом, текстом и номером строки в исходнике.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

// keywords — словарь: слово → тип токена.
var keywords = map[string]TokenType{
	"set":     SET,
	"to":      TO,
	"print":   PRINT,
	"if":      IF,
	"else":    ELSE,
	"repeat":  REPEAT,
	"times":   TIMES,
	"while":   WHILE,
	"define":  DEFINE,
	"with":    WITH,
	"call":    CALL,
	"return":  RETURN,
	"is":      IS,
	"greater": GREATER,
	"less":    LESS,
	"equal":   EQUAL,
	"than":    THAN,
	"true":    TRUE,
	"false":   FALSE,
	"plus":    PLUS,
	"minus":   MINUS,
	"divided": DIVIDED,
	"by":      BY,
	"and":     AND,
	"or":      OR,
	"not":     NOT,
}

// LookupIdent возвращает тип токена для идентификатора или ключевого слова.
func LookupIdent(word string) TokenType {
	if tok, ok := keywords[word]; ok {
		return tok
	}
	return IDENT
}
