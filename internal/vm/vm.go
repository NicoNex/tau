package vm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
)

type State struct {
	Consts  []obj.Object
	Globals []obj.Object
	Symbols *compiler.SymbolTable
}

func NewState() *State {
	st := compiler.NewSymbolTable()
	for i, builtin := range obj.Builtins {
		st.DefineBuiltin(i, builtin.Name)
	}

	return &State{
		Consts:  []obj.Object{},
		Globals: make([]obj.Object, GlobalSize),
		Symbols: st,
	}
}

type VM struct {
	stack      []obj.Object
	sp         int
	frames     []*Frame
	frameIndex int
	dir        string
	file string
	// Keeps track of the locally defined globals.
	localTable []bool
	*State
}

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

var (
	True  = obj.True
	False = obj.False
	Null  = obj.NullObj
)

func assertTypes(o obj.Object, types ...obj.Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func toFloat(l, r obj.Object) (obj.Object, obj.Object) {
	if i, ok := l.(obj.Integer); ok {
		l = obj.NewFloat(float64(i))
	}
	if i, ok := r.(obj.Integer); ok {
		r = obj.NewFloat(float64(i))
	}
	return l, r
}

func isTruthy(o obj.Object) bool {
	switch val := o.(type) {
	case *obj.Boolean:
		return o == obj.True
	case obj.Integer:
		return val.Val() != 0
	case obj.Float:
		return val.Val() != 0
	case *obj.Null:
		return false
	default:
		return true
	}
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}

func parserError(prefix string, errs []string) error {
	var buf strings.Builder

	buf.WriteString(prefix)
	buf.WriteRune('\n')
	for _, e := range errs {
		buf.WriteRune('\t')
		buf.WriteString(e)
		buf.WriteRune('\n')
	}

	return errors.New(buf.String())
}

func wait(fns ...func()) {
	var wg sync.WaitGroup

	wg.Add(len(fns))
	for _, fn := range fns {
		go func(fn func()) { fn(); wg.Done() }(fn)
	}
	wg.Wait()
}

func New(file string, bytecode *compiler.Bytecode) *VM {
	vm := &VM{
		stack:      make([]obj.Object, StackSize),
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
		localTable: make([]bool, GlobalSize),
		State:      NewState(),
	}

	vm.dir, vm.file = filepath.Split(file)
	vm.Consts = bytecode.Constants
	fn := &obj.CompiledFunction{Instructions: bytecode.Instructions}
	vm.frames[0] = NewFrame(&obj.Closure{Fn: fn}, 0)
	return vm
}

func NewWithState(file string, bytecode *compiler.Bytecode, state *State) *VM {
	vm := &VM{
		stack:      make([]obj.Object, StackSize),
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
		localTable: make([]bool, GlobalSize),
		State:      state,
	}

	vm.dir, vm.file = filepath.Split(file)
	fn := &obj.CompiledFunction{Instructions: bytecode.Instructions}
	vm.frames[0] = NewFrame(&obj.Closure{Fn: fn}, 0)
	return vm
}

func (vm *VM) clone() *VM {
	var tvm = &VM{
		stack:      make([]obj.Object, StackSize),
		sp:         vm.sp,
		dir:        vm.dir,
		frames:     make([]*Frame, MaxFrames),
		frameIndex: vm.frameIndex,
		localTable: make([]bool, GlobalSize),
		State: &State{
			Consts:  make([]obj.Object, len(vm.Consts)),
			Globals: make([]obj.Object, GlobalSize),
			Symbols: vm.Symbols,
		},
	}

	wait(
		func() {
			for i := range tvm.stack {
				tvm.stack[i] = vm.stack[i]
			}
		},
		func() {
			for i := range tvm.frames {
				tvm.frames[i] = vm.frames[i]
			}
		},
		func() {
			for i := range tvm.localTable {
				tvm.localTable[i] = vm.localTable[i]
			}
		},
		func() {
			for i := range tvm.Consts {
				tvm.Consts[i] = vm.Consts[i]
			}
		},
		func() {
			for i := range tvm.Globals {
				tvm.Globals[i] = vm.Globals[i]
			}
		},
	)

	return tvm
}

func (vm *VM) SetDir(dir string) {
	vm.dir = dir
}

func (vm *VM) isLocal(i int) bool {
	return vm.localTable[i]
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) LastPoppedStackElem() obj.Object {
	return vm.stack[vm.sp]
}

func (vm VM) execLoadModule() error {
	var taupath = vm.pop()

	pathObj, ok := taupath.(obj.String)
	if !ok {
		return fmt.Errorf("import: expected string, got %v", taupath.Type())
	}

	path, err := obj.ImportLookup(filepath.Join(vm.dir, string(pathObj)))
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tree, errs := parser.Parse(string(b))
	if len(errs) > 0 {
		p := fmt.Sprintf("import: multiple errors in module %s:", path)
		return parserError(p, errs)
	}

	c := compiler.NewWithState(vm.Symbols, &vm.Consts)
	if err := c.Compile(tree); err != nil {
		return err
	}

	tvm := NewWithState(path, c.Bytecode(), vm.State)
	tvm.dir, _ = filepath.Split(path)
	if err := tvm.Run(); err != nil {
		return err
	}

	mod := obj.NewModule()
	for name, symbol := range vm.Symbols.Store {
		if symbol.Scope == compiler.GlobalScope && tvm.isLocal(symbol.Index) {
			o := vm.Globals[symbol.Index]
			if m, ok := o.(obj.Moduler); ok {
				o = m.Module()
			}

			if isExported(name) {
				mod.Exported[name] = o
			} else {
				mod.Unexported[name] = o
			}
		}
	}

	return vm.push(mod)
}

func (vm *VM) pushInterpolated(strIdx, numSub int) error {
	var (
		str    = vm.Consts[strIdx].String()
		substr = make([]any, numSub)
	)

	for i := numSub - 1; i >= 0; i-- {
		substr[i] = vm.pop()
	}

	str = fmt.Sprintf(str, substr...)
	return vm.push(obj.NewString(str))
}

func (vm *VM) execDot() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch l := left.(type) {
	case obj.MapGetSetter:
		return vm.push(&obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				return l.Get(right.String())
			},
			SetFunc: func(o obj.Object) obj.Object {
				return l.Set(right.String(), o)
			},
		})

	case obj.GetSetter:
		m := l.Object().(obj.MapGetSetter)
		return vm.push(&obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				return m.Get(right.String())
			},

			SetFunc: func(o obj.Object) obj.Object {
				return m.Set(right.String(), o)
			},
		})

	default:
		return fmt.Errorf("%v object has no attribute %s", left.Type(), right)
	}
}

