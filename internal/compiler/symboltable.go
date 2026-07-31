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
	// Names used before their definition, with the position of their first
	// use for the error message. They are cleared as the definitions show up.
	pending map[string]int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{Store: make(map[string]Symbol), pending: make(map[string]int)}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer:   outer,
		Store:   make(map[string]Symbol),
		pending: make(map[string]int),
	}
}

// Names are the global names defined so far, which is what a prompt completes
// and what it lists when asked what is defined.
func (s *SymbolTable) Names() []string {
	g := s.global()

	out := make([]string, 0, len(g.Store))
	for name, sym := range g.Store {
		if sym.Scope == GlobalScope {
			out = append(out, name)
		}
	}
	return out
}

// global returns the outermost table, the one holding the global names.
func (s *SymbolTable) global() *SymbolTable {
	for s.outer != nil {
		s = s.outer
	}
	return s
}

// DefineForward reserves a global for a name used before its definition, so
// that functions can call each other whatever their order in the file.
func (s *SymbolTable) DefineForward(name string, pos int) Symbol {
	g := s.global()

	// Already known, either defined or reserved by an earlier use.
	if symbol, ok := g.Store[name]; ok {
		return symbol
	}

	symbol := g.Define(name)
	g.pending[name] = pos
	return symbol
}

// Pending returns a name still unresolved and the position of its first use.
func (s *SymbolTable) Pending() (string, int, bool) {
	for name, pos := range s.global().pending {
		return name, pos, true
	}
	return "", 0, false
}

func (s *SymbolTable) Define(name string) Symbol {
	// The name is defined now: whoever used it earlier was right to.
	delete(s.global().pending, name)

	// A builtin is only a default: a name given to something else takes it
	// over, the way a module called hex or a variable called len does.
	//
	// A captured name is only a default too. A closure reads what it takes
	// from the function around it but writing to that name makes a local of
	// its own, which shadows it from here on. The value it starts from is
	// whatever was read on the right of the assignment, which was compiled
	// while the name still meant the captured one.
	if symbol, ok := s.Store[name]; ok && symbol.Scope != BuiltinScope && symbol.Scope != FreeScope {
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
