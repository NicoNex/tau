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
	Store       map[string]Symbol
	FreeSymbols []Symbol
	NumDefs     int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{Store: make(map[string]Symbol)}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer: outer,
		Store: make(map[string]Symbol),
	}
}

func (s *SymbolTable) Define(name string) Symbol {
	if symbol, ok := s.Store[name]; ok {
		return symbol
	}

	symbol := Symbol{
		Name:  name,
		Index: s.NumDefs,
		Scope: GlobalScope,
	}

	if s.outer != nil {
		symbol.Scope = LocalScope
	}

	s.Store[name] = symbol
	s.NumDefs++
	return symbol
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.Store[name]

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
	s.Store[name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{
		Name:  original.Name,
		Index: len(s.FreeSymbols) - 1,
		Scope: FreeScope,
	}
	s.Store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) DefineFunctionName(n string) Symbol {
	symbol := Symbol{Name: n, Index: 0, Scope: FunctionScope}
	s.Store[n] = symbol
	return symbol
}
