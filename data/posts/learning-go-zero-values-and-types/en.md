---
title: "Zero values, explicit types, and the compiler"
description: "Notes on chapter 2 of Learning Go: built-in types, variable declarations, zero values, explicit type conversion, truthiness, and constants."
tags: [go, learning, fundamentals]
date: 2026-03-09
---

## Chapter 2

Topics: built-in types, variable declarations, zero values, type conversion, constants, unused variables.

## Zero values

Every declared variable in Go is initialized with a zero value, even without an explicit assignment.

```go
var x int
var s string
var flag bool
fmt.Println(x)    // 0
fmt.Println(s)    // "" (empty string)
fmt.Println(flag) // false
```

There is no `undefined`, no `null`. The default value is defined per type.

## No automatic type conversion

Go does not convert between numeric types automatically. An `int` and a `float64` can only be added with an explicit conversion.

```go
var x int = 10
var y float64 = 30.2
var sum float64 = float64(x) + y
```

## No truthiness

Conditions must evaluate to a boolean. An empty string or `0` is not implicitly treated as `false`.

```go
if s == "" {
    // string is empty
}

if x == 0 {
    // number is zero
}
```

`if s` alone does not compile.

## `var` vs `:=`

Two forms for variable declaration.

```go
var x int = 10
```

and

```go
x := 10
```

`:=` infers the type. Inside functions this is the common form. `var` is useful when the zero value is used intentionally:

```go
var x int
```

or when a specific type matters:

```go
var x byte = 20
```

More readable than `x := byte(20)`.

## Constants

Constants are evaluated at compile time. No function calls, no runtime computation.

```go
const x = 10        // fine
const y = 20 * 10   // fine
const z = len("hi") // fine (built-in)

a := 5
const b = a + 1     // doesn't compile, a is a variable
```

There is no mechanism for making a runtime-computed value immutable.

## Unused variables

A local variable that is declared but never read is a compile error.

```go
func main() {
    x := 10  // compile error: x declared but not used
}
```

## Summary

- Every variable has a zero value, no undefined state
- Type conversion is always explicit
- Conditions require real booleans
- `:=` inside functions, `var` for zero values or explicit types
- Constants are compile-time only
- Unused local variables are a compile error
