package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
	BuiltinScope
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	outer   *SymbolTable
	store   map[string]Symbol
	NumDefs int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{store: make(map[string]Symbol)}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer: outer,
		store: make(map[string]Symbol),
	}
}

func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{
		Name:  name,
		Index: s.NumDefs,
		Scope: GlobalScope,
	}

	if s.outer != nil {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.NumDefs++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]

	if !ok && s.outer != nil {
		return s.outer.Resolve(name)
	}

	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}
