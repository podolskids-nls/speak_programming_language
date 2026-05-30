# Спецификация языка Speak

## 1. Обзор

Speak — динамически типизированный интерпретируемый язык. Программа — последовательность операторов, блоки задаются отступами (4 пробела или 1 таб). Комментарии начинаются с `--`.

## 2. Токены

| Токен | Описание |
|-------|----------|
| `SET` | ключевое слово `set` |
| `TO` | ключевое слово `to` |
| `PRINT` | `print` |
| `IF` | `if` |
| `ELSE` | `else` |
| `REPEAT` | `repeat` |
| `TIMES` | `times` (в `repeat N times:` или умножение) |
| `WHILE` | `while` |
| `DEFINE` | `define` |
| `WITH` | `with` |
| `CALL` | `call` |
| `RETURN` | `return` |
| `IS` | `is` (в условиях) |
| `GREATER` | `greater` |
| `LESS` | `less` |
| `EQUAL` | `equal` |
| `THAN` | `than` |
| `TRUE` / `FALSE` | `true` / `false` |
| `PLUS` / `MINUS` | `plus` / `minus` |
| `DIVIDED` / `BY` | `divided` / `by` |
| `AND` / `OR` / `NOT` | зарезервированы |
| `IDENT` | идентификатор `[a-zA-Z_][a-zA-Z0-9_]*` |
| `NUMBER` | целое или дробное число |
| `STRING` | строка в `"..."` |
| `COLON` | `:` |
| `NEWLINE` | конец строки |
| `INDENT` / `DEDENT` | изменение уровня отступа |
| `EOF` | конец файла |

## 3. Грамматика (BNF)

```
program     ::= statement*

statement   ::= set_stmt
              | print_stmt
              | if_stmt
              | while_stmt
              | repeat_stmt
              | define_stmt
              | call_stmt
              | return_stmt

set_stmt    ::= "set" IDENT "to" expression

print_stmt  ::= "print" expression

if_stmt     ::= "if" condition ":" block else_part?
else_part   ::= "else" ":" block

while_stmt  ::= "while" condition ":" block

repeat_stmt ::= "repeat" expression "times" ":" block

define_stmt ::= "define" IDENT "with" IDENT ":" block

call_stmt   ::= "call" IDENT "with" expression

return_stmt ::= "return" expression

block       ::= INDENT statement* DEDENT

condition   ::= expression "is" "greater" "than" expression
              | expression "is" "less" "than" expression
              | expression "is" "equal" "to" expression
              | expression "is" "true"
              | expression "is" "false"

expression  ::= comparison ( ("plus" | "minus") comparison )*

comparison  ::= term ( "is" ... )?   /* только в condition, см. выше */

term        ::= factor ( "times" factor | "divided" "by" factor )*

factor      ::= NUMBER
              | STRING
              | "true" | "false"
              | IDENT
              | call_expr

call_expr   ::= "call" IDENT "with" expression
```

Примечание: в реализации `condition` разбирается отдельно; `expression` для присваиваний и `print` не содержит `is ...`.

## 4. Система типов

| Тип | Внутреннее представление | Литералы |
|-----|--------------------------|----------|
| number | `float64` | `10`, `3.14` |
| string | `string` | `"hello"` |
| boolean | `bool` | `true`, `false` |
| null | отсутствие значения | неявно при `return` без значения |

### Операции

| Операция | Правило |
|----------|---------|
| number `plus` number | number |
| number `minus` number | number |
| number `times` number | number |
| number `divided by` number | number (ошибка при делении на 0) |
| string `plus` string | string (конкатенация) |
| string `plus` number | string (число преобразуется в текст) |
| number `is greater than` number | boolean |
| number `is less than` number | boolean |
| number `is equal to` number | boolean |
| string `is equal to` string | boolean |
| expr `is true` / `is false` | boolean (expr должно быть bool) |

Несовместимые типы → `TypeError at line N: ...`

## 5. Области видимости (scoping)

- **Глобальная область** — переменные верхнего уровня.
- **Функция** (`define`) — параметр и локальные переменные в теле функции; замыкание захватывает среду на момент объявления.
- **if / while / repeat** — **не** создают новую область; переменные внутри блока видны снаружи.

## 6. Обработка ошибок

Форматы:

```
Error at line 3: Unknown variable 'total'
Error at line 7: Cannot divide by zero
Error at line 2: Expected expression after 'to'
TypeError at line 5: Cannot use 'times' with string and number
```

- Все ошибки содержат номер строки.
- Panic в Go перехватывается и преобразуется в сообщение для пользователя.
- При ошибке программа завершается с ненулевым кодом (`exit status 1`).

### Примеры для ручной проверки

| Файл | Код | Ожидаемое сообщение |
|------|-----|---------------------|
| `examples/error_test_1.speak` | `print total` | `Error at line 1: Unknown variable 'total'` |
| `examples/error_test_2.speak` | `set x to 10 divided by 0` | `Error at line 1: Cannot divide by zero` |

Запуск: `go run main.go examples/error_test_1.speak`

## 7. Приоритет операторов

От **высшего** к **низшему**:

| Приоритет | Операторы |
|-----------|-----------|
| 1 (выше) | `times`, `divided by` |
| 2 | `plus`, `minus` |
| 3 (ниже) | Сравнения (`is greater than`, …) — только в условиях |

Пример: `2 plus 3 times 4` = `2 + (3 * 4)` = **14**.

## 8. Функции

```speak
define square with n:
    return n times n

set result to call square with 7
```

- Одна функция — один параметр (`with name`).
- `return` завершает выполнение функции и возвращает значение.
- Рекурсия поддерживается через замыкание.

## 9. Примеры файлов

| Файл | Назначение |
|------|------------|
| `examples/hello.speak` | минимальный вывод |
| `examples/variables.speak` | переменные, строки |
| `examples/conditions.speak` | if / else |
| `examples/loops.speak` | repeat, while |
| `examples/functions.speak` | define, call, return |
| `examples/bonus.speak` | условие и арифметика |
| `examples/letter.speak` | строки и функции без математики |
| `examples/factorial.speak` | рекурсия, факториал |
| `examples/error_test_1.speak` | демо: неизвестная переменная |
| `examples/error_test_2.speak` | демо: деление на ноль |
