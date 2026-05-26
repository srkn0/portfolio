---
title: "Go: mental model and rules of thumb"
description: "Cross-cutting notes from Learning Go. Stack vs heap, call by value, zero values, pointer semantics in maps and slices, mechanical sympathy, GC, concurrency rules of thumb."
tags: [go, learning, mental-model, performance]
date: 2026-05-25
---

## What this is

Condensed rules and mechanics from the chapters on pointers, maps, slices, concurrency, and garbage collection. These statements showed up repeatedly throughout the book.

## Stack and heap

- Stack is free. Allocation is just bumping a pointer.
- Heap costs GC. Every heap allocation has to be collected later.
- The compiler decides via escape analysis where a value lives.
- A value escapes to the heap when the compiler cannot guarantee no pointer outlives the stack frame.
- Fewer pointers means fewer heap allocations means less GC pressure.

## Memory layout

- `[]Foo` sits sequentially in memory. Fast access via prefetching.
- `[]*Foo` scatters data across RAM. Pointer chasing is orders of magnitude slower.
- Mechanical sympathy: code that respects the hardware runs faster. Sequential beats random.

## Call by value

- Everything is call by value. Functions receive copies.
- For primitives and structs: the original stays unchanged.
- Maps and slices behave pointer-ish because they internally hold pointers.
- A map is internally a pointer to a struct.
- A slice is a header of length, capacity, and a pointer to the backing array.

## Slice behavior

- `append` may allocate a new backing array. Always reassign the result.
- Slicing shares memory. Changes show up in both variables.
- The three-index slice expression `x[:n:n]` caps the capacity.
- `make([]T, 0, n)` avoids repeated reallocations when the size is known.

## Map behavior

- Zero value is `nil`. Reading from a nil map returns the zero value. Writing to a nil map panics.
- Maps cannot be compared with `==`.
- Keys must be comparable. Slices and maps are excluded.
- `map[T]bool` or `map[T]struct{}` as a set substitute.
- Maps are not thread-safe. Concurrent use needs `sync.RWMutex` or `sync.Map`.

## Zero values

- Every variable has a zero value.
- `0`, `""`, `false`, `nil` depending on the type.
- Types are designed so that the zero value is useful.
- `sync.Mutex{}`, `bytes.Buffer{}`, `[]int(nil)` and `map[string]int(nil)` are usable in many operations directly.

## Pointers and nil

- The zero value of a pointer is `nil`. Dereferencing `nil` panics.
- A pointer receiver can accept a nil receiver if the method handles the case.
- An interface variable is only `nil` when both type and value are `nil`.
- Never return a nil concrete type wrapped in an interface.

## Method sets

- A pointer receiver belongs to the method set of the pointer type.
- A value receiver belongs to both method sets (value and pointer).
- Consequence: only `*T` satisfies an interface that requires a pointer-receiver method.

## Concurrency

- Concurrency is structure, not automatic performance.
- Goroutines are cheap, but each one must be cleanly terminated.
- Channels are synchronous by default. A send blocks until a receive happens.
- Buffered channels are used deliberately: known count or backpressure.
- The sender closes the channel.
- `select` picks randomly among ready cases. Prevents starvation.
- A `default` in a for-select is almost always wrong (busy loop).

## Context cancellation

- Context propagates cancellation and deadlines through the call chain.
- `defer cancel()` right after `context.WithCancel/Timeout/Deadline`.
- Calling `cancel()` more than once is idempotent.
- Child contexts are bounded by their parent.
- Long-running loops periodically check `context.Cause(ctx)`.

## Error handling

- Errors are values, not exceptions.
- Error is the last return value. `nil` means success.
- Wrap with `%w`. Unwrap with `errors.Is` and `errors.As`.
- `panic` only for unrecoverable states.
- `recover` only at the boundary of a library API.

## Tooling

- `go fmt` is mandatory, not a matter of taste.
- `go vet` catches common bugs (shadowing, wrong Printf formats).
- Unused imports and variables do not compile.
- `_test.go` files sit next to the code, in the same package.
- Reuse standard library interfaces (`io.Reader`, `io.Writer`).

## General stance

- Readability beats cleverness. Clear data flow matters more than short lines.
- Explicit beats implicit. No hidden conversions, no truthiness.
- Conventions instead of language features. `if err != nil` instead of exceptions, pointers as a mutability signal instead of an immutability keyword.
- Small, focused interfaces. Defined by the caller.
- Concurrency as a structuring tool, not a performance hammer.
