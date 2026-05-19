---
title: "Idiomatic Go in a nutshell"
description: "A collection of the main conventions from Learning Go. Naming, receivers, interfaces, error handling, context, defer, slices and maps, tests. Cross-reference post for all earlier chapters."
tags: [go, learning, conventions, idiomatic]
date: 2026-05-18
---

## What this is

A summary of the conventions covered in chapters 2 through 15 of Learning Go, all on one page. No new material, just a reference for later.

## Formatting and tooling

```bash
go fmt ./...
go vet ./...
go test ./...
```

Tabs for indentation, braces on the same line. `gofmt` is mandatory, not a matter of taste. The semicolon insertion rule enforces brace placement.

## Naming

- Packages: short, lowercase, no underscore, no camelCase
- Exported: first character uppercase
- Receivers: single letter, not `this` or `self`
- Interfaces: single-method interfaces often use the `er` suffix (`Reader`, `Writer`, `Closer`)
- Error variables: `Err` prefix, e.g. `ErrNotFound`
- Constants in `iota` blocks: typed underlying type

## Variable declaration

```go
x := 10           // typical form inside functions
var x int         // explicit zero value
var x byte = 20   // explicit type
const x = 10      // compile-time value
```

Unused local variables do not compile.

## Pointer vs value

| Situation | Recommendation |
|---|---|
| Method modifies the receiver | Pointer receiver |
| Method needs to handle `nil` | Pointer receiver |
| Otherwise | Value receiver |
| Large struct as a parameter | Pointer when performance is measurable |
| Otherwise parameter / return | Value |

Within a type, stay consistent: all methods with the same receiver style.

## Interfaces

Defined by the caller, kept small, often a single method.

```go
func Process(r io.Reader) error
func NewReader() *bufio.Reader
```

Accept interfaces, return structs. Reuse standard interfaces like `io.Reader`, `io.Writer`, `io.Closer`.

## Error handling

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething: %w", err)
}
```

- Error is the last return value
- `errors.New` for static strings, `fmt.Errorf` with `%w` to wrap
- `errors.Is` for values, `errors.As` for types
- `errors.Join` for multiple errors
- Sentinel errors used sparingly
- `panic` only for unrecoverable states

Error strings are lowercase, no period, no newline.

## Context

```go
func DoWork(ctx context.Context, in string) (string, error) {
    // ...
}
```

- First parameter, always `ctx context.Context`
- `context.Background()` as the root, `context.TODO()` as a placeholder
- Values in context only for metadata
- Key type unexported, often `struct{}`
- `WithCancel`, `WithTimeout`, `WithDeadline` with `defer cancel()`
- HTTP client and server respect context cancellation

## defer

Cleanup right after resource acquisition.

```go
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
```

Multiple `defer` calls run in LIFO order. `defer` with a named return is the standard wrapping pattern for errors.

## Slices and maps

```go
// slice with pre-allocated capacity
xs := make([]int, 0, n)

// map as a set
seen := map[string]struct{}{}
seen["foo"] = struct{}{}

// comma-ok for missing keys
v, ok := m["key"]
```

`append` result always reassigned. The three-index slice expression `x[:n:n]` caps the capacity when independence is required.

## Goroutines and channels

```go
go func() {
    for v := range in {
        out <- process(v)
    }
}()
```

- Unbuffered channels as the default
- Sender closes, receiver reads with comma-ok
- `select` for multiple channels
- for-select with a done channel for shutdown
- No channels or mutexes in public APIs

## Tests

```go
func TestThing(t *testing.T) {
    t.Run("subtest", func(t *testing.T) {
        // ...
    })
}
```

- File ends in `_test.go`, same package
- `t.Error` to accumulate, `t.Fatal` to stop immediately
- `t.Cleanup`, `t.TempDir`, `t.Setenv` for test resources
- Table tests with a slice and `t.Run`
- `testdata/` for fixtures
- `go test -cover` and `go test -fuzz=...`

## Types as documentation

```go
type Percentage int
type UserID string

func ApplyDiscount(p Percentage, id UserID) ...
```

Named types instead of bare `int` and `string` make function signatures more expressive.

## Zero-value design

Types are designed so that the zero value is usable. `sync.Mutex{}` works without a `New...` function. So does `bytes.Buffer{}`.

## No

- No inheritance, embedding instead
- No `try/catch`, `error` return instead
- No `class`, structs with methods instead
- No operator overloading
- No implicit constructor
- No method type parameters, no method chaining for generic methods
- No implicit conversion between types
- No truthiness, no `null` (only `nil` for pointer-shaped types)
