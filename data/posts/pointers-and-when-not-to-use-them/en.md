---
title: "Pointers in Go, and when they are not needed"
description: "Notes on chapter 6 of Learning Go: pointer syntax, pointers as mutability signals, stack and heap, escape analysis, and mechanical sympathy."
tags: [go, learning, pointers, memory]
date: 2026-03-30
---

## Chapter 6

Topics: pointer syntax, pointer conventions, pointers and performance, stack vs heap.

## Syntax

`&` is the address-of operator. It returns the memory address of a value.

```go
x := 10
p := &x
fmt.Println(p)  // address
fmt.Println(*p) // 10
```

`*` in front of a pointer dereferences it. The zero value of a pointer is `nil`. Dereferencing `nil` is a panic.

```go
var x *int
fmt.Println(x == nil) // true
fmt.Println(*x)       // panic
```

Go has no pointer arithmetic. The `unsafe` package exists for low-level operations but is not used in regular code.

## Other languages hide pointers

In many languages a class instance is internally a pointer. When an instance is passed to a method and a field is changed, the change is visible in the original. Reassigning the parameter itself has no effect. That is exactly pointer semantics.

Go makes the choice explicit: value type or pointer. Default is value type.

## Pointers signal mutability

Go has no immutable keyword. `const` only works for compile-time values. Instead there is a convention: when a function takes a pointer, mutation is allowed. Value types are treated as unchanged.

```go
func update(x *Thing) {
    x.field = "new"
}
```

## Pointers as a last resort

Idiomatic Go prefers value types.

```go
// not this
func MakeFoo(f *Foo) error {
    f.Field1 = "val"
    f.Field2 = 20
    return nil
}

// this
func MakeFoo() (Foo, error) {
    return Foo{Field1: "val", Field2: 20}, nil
}
```

Exception: interfaces like `json.Unmarshal` take a pointer so the function can write into caller-allocated memory.

## nil as "no value"

Pointers can encode the difference between zero value and "not set". For JSON with nullable fields this is idiomatic. Elsewhere a value type plus a `bool` is usually clearer.

## Maps and slices behave pointer-ish

A map is internally a pointer to a struct. A slice is a struct with length, capacity, and a pointer to the backing array.

When passed to a function, the pointer or header is copied. Mutations to contents show up in the original. Reassigning inside the function does not.

```go
func modSlice(s []int) {
    s[0] = 100         // visible in original
    s = append(s, 999) // NOT visible in original if a new backing array is allocated
}
```

Maps are not a good fit for public API parameters because nothing documents which keys are expected. A struct is clearer.

## Stack, heap, escape analysis

The compiler decides via escape analysis whether a value lives on the stack or heap. If the compiler cannot guarantee that a pointer does not outlive the current stack frame, the value goes on the heap.

Stack allocation is cheap. Heap allocation costs GC work.

Consequence: fewer pointers in code reduces GC pressure.

## Mechanical sympathy

Term borrowed from racing. Code that gets along with the hardware runs faster. RAM is "random access", but sequential reads are still significantly faster.

A `[]Foo` (slice of structs) sits sequentially in memory. A `[]*Foo` (slice of pointers) scatters data. The book references tests showing roughly a 100x difference for scattered pointer access.

Idiomatic Go naturally leads to layouts with less indirect memory access.

## Summary

- `&` for address, `*` to dereference
- Pointer zero value is `nil`, dereferencing that panics
- Pointers signal mutability to the reader
- Value types are the default, pointers the exception
- Maps and slices carry internal pointers, hence the different behavior
- Stack allocation is cheap, heap allocation costs GC
- Sequential memory is significantly faster than scattered
