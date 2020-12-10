# τ

Tau is an open source programming language designed to be minimal, fast and efficient.

## Syntax

### Hello World
We all start from here...
```
println("Hello World")
```

### Examples of syntax

#### File
![file](./images/taufile.png)

#### if-else blocks

```
if 1 > 0 {
	println("yes")
} else {
	println("no")
}
```

```
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
```
loop = fn(times, function) {
	if times != 0 {
		function()
		loop(times-1, function)
	}
}

loop(5, fn() { println("Hello World") })
```

#### Noteworthy features
The return value can be implicit:
```
add = fn(x, y) { x + y }
sum = add(9, 1)
println(sum)
10
```

Also you can inline the if expressions:
```
a = 0
b = 1

minimum = if a < b { a } else { b }
```

The semicolon character `;` is implicit on a newline but can be used to separate multiple expressions on a single line.
```
printData = fn(a, b, c) { println(a); println(b); println(c) }
```

##### REPL
Tau also supports a REPL:
![repl](./images/tauloop.png)
