# τ

Tau is an open source programming language designed to be minimal, fast and efficient.

## Syntax

### Hello World
We all start from here...
```
print("Hello World")
```

### Examples of syntax

#### if-else blocks

```
if 1 > 0 {
	print("yes")
} else {
	print("no")
}
```

```
myVar = 10

if myVar > 10 {
	print("more than 10")
} else if myVar == 10 {
	print("it's exactly 10")
} else {
	print(myVar)
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

loop(5, fn() { print("Hello World") })
```

#### Noteworthy features
The return value can be implicit:
```
add = fn(x, y) { x + y }
sum = add(9, 1)
print(sum)
10
```

Also you can inline the if expressions:
```
a = 0
b = 1

minimum = if a < b { a } else { b }
```
