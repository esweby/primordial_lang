package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]Object),
		outer: nil,
	}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}

	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// Assign updates an existing binding in the nearest enclosing environment.
func (e *Environment) Assign(name string, val Object) (Object, bool) {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val, true
	}

	if e.outer != nil {
		return e.outer.Assign(name, val)
	}

	return nil, false
}