func (vm *VM) execDefine() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = vm.pop()
	)

	l, ok := left.(obj.Setter)
	if !ok {
		return fmt.Errorf("cannot assign to type %v", left.Type())
	}
	return vm.push(l.Set(right))
}

func (vm *VM) execAdd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.NewInteger(l + r))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(obj.String).Val()
		r := right.(obj.String).Val()
		return vm.push(obj.NewString(l + r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.NewFloat(l + r))

	default:
		return fmt.Errorf("unsupported operator '+' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execSub() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.NewInteger(l - r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.NewFloat(l - r))

	default:
		return fmt.Errorf("unsupported operator '-' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execMul() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.NewInteger(l * r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.NewFloat(l * r))

	default:
		return fmt.Errorf("unsupported operator '*' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execDiv() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType, obj.FloatType) || !assertTypes(right, obj.IntType, obj.FloatType) {
		return fmt.Errorf("unsupported operator '/' for types %v and %v", left.Type(), right.Type())
	}

	left, right = toFloat(left, right)
	l := left.(obj.Float).Val()
	r := right.(obj.Float).Val()
	return vm.push(obj.NewFloat(l / r))
}

func (vm *VM) execMod() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '%%' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()

	if r == 0 {
		return fmt.Errorf("can't divide by 0")
	}
	return vm.push(obj.NewInteger(l % r))
}

func (vm *VM) execBwAnd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '&' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return vm.push(obj.NewInteger(l & r))
}

