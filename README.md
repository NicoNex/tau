# τ

Tau is a dinamically-typed open source programming language designed to be minimal, fast and efficient.

## Installation
In order to install Tau, you'll need [Go](https://golang.org/).

Once done, running the following command will successfully install the tau interpreter:
```bash
go install github.com/NicoNex/tau@latest
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
As every interpreter Tau supports files either by passing the path to the interpreter or by using the shabang.

```python
# helloworld.tau

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

##### REPL
Tau also supports REPL:
```
>>> add = fn(a, b) { a + b }
>>> string(add)
fn(a, b) { (a + b) }
>>> string(21)
21
>>> recursive_loop = fn(n, func) { if n != 0 { func(n); recursive_loop(n-1, func) } }
>>> recursive_loop(10, fn(n) { println("hello", n) })
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
my_var = 10
```

#### Float
```python
my_var = 2.5
```

#### String
```python
str = "My string here"
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

#### List
```js
empty = []
stuff = ["Hello World", 1, 2, 3, true]
```

#### Map
```js
empty = {}
stuff = {"Hello": "World", 123: true}
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

obj.sum_values = fn() {
	obj.value1 + obj.value2
}

obj.child = new()
obj.child.value = obj.sum_values()
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
