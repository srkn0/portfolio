---
title: "Errors as values"
description: "Notes on chapter 9 of Learning Go: the error interface, sentinel errors, custom error types, the nil-interface trap, error wrapping with %w, errors.Is, errors.As, errors.Join, panic and recover."
tags: [go, learning, errors, error-handling]
date: 2026-04-20
---

## Chapter 9

Topics: the `error` interface, construction, sentinel errors, wrapping, `errors.Is`, `errors.As`, `panic` and `recover`.

## The interface

```go
type error interface {
    Error() string
}
```

Anything with this one method is an error. Returning `nil` means: no error.

```go
func calcRemainderAndMod(num, denom int) (int, int, error) {
    if denom == 0 {
        return 0, 0, errors.New("denominator is 0")
    }
    return num / denom, num % denom, nil
}
```

Convention for error strings: lowercase, no trailing period, no newline. Makes sense once wrapping is involved. Multiple fragments stay readable.

## errors.New and fmt.Errorf

```go
return 0, errors.New("only even numbers are processed")

return 0, fmt.Errorf("%d isn't an even number", i)
```

`fmt.Errorf` for runtime values in the error string, `errors.New` for static text.

## Sentinel errors

A sentinel error is a predefined error value at package level. Callers compare with `==`.

```go
if err == io.EOF {
    // end of data
}
```

Naming convention: `ErrXxx`, `io.EOF` is the well-known exception. Sentinel errors are part of the public API and must be maintained. Use sparingly, only when a specific state needs to be communicated without extra data.

## Custom error types

```go
type Status int

const (
    InvalidLogin Status = iota + 1
    NotFound
)

type StatusErr struct {
    Status  Status
    Message string
}

func (se StatusErr) Error() string {
    return se.Message
}
```

Usage:

```go
return nil, StatusErr{
    Status:  InvalidLogin,
    Message: fmt.Sprintf("invalid credentials for %s", uid),
}
```

## The nil-interface trap

```go
func GenerateErrorBroken(flag bool) error {
    var genErr StatusErr // concrete type, not error
    if flag {
        genErr = StatusErr{Status: NotFound}
    }
    return genErr
}
```

When `flag` is false, `genErr` is the zero value of `StatusErr`. The return type is `error`, so the concrete type is boxed into an interface. The interface is not nil. Caller code doing `if err != nil` finds an error that should not exist.

Fix: either explicit `return nil` or declare the local variable as `error`.

```go
func GenerateErrorOK(flag bool) error {
    if flag {
        return StatusErr{Status: NotFound}
    }
    return nil
}
```

## Error wrapping with %w

```go
func fileChecker(name string) error {
    f, err := os.Open(name)
    if err != nil {
        return fmt.Errorf("in fileChecker: %w", err)
    }
    f.Close()
    return nil
}
```

`%w` wraps the underlying error. `errors.Unwrap` peels it back but is rarely called directly.

For plain text embedding without identity: `%v` instead of `%w`.

## errors.Is

```go
if errors.Is(err, os.ErrNotExist) {
    fmt.Println("file doesn't exist")
}
```

`errors.Is` walks the error tree and checks each layer against the sentinel value.

## errors.As

```go
var statusErr StatusErr
if errors.As(err, &statusErr) {
    fmt.Println(statusErr.Status, statusErr.Message)
}
```

Pointer to a variable of the desired type. If the tree contains a matching error, the variable is populated.

Rule: `errors.Is` for specific values, `errors.As` for specific types.

## Joining multiple errors

```go
func ValidatePerson(p Person) error {
    var errs []error
    if len(p.FirstName) == 0 {
        errs = append(errs, errors.New("FirstName is required"))
    }
    if len(p.LastName) == 0 {
        errs = append(errs, errors.New("LastName is required"))
    }
    if p.Age < 0 {
        errs = append(errs, errors.New("Age cannot be negative"))
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

`errors.Join` is useful when multiple errors arise in one step and all should be reported.

## defer for wrapping

When every error in a function gets wrapped with the same message, `defer` plus a named return collapses the repetition.

```go
func DoSomeThings(val1 int, val2 string) (_ string, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("in DoSomeThings: %w", err)
        }
    }()
    val3, err := doThing1(val1)
    if err != nil {
        return "", err
    }
    val4, err := doThing2(val2)
    if err != nil {
        return "", err
    }
    return doThing3(val3, val4)
}
```

## panic and recover

`panic` is not the way for ordinary errors. It signals a state in which the runtime cannot continue: index out of range, division by zero, nil dereference.

`recover` catches a panic and must be inside a `defer`.

```go
func div60(i int) {
    defer func() {
        if v := recover(); v != nil {
            fmt.Println(v)
        }
    }()
    fmt.Println(60 / i)
}
```

`recover` is typically used only at the boundary of a library API to keep panics out of the public surface. In application code, explicit error returns remain the standard.

## Stack traces

Standard errors do not carry a stack trace. Third-party libraries like Cockroach's wrap errors with traces. Otherwise the "trail" is built up via wrapping by hand.

## Summary

- `error` is an interface with `Error() string`
- Error strings lowercase, no period, no newline
- Sentinel errors used sparingly
- Custom error types are structs with `Error() string`
- Avoid local variables of a concrete error type, use `error`
- `%w` to wrap, `%v` for text embedding
- `errors.Is` for values, `errors.As` for types
- `errors.Join` for multiple errors
- `defer` plus named return for uniform wrapping
- `panic` and `recover` are not an exception substitute
- Stack traces via third-party libraries
