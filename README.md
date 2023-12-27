# τ

Tau is a dynamically-typed open source programming language designed to be minimal, fast and efficient.

## Installation
In order to install Tau, you'll need [Go](https://golang.org/).

Once done, running the following command will successfully install the tau interpreter:
```bash
go install github.com/NicoNex/tau/cmd/tau@latest
```

You can try it out in the terminal by simply running `$ tau`, alternatively to take advantage of the builtin virtual machine and gain a lot of performance run it with `$ tau -vm`.
The flag `-vm` works when executing files too.
For additional info run `$ tau --help`.

## Syntax

### Hello World
_We all start from here..._
```python
println("Hello World")
```

### Examples

#### File
As every interpreter Tau supports files either by passing the path to the interpreter or by using the shebang.

```python
#!/path/to/tau

println("hello world")
```

```
$ tau helloworld.tau
hello world
```

#### if-else blocks

```python
if 1 > 0 {
	println("yes")
} else {
	println("no")
}
```

```python
myVar = 10

if myVar > 10 {
	println("more than 10")
} else if myVar == 10 {
	println("it's exactly 10")
} else {
	println(myVar)
}
```

#### Declaring a function
```python
loop = fn(times, function) {
	if times > 0 {
		function()
		loop(times-1, function)
	}
}

loop(5, fn() { println("Hello World") })
```

#### Noteworthy features
The return value can be implicit:
```python
add = fn(x, y) { x + y }
sum = add(9, 1)
println(sum)
```
```
>>> 10
```

Also you can inline the if expressions:
```rust
a = 0
b = 1

minimum = if a < b { a } else { b }
```

The semicolon character `;` is implicit on a newline but can be used to separate multiple expressions on a single line.
```python
printData = fn(a, b, c) { println(a); println(b); println(c) }
```

Functions are first-class and treated as any other data type.
```python
min = fn(a, b) {
	if a < b {
		return a
	}
	b
}

var1 = 1
var2 = 2

m = min(var1, var2)
println(m)
```
```
>>> 1
```

##### Error handling
```python
# errtest.tau

div = fn(n, d) {
	if d == 0 {
		return error("zero division error")
	}
	n / d
}

if failed(result1 = div(16, 2)) {
	exit(result1)
}
println("the result of 16 / 2 is", result1)

if failed(result2 = div(32, 0)) {
	exit(result2)
}
println("the result of 32 / 0 is", result2)
```
```
$ tau errtest.tau
the result of 16 / 2 is 8
error: zero division error
$
```

##### Beautiful error messages
```python
# errtest.tau

increment = fn(n) {
	return n + 1
}

increment("this will raise a runtime error")
```
```
error in file errtest.tau at line 4:
    return n + 1
             ^
unsupported operator '+' for types string and int
```

#### Concurrency
Tau supports go-style concurrency. 
This is obtained by the use of four builtins `pipe`, `send`, `recv` `close`. 
- `pipe` creates a new FIFO pipe and optionally you can pass an integer to it to create a buffered pipe.
- `send` is used to send values to the pipe.
- `recv` is used to receive values from the pipe.
- `close` closes the pipe.

Pipes can be buffered or unbuffered. Buffered pipes make the tau-routine sleep once `send` is called until at least one value is read from the pipe.
Once `recv` is called on an empty pipe it will cause the tau-routine to sleep until a new value is sent to the pipe.
`send` is used to send values to the pipe.
`close` closes the pipe thus allowing it to be garbage collected. 
Calling `recv` on a closed pipe will return `null`.

```python
# concurrency_example.tau

listen = fn(p) {
	for val = recv(p) {
		println(val)
	}
	println("bye bye...")
}

p = pipe()
tau listen(p)

send(p, "hello")
send(p, "world")
send(p, 123)
send(p, "this is a test")
close(p)
```

##### REPL
Tau also supports REPL:
```
>>> add = fn(a, b) { a + b }
>>> string(add)
fn(a, b) { (a + b) }
>>> string(21)
21
>>> recursiveLoop = fn(n, func) { if n != 0 { func(n); recursiveLoop(n-1, func) } }
>>> recursiveLoop(10, fn(n) { println("hello", n) })
hello 10
hello 9
hello 8
hello 7
hello 6
hello 5
hello 4
hello 3
hello 2
hello 1
```

### Data types
Tau is a dynamically-typed programming language and it supports the following primitive types:

#### Integer
```python
myVar = 10
```

#### Float
```python
myVar = 2.5
```

#### String
```python
myString = "My string here"
```
Tau also supports strings interpolation.
```python
temp = 25
myString = "The temperature is { if temp > 20 { \"hot\" } else { \"cold\" } }"
println(myString)
```
```
>>> The temperature is hot
```
For raw strings use the backtick instead of double quotes.
```python
s = `this is a raw string\n {}`
println(s)
```
```
>>> this is a raw string\n {}
```

#### Boolean
```js
t = true
f = false
```

#### Function
```python
pow = fn(base, exponent) {
	if exponent > 0 {
		return base * pow(base, exponent-1)
	}
	1 # You could optionally write 'return 1', but in this case the return is implicit.
}
```

#### Builtin Functions

Tau has an assortment of useful builtin functions that operate on many data types:

