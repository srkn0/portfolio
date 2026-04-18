---
title: "Slices, maps, structs and shared memory"
description: "Notes on chapter 3 of Learning Go: arrays, slices, maps, and structs. Focus on shared backing storage in slices and the comma-ok idiom in maps."
tags: [go, learning, data-structures]
date: 2026-03-16
---

## Chapter 3

Topics: arrays, slices, maps, structs.

## Arrays

In Go the size is part of the type. `[3]int` and `[4]int` are different types. A function that accepts `[3]int` does not accept `[4]int`.

```go
var x [3]int
var y [4]int
x = y // compile error
```

Arrays are rarely used directly in practice. They act as backing storage for slices.

## Slices

Slices are declared without a size in the type.

```go
x := []int{10, 20, 30}
```

Three internal fields: length, capacity, and a pointer to the backing array.

### `append` returns a new slice

```go
var x []int
x = append(x, 10)
```

The result must be assigned back. Go is call-by-value: `append` receives a copy of the slice header, appends, and returns the new version.

### Length vs capacity

Length is the number of in-use elements. Capacity is the size of the reserved backing array.

```go
x := make([]int, 0, 10) // length 0, capacity 10
x = append(x, 1, 2, 3)  // length 3, capacity 10
```

When an `append` exceeds the capacity, Go allocates a new, larger backing array and copies. Pre-allocating with `make` reduces allocations and GC pressure.

### Slicing shares memory

```go
x := []string{"a", "b", "c", "d"}
y := x[:2] // y = ["a", "b"]
x[1] = "z"
fmt.Println(y) // ["a", "z"]
```

Slicing does not copy. Both variables point at the same backing array.

The three-index slice expression caps the capacity, so a later `append` allocates a new array.

```go
y := x[:2:2] // capacity capped at 2
```

## Maps

```go
teams := map[string]int{
    "Orcas": 1,
    "Lions": 2,
}
```

Important properties:

- Zero value is `nil`. Reading from a nil map returns the zero value of the value type. Writing to a nil map panics.
- Maps cannot be compared with `==`.
- Keys must be comparable. Slices and maps cannot be keys.

### Comma-ok idiom

A missing key returns the zero value without an error. To distinguish missing from zero, use the comma-ok form.

```go
v, ok := m["key"]
if ok {
    // key exists, v holds the value
}
```

### Map as a set

Go has no built-in set type. Common patterns are `map[T]bool` or `map[T]struct{}` for minimal memory use.

```go
intSet := map[int]bool{}
intSet[5] = true
if intSet[5] {
    fmt.Println("5 is in the set")
}
```

## Structs

Used to group related fields.

```go
type Person struct {
    Name string
    Age  int
    Pet  string
}
```

Properties:

- Struct literals can use field names (recommended) or positional values.
- Structs are comparable with `==` if all fields are comparable.
- Anonymous structs are possible, often used for one-off shapes in tests.

Go has no classes and no inheritance. Structs with methods replace those concepts.

## Summary

- Arrays exist, rarely used directly
- `append` returns a new slice, always reassign
- Length and capacity are two different values
- Slicing shares memory, three-index expression caps the capacity
- Comma-ok idiom distinguishes missing keys from zero values
- Structs group data, no inheritance
