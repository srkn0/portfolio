---
title: "Generics in Go"
description: "Notes on chapter 8 of Learning Go: type parameters, comparable, generic functions, type terms with union syntax, the tilde constraint, cmp.Ordered, and the deliberate omissions in Go's generics."
tags: [go, learning, generics]
date: 2026-04-13
---

## Chapter 8

Topics: type parameters, constraints, generic functions, type terms, what Go's generics omit.

## Before generics

Before Go 1.18 there were two ways to write a data structure for arbitrary types. One copy per type, or an `Orderable` interface using `any`. Both had drawbacks.

```go
type Orderable interface {
    Order(any) int
}
```

Using `any` lets mixed types through without a compile-time check.

```go
it.Insert(OrderableInt(5))
it.Insert(OrderableString("nope")) // compiles, panics at runtime
```

Generics close that gap.

## Syntax

```go
type Stack[T any] struct {
    vals []T
}

func (s *Stack[T]) Push(val T) {
    s.vals = append(s.vals, val)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.vals) == 0 {
        var zero T
        return zero, false
    }
    top := s.vals[len(s.vals)-1]
    s.vals = s.vals[:len(s.vals)-1]
    return top, true
}
```

`[T any]` after the type name declares the type parameter. After that, `T` is used like a normal type. On methods, `[T]` appears on the receiver: `func (s *Stack[T])`.

`var zero T` returns the zero value of the generic type. `nil` is not possible because `T` could also be `int`.

## Constraints are interfaces

`any` allows any type but excludes operators. `==` needs the built-in `comparable`.

```go
type Stack[T comparable] struct {
    vals []T
}

func (s Stack[T]) Contains(val T) bool {
    for _, v := range s.vals {
        if v == val {
            return true
        }
    }
    return false
}
```

`comparable` matches all types that work with `==` and `!=`. Slices and maps are excluded.

## Generic functions

```go
func Map[T1, T2 any](s []T1, f func(T1) T2) []T2 {
    r := make([]T2, len(s))
    for i, v := range s {
        r[i] = f(v)
    }
    return r
}

func Filter[T any](s []T, f func(T) bool) []T {
    var r []T
    for _, v := range s {
        if f(v) {
            r = append(r, v)
        }
    }
    return r
}
```

Type inference works in most cases.

```go
words := []string{"One", "Potato", "Two", "Potato"}
filtered := Filter(words, func(s string) bool {
    return s != "Potato"
})
```

## Type terms for operators

For arithmetic operators, `any` is not enough. An interface with type terms lists the allowed types.

```go
type Integer interface {
    int | int8 | int16 | int32 | int64 |
        uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

func divAndRemainder[T Integer](num, denom T) (T, T, error) {
    if denom == 0 {
        return 0, 0, errors.New("cannot divide by zero")
    }
    return num / denom, num % denom, nil
}
```

The `|` is a union. The allowed operators are those defined on every listed type term.

Type terms match exactly. For user-defined types whose underlying type is one of the listed types, the tilde is needed.

```go
type Integer interface {
    ~int | ~int8 | ...
}
```

`~int` matches any type whose underlying type is `int`.

## cmp.Ordered and cmp.Compare

Since Go 1.21 the standard library contains the `cmp` package with `cmp.Ordered` and `cmp.Compare`.

```go
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

For a generic tree, `cmp.Compare[int]` can be passed as the ordering function.

## What Go generics omit

Deliberate restrictions.

**No operator overloading.** Custom container types support neither `range` nor `[]`. Reasoning: many operators, no method overloading, and operator overloading hurts readability.

**No type parameters on methods.** Method chaining like `xs.Map(...).Reduce(...)` does not work for generic methods. Methods can only use the receiver's type parameters.

**No variadic type parameters.** A function cannot take an arbitrary number of differing type parameters.

**No specialization, currying, or metaprogramming.**

## Comparable and runtime

`comparable` excludes non-comparable types at compile time. An interface value whose concrete type is a slice still compiles under a `comparable` constraint and panics on `==` at runtime.

## Summary

- `[T any]` declares a type parameter
- `var zero T` for the zero value of a generic type
- `comparable` as a constraint for `==`-friendly types
- Type terms (`int | int8 | ...`) enable operator support
- `~int` matches user-defined types whose underlying type is `int`
- Type inference works in most cases
- `cmp.Ordered` and `cmp.Compare` since Go 1.21
- No operator overloading, no method type parameters, no fluent chaining
- `comparable` does not prevent runtime panics for interfaces holding non-comparable concrete types
