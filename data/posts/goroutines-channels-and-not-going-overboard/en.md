---
title: "Concurrency in Go: goroutines, channels, select"
description: "Notes on chapter 12 of Learning Go: goroutines, channels, unbuffered and buffered, select, for-select, goroutine leaks, context cancellation, backpressure, timeouts."
tags: [go, learning, concurrency, goroutines, channels]
date: 2026-04-27
---

## Chapter 12

Topics: goroutines, channels, `select`, concurrency patterns, cancellation.

## Concurrency vs parallelism

Concurrency structures a problem into independent components. Whether those components actually run in parallel depends on hardware and algorithm. More goroutines do not automatically mean more speed. For in-memory computations, the overhead of channel communication can outweigh the gain.

Concurrency pays off when combining independent operations, especially with IO (network, disk).

## Goroutines

A goroutine is a lightweight thread managed by the Go runtime. Creation is cheap, the initial stack is small and grows dynamically. Switching happens inside the process, without syscalls.

```go
go someFunction()
```

Return values are ignored. The common pattern is a closure that wraps the concurrency logic and calls the business function.

```go
go func() {
    for val := range in {
        out <- process(val)
    }
}()
```

Any function can run as a goroutine. There is no `async` keyword on the function. This removes the function-coloring problem.

## Channels

```go
ch := make(chan int)
```

Channels are the link between goroutines, a reference type. Send with `ch <- v`, receive with `v := <-ch`.

Function signatures can constrain direction.

```go
func produce(out chan<- int) // send only
func consume(in <-chan int)  // receive only
```

## Unbuffered vs buffered

By default channels are unbuffered. A send blocks until a receive happens, and vice versa.

```go
ch := make(chan int, 10)
```

Buffered with a capacity. Send blocks only when the buffer is full. Receive blocks only when it is empty.

Unbuffered is the default. Buffered is used deliberately, for example with a known number of goroutines or as a backpressure mechanism.

## close, comma-ok, for-range

```go
close(ch)
```

Closes a channel. Writing to a closed channel panics. Reading still works but returns the zero value.

```go
v, ok := <-ch
```

`ok` is false when the channel is closed and drained.

for-range over a channel runs until it is closed.

```go
for v := range ch {
    fmt.Println(v)
}
```

Convention: the sender closes the channel. With multiple senders, a `sync.WaitGroup` helps.

## select

```go
select {
case v := <-ch1:
    fmt.Println(v)
case v := <-ch2:
    fmt.Println(v)
case ch3 <- x:
    fmt.Println("wrote", x)
default:
    fmt.Println("nothing ready")
}
```

`select` waits until one of the cases can proceed. Multiple ready cases are chosen at random. The randomization prevents starvation.

`select` also reduces deadlock risk across multiple channels because order does not matter.

## for-select

```go
for {
    select {
    case <-done:
        return
    case v := <-ch:
        process(v)
    }
}
```

A loop listening on multiple channels with an explicit done channel for shutdown. A `default` inside a for-select produces a CPU-burning busy loop and is almost always wrong.

## Goroutine leaks and context

A goroutine that never exits leaks memory. Example: a generator that writes but no reader is left.

```go
func countTo(max int) <-chan int {
    ch := make(chan int)
    go func() {
        for i := 0; i < max; i++ {
            ch <- i
        }
        close(ch)
    }()
    return ch
}
```

If the caller breaks out of the range loop early, the goroutine is stuck on the next send.

Fix: `context.Context` plus `select`.

```go
func countTo(ctx context.Context, max int) <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < max; i++ {
            select {
            case <-ctx.Done():
                return
            case ch <- i:
            }
        }
    }()
    return ch
}
```

`cancel()` at the caller closes the done channel, the goroutine exits.

## Buffered for a known count

```go
results := make(chan int, conc)
for i := 0; i < conc; i++ {
    go func() {
        v := <-ch
        results <- process(v)
    }()
}
```

Buffer size equals the number of goroutines. Each goroutine can write and exit.

## Backpressure with a buffered channel

A buffered channel as a token bucket.

```go
type PressureGauge struct {
    ch chan struct{}
}

func New(limit int) *PressureGauge {
    return &PressureGauge{ch: make(chan struct{}, limit)}
}

func (pg *PressureGauge) Process(f func()) error {
    select {
    case pg.ch <- struct{}{}:
        f()
        <-pg.ch
        return nil
    default:
        return errors.New("no more capacity")
    }
}
```

If a token fits, the work runs. Otherwise the `default` fires and returns an error. Limits concurrent requests without explicit locks.

## Timeouts

```go
func timeLimit[T any](worker func() T, limit time.Duration) (T, error) {
    out := make(chan T, 1)
    ctx, cancel := context.WithTimeout(context.Background(), limit)
    defer cancel()

    go func() {
        out <- worker()
    }()

    select {
    case result := <-out:
        return result, nil
    case <-ctx.Done():
        var zero T
        return zero, errors.New("work timed out")
    }
}
```

On timeout the context closes, `Done()` fires, the function returns an error. The buffer for `out` is 1 so the worker goroutine does not hang.

## APIs without concurrency details

Channels and mutexes do not belong in the public API. Otherwise every caller has to know whether a channel is buffered, when it is closed, and in what order it can be read.

## Summary

- Concurrency structures a problem, it is not automatic speed
- Goroutines are cheap but must be cleanly terminated
- Channels are reference types, direction belongs in signatures
- Unbuffered is the default, buffered for a reason
- Sender closes the channel, receiver uses comma-ok
- `select` picks randomly among ready cases
- for-select with a done channel for shutdown
- Context for cancellation and timeouts
- Public APIs without channels or mutexes
