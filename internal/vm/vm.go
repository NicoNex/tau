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
	"github.com/NicoNex/tau/internal/tauerr"
)

type State struct {
	Symbols *compiler.SymbolTable
	Consts  []obj.Object
	Globals []obj.Object
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
	*State
	dir        string
	file       string
	stack      []obj.Object
	frames     []*Frame
	localTable []bool
	sp         int
	frameIndex int
}

type getter interface {
	Get(string) (obj.Object, bool)
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

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}

func parserError(prefix string, errs []error) error {
	var buf strings.Builder

	buf.WriteString(prefix)
	buf.WriteByte('\n')
	for _, e := range errs {
		buf.WriteString(e.Error())
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
	fn := &obj.CompiledFunction{
		Instructions: bytecode.Instructions,
		Bookmarks:    bytecode.Bookmarks,
	}
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
	fn := &obj.CompiledFunction{
		Instructions: bytecode.Instructions,
		Bookmarks:    bytecode.Bookmarks,
	}
	vm.frames[0] = NewFrame(&obj.Closure{Fn: fn}, 0)
	return vm
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

// Returns the bookmark corresponding to the current position in the bytecode.
func (vm *VM) bookmark() tauerr.Bookmark {
	var (
		frame     = vm.currentFrame()
		offset    = frame.ip
		bookmarks = frame.cl.Fn.Bookmarks
	)

	if len(bookmarks) == 0 {
		return tauerr.Bookmark{}
	}

	prev := bookmarks[0]
	for _, cur := range bookmarks[1:] {
		if offset < prev.Offset {
			return prev
		} else if offset > prev.Offset && offset <= cur.Offset {
			return cur
		}
		prev = cur
	}
	return prev
}

func (vm *VM) errorf(s string, a ...any) error {
	return tauerr.NewFromBookmark(
		filepath.Join(vm.dir, vm.file),
		vm.bookmark(),
		s,
		a...,
	)
}

func (vm *VM) execLoadModule() error {
	var taupath = vm.pop()

	pathObj, ok := taupath.(obj.String)
	if !ok {
		return vm.errorf("import: expected string, got %v", taupath.Type())
	}

	path, err := obj.ImportLookup(filepath.Join(vm.dir, string(pathObj)))
	if err != nil {
		return vm.errorf("import: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tree, errs := parser.Parse(path, string(b))
	if len(errs) > 0 {
		p := fmt.Sprintf("import: multiple errors in module %s:", path)
		return parserError(p, errs)
	}

	c := compiler.NewWithState(vm.Symbols, &vm.Consts)
	c.SetFileInfo(path, string(b))
	if err := c.Compile(tree); err != nil {
		return err
	}

	tvm := NewWithState(path, c.Bytecode(), vm.State)
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
		buf    strings.Builder
		str    = vm.Consts[strIdx].String()
		substr = make([]string, numSub)
		count  = 0
	)

	for i := numSub - 1; i >= 0; i-- {
		substr[i] = vm.pop().String()
	}

	for _, b := range []byte(str) {
		if b == 0xff {
			buf.WriteString(substr[count])
			count++
			continue
		}
		buf.WriteByte(b)
	}
	return vm.push(obj.String(buf.String()))
}

func (vm *VM) execDot() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.ObjectType) {
		return vm.errorf("%v object has no attribute %s", left.Type(), right)
	}

	switch l := left.(type) {
	case getter:
		o, _ := l.Get(right.String())
		return vm.push(o)

	default:
		return vm.errorf("%v object has no attribute %s", left.Type(), right)
	}
}

func (vm *VM) execDefine() error {
	var (
		val    = vm.pop()
		field  = vm.pop()
		target = vm.pop()
	)

	switch t := target.(type) {
	case obj.TauObject:
		t.Set(string(field.(obj.String)), val)
		vm.push(val)

	case obj.List:
		idx := int(field.(obj.Integer))

		if idx < 0 || idx >= len(t) {
			return vm.errorf("index out of range")
		}
		t[idx] = val
		vm.push(val)

	case obj.Map:
		h, ok := field.(obj.Hashable)
		if !ok {
			return vm.errorf("%v object is not hashable", field.Type())
		}
		t.Set(h.KeyHash(), obj.MapPair{Key: field, Value: val})
		vm.push(val)

	default:
		return vm.errorf("cannot assign to type %q", target.Type())
	}
	return nil
}

func (vm *VM) execAdd() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.Integer(l + r))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return vm.push(obj.String(l + r))

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.Float(l + r))

	default:
		return vm.errorf("unsupported operator '+' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execSub() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.Integer(l - r))

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.Float(l - r))

	default:
		return vm.errorf("unsupported operator '-' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execMul() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.Integer(l * r))

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.Float(l * r))

	default:
		return vm.errorf("unsupported operator '*' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execDiv() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) || !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return fmt.Errorf("unsupported operator '/' for types %v and %v", left.Type(), right.Type())
	}

	left, right = obj.ToFloat(left, right)
	l := left.(obj.Float)
	r := right.(obj.Float)
	return vm.push(obj.Float(l / r))
}

