package semantic

type Scope map[string]Symbol

type SymbolTable struct {
	scope Scope
	outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{scope: make(Scope)}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.outer = outer
	return st
}

func (st *SymbolTable) Get(name string) (Symbol, bool) {
	if sym, ok := st.scope[name]; ok {
		return sym, true
	}

	if st.outer != nil {
		return st.outer.Get(name)
	}

	return nil, false
}

func (st *SymbolTable) Set(name string, sym Symbol) {
	st.scope[name] = sym
}

func (st *SymbolTable) ExistsInCurrentScope(name string) bool {
	_, ok := st.scope[name]
	return ok
}