func (vm *VM) execBwOr() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '|' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return vm.push(obj.NewInteger(l | r))
}

func (vm *VM) execBwXor() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '^' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return vm.push(obj.NewInteger(l ^ r))
}

func (vm *VM) execBwNot() error {
	var left = obj.Unwrap(vm.pop())

	if !assertTypes(left, obj.IntType) {
		return fmt.Errorf("unsupported operator '~' for type %v", left.Type())
	}

	l := left.(obj.Integer).Val()
	return vm.push(obj.NewInteger(^l))
}

func (vm *VM) execBwLShift() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '<<' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return vm.push(obj.NewInteger(l << r))
}

func (vm *VM) execBwRShift() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType) || !assertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '>>' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return vm.push(obj.NewInteger(l >> r))
}

func (vm *VM) execEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left == right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(obj.String).Val()
		r := right.(obj.String).Val()
		return vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.ParseBool(l == r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.ParseBool(l == r))

	default:
		return vm.push(False)
	}
}

func (vm *VM) execNotEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left != right))

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(obj.String).Val()
		r := right.(obj.String).Val()
		return vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.ParseBool(l != r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.ParseBool(l != r))

	default:
		return vm.push(True)
	}
}

func (vm *VM) execAnd() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	return vm.push(obj.ParseBool(isTruthy(left) && isTruthy(right)))
}

func (vm *VM) execIn() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return fmt.Errorf("unsupported operator 'in' for type %v", left.Type())
	}
	if !assertTypes(right, obj.ListType, obj.StringType) {
		return fmt.Errorf("unsupported operator 'in' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(obj.String).Val()
		r := right.(obj.String).Val()
		return vm.push(obj.ParseBool(strings.Contains(r, l)))

	case assertTypes(right, obj.ListType):
		for _, o := range right.(obj.List).Val() {
			if !assertTypes(left, o.Type()) {
				continue
			}
			if assertTypes(left, obj.BoolType, obj.NullType) && left == o {
				return vm.push(obj.True)
			}

			switch l := left.(type) {
			case obj.String:
				r := o.(obj.String)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}

			case obj.Integer:
				r := o.(obj.Integer)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}

			case obj.Float:
				r := o.(obj.Float)
				if l.Val() == r.Val() {
					return vm.push(obj.True)
				}
			}
		}
		return vm.push(obj.False)

	default:
		return fmt.Errorf(
			"invalid operation %v in %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (vm *VM) execOr() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	return vm.push(obj.ParseBool(isTruthy(left) || isTruthy(right)))
}

func (vm *VM) execGreaterThan() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.ParseBool(l > r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.ParseBool(l > r))

	default:
		return fmt.Errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execGreaterThanEqual() error {
	var (
		right = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return vm.push(obj.ParseBool(l >= r))

	case assertTypes(left, obj.IntType, obj.FloatType) && assertTypes(right, obj.IntType, obj.FloatType):
		left, right = toFloat(left, right)
		l := left.(obj.Float).Val()
		r := right.(obj.Float).Val()
		return vm.push(obj.ParseBool(l >= r))

	default:
		return fmt.Errorf("unsupported operator '>=' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execIndex() error {
	var (
		index = obj.Unwrap(vm.pop())
		left  = obj.Unwrap(vm.pop())
	)

	switch {
	case assertTypes(left, obj.ListType) && assertTypes(index, obj.IntType):
		l := left.(obj.List)
		i := int(index.(obj.Integer).Val())

		return vm.push(&obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				if i < 0 || i >= len(l) {
					return obj.NewError("intex out of range"), false
				}
				return l[i], true
			},

			SetFunc: func(o obj.Object) obj.Object {
				if i < 0 || int(i) >= len(l) {
					return obj.NewError("intex out of range")
				}
				l[i] = o
				return o
			},
		})

	case assertTypes(left, obj.BytesType) && assertTypes(index, obj.IntType):
		b := left.(obj.Bytes)
		i := int(index.(obj.Integer))

		if i < 0 || i >= len(b) {
			return fmt.Errorf("index out of range")
		}
		return vm.push(obj.NewInteger(int64(b[i])))

	case assertTypes(left, obj.StringType) && assertTypes(index, obj.IntType):
		s := left.(obj.String)
		i := int(index.(obj.Integer))

		if i < 0 || i >= len(s) {
			return fmt.Errorf("index out of range")
		}
		return vm.push(obj.NewString(string(s[i])))

	case assertTypes(left, obj.MapType) && assertTypes(index, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
		m := left.(obj.Map)
		k := index.(obj.Hashable)

		return vm.push(&obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				v := m.Get(k.KeyHash()).Value
				return v, v != obj.NullObj
			},

			SetFunc: func(o obj.Object) obj.Object {
				m.Set(k.KeyHash(), obj.MapPair{Key: index, Value: o})
				return o
			},
		})

	default:
		return fmt.Errorf("invalid index operator for types %v and %v", left.Type(), index.Type())
	}
}

