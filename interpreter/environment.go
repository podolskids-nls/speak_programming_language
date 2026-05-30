// Пакет interpreter — выполнение AST (tree-walking интерпретатор).
package interpreter

// Environment хранит переменные текущей области видимости и ссылку на родительскую.
type Environment struct {
	vars   map[string]interface{} // имя переменной → значение
	parent *Environment             // родительская область (nil для глобальной)
}

// NewEnvironment создаёт пустую глобальную область видимости.
func NewEnvironment() *Environment {
	return &Environment{vars: make(map[string]interface{})}
}

// NewEnclosed создаёт вложенную область (например, при вызове функции).
func NewEnclosed(parent *Environment) *Environment {
	return &Environment{vars: make(map[string]interface{}), parent: parent}
}

// Get ищет переменную в текущей области и выше по цепочке parent.
func (e *Environment) Get(name string) (interface{}, bool) {
	if val, ok := e.vars[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set записывает переменную в текущую область (не в parent).
func (e *Environment) Set(name string, val interface{}) {
	e.vars[name] = val
}

// Assign обновляет переменную: ищет существующую в цепочке или создаёт в текущей.
func (e *Environment) Assign(name string, val interface{}) {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return
	}
	if e.parent != nil {
		if _, ok := e.parent.Get(name); ok {
			e.parent.Assign(name, val)
			return
		}
	}
	e.vars[name] = val
}
