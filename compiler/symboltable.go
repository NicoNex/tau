package compiler

type SymbolScope int

const (
	GlobalScope SymbolScope = iota
	LocalScope
	BuiltinScope
	FreeScope
	FunctionScope
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	outer       *SymbolTable
	store       map[string]Symbol
	NumDefs     int
	FreeSymbols []Symbol
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
		obj, ok := s.outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}

		return s.DefineFree(obj), true
	}

	return obj, ok
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{
		Name:  original.Name,
		Index: len(s.FreeSymbols) - 1,
		Scope: FreeScope,
	}
	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(n string) Symbol {
	symbol := Symbol{Name: n, Index: 0, Scope: FunctionScope}
	s.store[n] = symbol
	return symbol
}
