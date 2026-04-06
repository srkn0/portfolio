---
title: "Types, methods, interfaces"
description: "Notes on chapter 7 of Learning Go: named types, methods with receivers, embedding instead of inheritance, implicit interfaces, accept interfaces return structs, the nil-interface trap, type assertions and switches."
tags: [go, learning, interfaces, types]
date: 2026-04-06
---

## Chapter 7

Topics: named types, methods, embedding, interfaces, type assertions.

## Named types

Any underlying type can be given a new name.

```go
type Score int
type Converter func(string) Score
type TeamScores map[string]Score
```

`Score` has `int` as the underlying type but is a separate type. Assigning an `int` to a `Score` variable requires an explicit conversion. The benefit: the type documents the meaning of the value.

## Methods

Methods are functions with a receiver between `func` and the name.

```go
type Person struct {
    FirstName string
    LastName  string
}

func (p Person) FullName() string {
    return p.FirstName + " " + p.LastName
}
```

Receiver name convention: short, usually the first letter of the type. `this` or `self` is not idiomatic.

Methods must be declared in the same package as the type. Types from external packages cannot be extended.

## Value receiver vs pointer receiver

```go
func (c Counter) String() string { ... }  // value receiver
func (c *Counter) Increment()    { ... }  // pointer receiver
```

Rules:

- If the method modifies the receiver, use a pointer receiver
- If the method should handle a nil receiver, use a pointer receiver
- Otherwise, a value receiver

When a type has at least one pointer-receiver method, all methods are usually written with pointer receivers for consistency.

Go automatically performs the conversion at the call site.

```go
var c Counter
c.Increment() // becomes (&c).Increment()
```

Method sets remain strict: a pointer-receiver method only belongs to the method set of the pointer type. This matters with interfaces.

## Methods for nil receivers

A method call on a nil pointer receiver can work if the method handles the case. A nil value receiver panics.

Example from the book: a binary tree that interprets `nil` as an empty tree.

```go
func (t *IntTree) Insert(val int) *IntTree {
    if t == nil {
        return &IntTree{val: val}
    }
    if val < t.val {
        t.left = t.left.Insert(val)
    } else if val > t.val {
        t.right = t.right.Insert(val)
    }
    return t
}
```

## No inheritance

Go has no `class` and no `extends`. `type HighScore Score` creates a new type with the same underlying type. Methods on `Score` do not belong to `HighScore`. Substitution is not possible.

## Composition via embedding

Instead of inheritance, embedding. A struct can contain other types as anonymous fields.

```go
type Employee struct {
    Name string
    ID   string
}

func (e Employee) Description() string {
    return e.Name + " (" + e.ID + ")"
}

type Manager struct {
    Employee
    Reports []Employee
}
```

Fields and methods of the embedded type are reachable directly on the outer type.

```go
m := Manager{
    Employee: Employee{Name: "Bob", ID: "12345"},
}
fmt.Println(m.Description()) // Bob (12345)
```

Embedding is not inheritance. A `Manager` cannot be assigned to an `Employee` variable. There is no dynamic dispatch for concrete types.

## Implicit interfaces

```go
type Stringer interface {
    String() string
}
```

Any type with a `String() string` method satisfies the interface. No explicit declaration, no `implements` keyword.

```go
type Counter struct{ n int }

func (c Counter) String() string {
    return fmt.Sprintf("%d", c.n)
}

// Counter satisfies fmt.Stringer.
```

The book frames this as a combination of two worlds. Explicit interfaces like Java's make requirements visible. Duck typing like Python's skips the paperwork. In Go, the caller declares the interface it needs. The implementation only has to match.

Interfaces in Go are kept small, often a single method. Defined at the call site where the interface is needed.

## Accept interfaces, return structs

Idiomatic pattern. Functions that do work take interfaces. Functions that produce values return concrete types.

```go
func Process(r io.Reader) error
func NewReader() *bufio.Reader
```

Benefit: callers can pass any implementation. Producer functions can be extended without breaking interfaces.

## The nil-interface trap

Interfaces are internally a pair of type pointer and value pointer. An interface is only `nil` when both are `nil`.

```go
var pc *Counter        // nil pointer
var i Incrementer      // nil interface
fmt.Println(pc == nil) // true
fmt.Println(i == nil)  // true

i = pc
fmt.Println(i == nil)  // false
```

Consequence: do not return a nil concrete type wrapped in an interface.

```go
func getCounter() Incrementer {
    var pc *Counter
    return pc // interface is not nil
}
```

## Type assertions and type switches

```go
var i any = 42
n, ok := i.(int)
if !ok {
    // wasn't an int
}
```

The comma-ok form makes type assertions safe. Without `ok`, a failed assertion panics.

For multiple possible types, a type switch.

```go
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("other type")
}
```

## any (the empty interface)

`any` is an alias for `interface{}`. A variable of type `any` can hold any value.

```go
data := map[string]any{}
json.Unmarshal(contents, &data)
```

Reasonable for JSON of unknown shape. Otherwise `any` undermines type safety and should be used sparingly.

## Summary

- Methods belong to a type via the receiver, no `this`
- Pointer receivers for mutation or nil-receiver support
- Embedding is composition with promotion, not inheritance
- Interfaces are implicit, the compiler checks methods
- Interfaces belong to the caller, kept small
- Accept interfaces, return structs
- A nil interface requires both type and value to be nil
- Type assertion with comma-ok, type switch for multiple options
- `any` is an escape hatch, not a default