func (vm *VM) execBang() error {
	var right = obj.Unwrap(vm.pop())

	switch b := right.(type) {
	case *obj.Boolean:
		return vm.push(obj.ParseBool(!b.Val()))

	case *obj.Null:
		return vm.push(True)

	default:
		return vm.push(False)
	}
}

func (vm *VM) execMinus() error {
	var right = obj.Unwrap(vm.pop())

	switch r := right.(type) {
	case obj.Integer:
		return vm.push(obj.NewInteger(-r.Val()))

	case obj.Float:
		return vm.push(obj.NewFloat(-r.Val()))

	default:
		return fmt.Errorf("unsupported prefix operator '-' for type %v", r.Type())
	}
}

func (vm *VM) execReturnValue() error {
	retVal := obj.Unwrap(vm.pop())
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1

	return vm.push(retVal)
}

func (vm *VM) execReturn() error {
	frame := vm.popFrame()
	vm.sp = frame.basePointer - 1

	return vm.push(Null)
}

func (vm *VM) execCurrentClosure() error {
	return vm.push(vm.currentFrame().cl)
}

func (vm *VM) call(o obj.Object, numArgs int) error {
	switch fn := o.(type) {
	case *obj.Closure:
		return vm.callClosure(fn, numArgs)
	case obj.Builtin:
		return vm.callBuiltin(fn, numArgs)
	case obj.Getter:
		return vm.call(fn.Object(), numArgs)
	default:
		return fmt.Errorf("calling non-function")
	}
}

func (vm *VM) execCall(numArgs int) error {
	return vm.call(vm.stack[vm.sp-1-numArgs], numArgs)
}

func (vm *VM) execConcurrentCall(numArgs int) error {
	tvm := vm.clone()
	go tvm.call(tvm.stack[tvm.sp-1-numArgs], numArgs)
	return nil
}

func (vm *VM) buildList(start, end int) obj.Object {
	var elements = make([]obj.Object, end-start)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	return obj.NewList(elements...)
}

func (vm *VM) buildMap(start, end int) (obj.Object, error) {
	var m = obj.NewMap()

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := obj.MapPair{Key: key, Value: val}
		mapKey, ok := key.(obj.Hashable)
		if !ok {
			return nil, fmt.Errorf("invalid map key type %v", key.Type())
		}
		m.Set(mapKey.KeyHash(), pair)
	}

	return m, nil
}