- `len(x)` -- Returns the length of the given object `x` which could be a String, List, Map or Bytes.
- `println(s)` -- Prints the String `s` to the terminal (standard out) along with a new-line.
- `print(s)` -- Same as `print()` without a new-line.
- `input(prompt)` -- Asks for input from the user by reading from the terminal (standard in) with an optional prompt.
- `string(x)` -- Converts the object `x` to a String.
- `error(s)` -- Constructs a new error with the contents of the String `s`.
- `type(x)` -- Returns the type of the object `x`.
- `int(x)` -- Converts the object `x` to an Integer.
- `float(x)` -- Converts the object `x` to a Float.
- `exit([code | message, code]) -- Terminates the program with the optional exit code and/or message.
- `append(xs, x)` -- Appends the object `x` to the List `xs` and returns the new List.
- `new` -- Constructs a new empty object.
- `failed(f)` -- Calls the Function `f` and returns true if an error occurred.
- `plugin(path)` -- Loads the Plugin at the given path.
- `pipe` -- Creates a new pipe for sending/receiving messages to/from coroutines.
- `send(p, x)` -- Sends the object `x` to the pipe `p`.
- `recv(p)` -- Reads from the pipe `p` and returns the next object sent to it.
- `close(p)` -- Closes the pipe `p`.
- `hex(x)` -- Returns a hexadecimal representation of `x`.
- `oct(x)` -- Returns an octal representation of `x`.
- `bin(x)` -- Returns a binary representation of `x`.
- `slice(x, start, end)` -- Returns a slice of `x` from `start` to `end` which could be a String, List or Bytes.
- `keys(x)` -- Returns a List of keys of the Map `x`.
- `delete(xs, x)` -- Deletes the key `x` or the `x`th item from the Map `xs` or List `xs`.
- `bytes(x)` -- Converts the String `x` to Bytes.

#### List
```js
empty = []
stuff = ["Hello World", 1, 2, 3, true]
```

You can append to a list with the `append()` builtin:

```js
xs =[]
xs = append(xs, 1)
```

Lists can be indexed using the indexing operator `[n]`:

```js
xs = [1, 2, 3]
xs[1]
```

#### Map
```js
empty = {}
stuff = {"Hello": "World", 123: true}
```

Keys can be added using the set operator `[key] = value`:

```js
kv = {}
k["foo"] = "bar"
```

Keys can be accessed using the get operator `[key]`:

```js
kv = ["foo": "bar"}
kv["foo"]
```

#### Loop
```python
for i = 0; i < 10; ++i {
	println("hello world", i)
}

lst = [0, 1, 2, 3, 4]

println(lst)
for len(lst) > 0 {
	println(lst = tail(lst))
}
```

#### Objects
```python
obj = new()
obj.value1 = 123
obj.value2 = 456

obj.sumValues = fn() {
	obj.value1 + obj.value2
}

obj.child = new()
obj.child.value = obj.sumValues()
```

##### Recommended usage
```python
Dog = fn(name, age) {
	dog = new()

	dog.name = name
	dog.age = age

	dog.humanage = fn() {
		dog.age * 7
	}

	return dog
}

snuffles = Dog("Snuffles", 8)
println(snuffles.humanage())
```
```
>>> 56
```

#### Files
It's possible to open files with the `open` builtin. 
The aforementioned builtin supports the following file modes:
- `r` opens a file read-only and it's the default mode when no mode is specified.
- `w` opens a file write-only truncating it to zero length. If the file doesn't exist it creates it.
- `a` opens a file in append mode for reading and writing and it creates it if doesn't exist.
- `x` opens a file in exclusive mode for reading and writing, if the file doesn't exist it creates it and fails otherwise.
- `rw` opens a file for reading and writing truncating it to zero length first.

```python
# file_example.tau

f = open("myfile.txt")
content = f.Read()
f.Close()
```

```python
# file_example.tau

f = open("myfile.txt", "a")
content = f.Read()
f.Write("Hello World")
f.Close()

println("previous content: {content}")
```

#### Modules
##### Import
When importing a module only the fields whose name start with an upper-case character will be exported.
Same thing applies for exported objects, in the example `Snuffles` is exported but the field `id` won't be visible ouside the module.
```python
# import_test.tau

data = 123

printData = fn() {
	println(data)
}

printText = fn() {
	println("example text")
}

TestPrint = fn() {
	printData()
	printText()
}

dog = fn(name, age) {
	d = new()
	d.Name = name
	d.Age = age
	d.id = 123

	d.ID = fn() {
		d.id
	}

	return d
}

Snuffles = dog("Mr Snuffles", 5)

```

```python
it = import("import_test")

it.TestPrint()

println(it.Snuffles.Name)
println(it.Snuffles.Age)
println(it.Snuffles.ID())
```

```
>>> 123
>>> example text
>>> Mr Snuffles
>>> 5
>>> 456
```

##### Plugin
Tau plugin system makes it possible to import and use Go plugins in Tau seamlessly.
To run your Go code in Tau just compile it with:
```bash
go build -buildmode=plugin -o myplugin.so
```
then you can import it in Tau with the `plugin` builtin function.
```python
myplugin = plugin("path/to/myplugin.so")
```
###### Example
Go code:
```golang
package main

import "fmt"

func Hello() {
	fmt.Println("Hello World")
}

func Sum(a, b int) int {
	return a + b
}
```

Tau code:
```python
myplugin = plugin("myplugin.so")

myplugin.Hello()
println("The sum is", myplugin.Sum(3, 2))
```
Output:
```
>>> Hello World
>>> The sum is 5
```