func (vm *VM) execMod() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '%%' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)

	if r == 0 {
		return vm.errorf("can't divide by 0")
	}
	return vm.push(obj.Integer(l % r))
}

func (vm *VM) execBwAnd() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '&' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return vm.push(obj.Integer(l & r))
}

func (vm *VM) execBwOr() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '|' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return vm.push(obj.Integer(l | r))
}

func (vm *VM) execBwXor() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '^' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return vm.push(obj.Integer(l ^ r))
}

func (vm *VM) execBwNot() error {
	var left = vm.pop()

	if !obj.AssertTypes(left, obj.IntType) {
		return fmt.Errorf("unsupported operator '~' for type %v", left.Type())
	}

	l := left.(obj.Integer)
	return vm.push(obj.Integer(^l))
}

func (vm *VM) execBwLShift() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '<<' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return vm.push(obj.Integer(l << r))
}

func (vm *VM) execBwRShift() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	if !obj.AssertTypes(left, obj.IntType) || !obj.AssertTypes(right, obj.IntType) {
		return fmt.Errorf("unsupported operator '>>' for types %v and %v", left.Type(), right.Type())
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return vm.push(obj.Integer(l >> r))
}

func (vm *VM) execEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.BoolType, obj.NullType) || obj.AssertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left == right))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return vm.push(obj.ParseBool(l == r))

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.ParseBool(l == r))

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.ParseBool(l == r))

	default:
		return vm.push(False)
	}
}

func (vm *VM) execNotEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.BoolType, obj.NullType) || obj.AssertTypes(right, obj.BoolType, obj.NullType):
		return vm.push(obj.ParseBool(left != right))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return vm.push(obj.ParseBool(l != r))

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.ParseBool(l != r))

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.ParseBool(l != r))

	default:
		return vm.push(True)
	}
}

func (vm *VM) execAnd() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	return vm.push(obj.ParseBool(obj.IsTruthy(left) && obj.IsTruthy(right)))
}

func (vm *VM) execOr() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	return vm.push(obj.ParseBool(obj.IsTruthy(left) || obj.IsTruthy(right)))
}

func (vm *VM) execGreaterThan() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.ParseBool(l > r))

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.ParseBool(l > r))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return vm.push(obj.ParseBool(l > r))

	default:
		return vm.errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execGreaterThanEqual() error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return vm.push(obj.ParseBool(l >= r))

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return vm.push(obj.ParseBool(l >= r))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return vm.push(obj.ParseBool(l >= r))

	default:
		return vm.errorf("unsupported operator '>=' for types %v and %v", left.Type(), right.Type())
	}
}