func (vm *VM) callClosure(cl *obj.Closure, nargs int) error {
	if nargs != cl.Fn.NumParams {
		return fmt.Errorf("wrong number of arguments: expected %d, got %d", cl.Fn.NumParams, nargs)
	}

	frame := NewFrame(cl, vm.sp-nargs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals
	return nil
}

func (vm *VM) callBuiltin(fn obj.Builtin, nargs int) error {
	args := vm.stack[vm.sp-nargs : vm.sp]
	res := fn(args...)
	vm.sp = vm.sp - nargs - 1

	if res == nil {
		return vm.push(Null)
	}
	return vm.push(res)
}

func (vm *VM) pushClosure(constIdx, numFree int) error {
	constant := vm.Consts[constIdx]
	fn, ok := constant.(*obj.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]obj.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree
	return vm.push(&obj.Closure{Fn: fn, Free: free})
}

func (vm *VM) Run() (err error) {
	var (
		ip  int
		ins code.Instructions
		op  code.Opcode
	)

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 && err == nil {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.Consts[constIndex])

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			if cond := obj.Unwrap(vm.pop()); !isTruthy(cond) {
				vm.currentFrame().ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.localTable[globalIndex] = true
			vm.Globals[globalIndex] = vm.peek()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err = vm.push(vm.Globals[globalIndex])

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.peek()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err = vm.push(vm.stack[frame.basePointer+int(localIndex)])

		case code.OpList:
			nElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			list := vm.buildList(vm.sp-nElements, vm.sp)
			vm.sp = vm.sp - nElements
			err = vm.push(list)

		case code.OpMap:
			nElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			mapObj, e := vm.buildMap(vm.sp-nElements, vm.sp)
			if e != nil {
				return e
			}
			vm.sp = vm.sp - nElements
			err = vm.push(mapObj)

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			err = vm.execCall(int(numArgs))

		case code.OpConcurrentCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			err = vm.execConcurrentCall(int(numArgs))

		case code.OpGetBuiltin:
			idx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			def := obj.Builtins[idx]
			err = vm.push(def.Builtin)

		case code.OpClosure:
			constIdx := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			err = vm.pushClosure(int(constIdx), int(numFree))

		case code.OpGetFree:
			freeIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			closure := vm.currentFrame().cl
			err = vm.push(closure.Free[freeIdx])

		case code.OpLoadModule:
			err = vm.execLoadModule()

		case code.OpInterpolate:
			strIdx := code.ReadUint16(ins[ip+1:])
			numArgs := code.ReadUint16(ins[ip+3:])
			vm.currentFrame().ip += 4
			err = vm.pushInterpolated(int(strIdx), int(numArgs))

		case code.OpDot:
			err = vm.execDot()

		case code.OpDefine:
			err = vm.execDefine()

		case code.OpCurrentClosure:
			err = vm.execCurrentClosure()

		case code.OpReturn:
			err = vm.execReturn()

		case code.OpReturnValue:
			err = vm.execReturnValue()

		case code.OpNull:
			err = vm.push(Null)

		case code.OpIndex:
			err = vm.execIndex()

		case code.OpTrue:
			err = vm.push(True)

		case code.OpFalse:
			err = vm.push(False)

		case code.OpAdd:
			err = vm.execAdd()

		case code.OpSub:
			err = vm.execSub()

		case code.OpMul:
			err = vm.execMul()

		case code.OpDiv:
			err = vm.execDiv()

		case code.OpMod:
			err = vm.execMod()

		case code.OpBwAnd:
			err = vm.execBwAnd()

		case code.OpBwOr:
			err = vm.execBwOr()

		case code.OpBwXor:
			err = vm.execBwXor()

		case code.OpBwLShift:
			err = vm.execBwLShift()

		case code.OpBwRShift:
			err = vm.execBwRShift()

		case code.OpEqual:
			err = vm.execEqual()

		case code.OpNotEqual:
			err = vm.execNotEqual()

		case code.OpGreaterThan:
			err = vm.execGreaterThan()

		case code.OpGreaterThanEqual:
			err = vm.execGreaterThanEqual()

		case code.OpAnd:
			err = vm.execAnd()

		case code.OpIn:
			err = vm.execIn()

		case code.OpOr:
			err = vm.execOr()

		case code.OpBang:
			err = vm.execBang()

		case code.OpMinus:
			err = vm.execMinus()

		case code.OpPop:
			vm.pop()
		}

	}
	return
}

func (vm *VM) push(o obj.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() obj.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) peek() obj.Object {
	return vm.stack[vm.sp-1]
}
