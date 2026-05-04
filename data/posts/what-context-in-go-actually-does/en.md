---
title: "What context in Go does"
description: "Notes on chapter 14 of Learning Go: the context package, carrying values through middleware, cancellation, timeouts and deadlines, integration with the HTTP client and server."
tags: [go, learning, context, http]
date: 2026-05-04
---

## Chapter 14

Topics: context as an interface, values, cancellation, timeouts, HTTP integration.

## What context is

Context is not a language feature, but an interface from the `context` package plus a convention.

```go
func logic(ctx context.Context, info string) (string, error) {
    // ...
}
```

Convention: context is the first parameter and named `ctx`. Mirrors `error` as the last return value.

Entry points:

```go
ctx := context.Background() // standard entry point
ctx := context.TODO()       // placeholder during development
```

`TODO` marks places where the context source is not yet decided. It should not remain in production code.

## Values with WithValue

Context can carry values. Intended for metadata that flows through APIs, not for regular function arguments.

```go
ctx = context.WithValue(parent, key, value)
v := ctx.Value(key)
```

Every `WithValue` call returns a new context that wraps the previous one. Context is immutable.

## Custom key types

The key in `WithValue` is `any`. A plain string key can collide with keys from other packages. Solution: a custom unexported type.

```go
type userKey int

const (
    _ userKey = iota
    key
)
```

or as an empty struct:

```go
type userKey struct{}
```

That way no other package can produce a key of the same type.

## Naming convention for accessors

Instead of touching `WithValue` and `Value` directly, wrap them.

```go
func ContextWithUser(ctx context.Context, user string) context.Context {
    return context.WithValue(ctx, userKey{}, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
    user, ok := ctx.Value(userKey{}).(string)
    return user, ok
}
```

Naming scheme: `ContextWithXxx` and `XxxFromContext`. Comma-ok matters because `Value` returns `nil` on miss.

Rule: unpack context values inside the handler and pass them to business logic as explicit arguments. Context stays in the middleware layer.

## Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

`WithCancel` returns a new context plus a `CancelFunc`. The function must be called or the context leaks resources. `defer cancel()` is the idiomatic shape.

Inside a goroutine, cancellation is checked through the done channel.

```go
select {
case <-ctx.Done():
    return
case ch <- value:
}
```

Calling `cancel()` more than once is fine, it is idempotent.

## WithCancelCause

Since Go 1.20. Allows attaching a reason for cancellation.

```go
ctx, cancel := context.WithCancelCause(context.Background())
cancel(errors.New("bad status from upstream"))

err := context.Cause(ctx)
```

Only the first `cancel(err)` call is recorded.

## Timeouts and deadlines

```go
ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel()

ctx, cancel := context.WithDeadline(parent, time.Now().Add(3*time.Second))
defer cancel()
```

`WithTimeout` takes a duration, `WithDeadline` an absolute time. Both behave like a cancellable context, with automatic cancellation when the deadline passes.

Child contexts are bounded by their parent. A child with a longer timeout still ends when the parent expires first.

## HTTP client with context

The built-in HTTP client respects context cancellation. A request built with `http.NewRequestWithContext` is aborted automatically on cancel.

```go
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := client.Do(req)
```

## Context in the HTTP server

`net/http` predates the context package, the handler interface has no context argument. Instead it sits on the request.

```go
ctx := req.Context()
req = req.WithContext(ctx)
```

Middleware pattern:

```go
func Middleware(h http.Handler) http.Handler {
    return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        ctx = ContextWithUser(ctx, "alice")
        req = req.WithContext(ctx)
        h.ServeHTTP(rw, req)
    })
}
```

## Cancellation in your own code

In long-running loops, cancellation should be checked periodically.

```go
for {
    if err := context.Cause(ctx); err != nil {
        return partialResult, err
    }
    // continue
}
```

Short functions do not need this.

## Err and Cause

`ctx.Err()` is `nil` when the context is active, otherwise `context.Canceled` or `context.DeadlineExceeded`. `context.Cause(ctx)` returns the original error if the context was canceled with `WithCancelCause`.

## Summary

- Context is an interface, not a hidden thread-local
- Convention: first argument `ctx context.Context`
- `context.Background()` as the entry point, `context.TODO()` as a placeholder
- Values in context only for metadata
- Keys with a custom unexported type
- `ContextWithXxx` and `XxxFromContext` naming
- `WithCancel`, `WithTimeout`, `WithDeadline`, each with `defer cancel()`
- `WithCancelCause` and `context.Cause` for propagating reasons
- Child contexts are bounded by the parent
- HTTP client and server have context integration
- Check `context.Cause(ctx)` periodically in long loops