func (vm *VM) execIndex() error {
	var (
		index = vm.pop()
		left  = vm.pop()
	)

	switch {
	case obj.AssertTypes(left, obj.ListType) && obj.AssertTypes(index, obj.IntType):
		l := left.(obj.List)
		i := int(index.(obj.Integer))

		if i < 0 || i >= len(l) {
			return vm.errorf("index out of range")
		}
		return vm.push(l[i])

	case obj.AssertTypes(left, obj.BytesType) && obj.AssertTypes(index, obj.IntType):
		b := left.(obj.Bytes)
		i := int(index.(obj.Integer))

		if i < 0 || i >= len(b) {
			return fmt.Errorf("index out of range")
		}
		return vm.push(obj.NewInteger(int64(b[i])))

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(index, obj.IntType):
		s := left.(obj.String)
		i := int(index.(obj.Integer))

		if i < 0 || i >= len(s) {
			return vm.errorf("index out of range")
		}
		return vm.push(obj.NewString(string(s[i])))

	case obj.AssertTypes(left, obj.MapType) && obj.AssertTypes(index, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
		m := left.(obj.Map)
		k := index.(obj.Hashable)
		return vm.push(m.Get(k.KeyHash()).Value)

	default:
		return vm.errorf("invalid index operator for types %v and %v", left.Type(), index.Type())
	}
}

func (vm *VM) execBang() error {
	var right = vm.pop()

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
	var right = vm.pop()

	switch r := right.(type) {
	case obj.Integer:
		return vm.push(obj.Integer(-r))

	case obj.Float:
		return vm.push(obj.Float(-r))

	default:
		return vm.errorf("unsupported prefix operator '-' for type %v", r.Type())
	}
}

func (vm *VM) execReturnValue() error {
	retVal := vm.pop()
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

func (vm *VM) execCall(numArgs int) error {
	switch fn := vm.stack[vm.sp-1-numArgs].(type) {
	case *obj.Closure:
		return vm.callClosure(fn, numArgs)
	case obj.Builtin:
		return vm.callBuiltin(fn, numArgs)
	default:
		return vm.errorf("calling non-function")
	}
}

func (vm *VM) execConcurrentCall(numArgs int) error {
	tvm := &VM{
		stack:      make([]obj.Object, StackSize),
		frames:     make([]*Frame, MaxFrames),
		frameIndex: 1,
		localTable: make([]bool, GlobalSize),
		dir:        vm.dir,
		file:       vm.file,
		sp:         vm.sp,
		State: &State{
			Consts:  vm.Consts,
			Globals: make([]obj.Object, GlobalSize),
			Symbols: vm.Symbols,
		},
	}

	wait(
		func() { copy(tvm.stack, vm.stack) },
		func() { copy(tvm.localTable, vm.localTable) },
		func() { copy(tvm.Globals, vm.Globals) },
	)

	if err := tvm.execCall(numArgs); err != nil {
		return err
	}
	go tvm.Run()
	return vm.push(Null)
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
			return nil, vm.errorf("invalid map key type %v", key.Type())
		}
		m.Set(mapKey.KeyHash(), pair)
	}

	return m, nil
}

func (vm *VM) callClosure(cl *obj.Closure, nargs int) error {
	if nargs != cl.Fn.NumParams {
		return vm.errorf("wrong number of arguments: expected %d, got %d", cl.Fn.NumParams, nargs)
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
		return vm.errorf("not a function: %+v", constant)
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

	for err == nil {
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

			if cond := vm.pop(); !obj.IsTruthy(cond) {
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

		case code.OpOr:
			err = vm.execOr()

		case code.OpBang:
			err = vm.execBang()

		case code.OpMinus:
			err = vm.execMinus()

		case code.OpPop:
			vm.pop()

		case code.OpHalt:
			return
		}

	}
	return
}

func (vm *VM) push(o obj.Object) error {
	if vm.sp >= StackSize {
		return vm.errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() obj.Object {
	vm.sp--
	o := vm.stack[vm.sp]
	return o
}

func (vm *VM) peek() obj.Object {
	return vm.stack[vm.sp-1]
}
