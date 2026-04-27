---
title: "Concurrency in Go: Goroutines, Channels, select"
description: "Notizen zu Kapitel 12 von Learning Go: Goroutines, Channels, unbuffered und buffered, select, for-select, Goroutine-Leaks, Context-Cancellation, Backpressure, Timeouts."
tags: [go, learning, concurrency, goroutines, channels]
date: 2026-04-27
---

## Kapitel 12

Themen: Goroutines, Channels, `select`, Concurrency-Patterns, Cancellation.

## Concurrency gegen Parallelism

Concurrency strukturiert ein Problem in unabhängige Komponenten. Ob diese Komponenten tatsächlich parallel laufen, hängt von Hardware und Algorithmus ab. Mehr Goroutines bedeuten nicht zwangsläufig mehr Geschwindigkeit. Bei In-Memory-Berechnungen kann der Overhead der Channel-Kommunikation den Gewinn übersteigen.

Concurrency lohnt sich, wenn unabhängige Operationen kombiniert werden, vor allem bei IO (Netzwerk, Disk).

## Goroutines

Eine Goroutine ist ein vom Go-Runtime verwalteter, leichtgewichtiger Thread. Creation ist günstig, der initiale Stack klein und wächst dynamisch. Switching passiert im Process, ohne Syscalls.

```go
go someFunction()
```

Rückgabewerte werden ignoriert. Üblich ist eine Closure, die die Concurrency-Logik kapselt und die Business-Funktion aufruft.

```go
go func() {
    for val := range in {
        out <- process(val)
    }
}()
```

Jede Funktion kann als Goroutine laufen, ein `async`-Keyword am Funktionskopf gibt es nicht. Damit entfällt das Function-Coloring-Problem.

## Channels

```go
ch := make(chan int)
```

Channels sind die Verbindung zwischen Goroutines, ein Referenz-Typ. Send mit `ch <- v`, Receive mit `v := <-ch`.

Funktions-Signaturen können die Richtung einschränken.

```go
func produce(out chan<- int) // nur Send
func consume(in <-chan int)  // nur Receive
```

## Unbuffered gegen buffered

Defaultmäßig sind Channels unbuffered. Ein Send blockiert, bis ein Receive erfolgt, und umgekehrt.

```go
ch := make(chan int, 10)
```

Buffered mit Kapazität. Send blockiert nur, wenn der Buffer voll ist. Receive blockiert nur, wenn er leer ist.

Unbuffered ist der Default. Buffered wird gezielt eingesetzt, etwa für eine bekannte Anzahl an Goroutines oder als Backpressure-Mechanismus.

## close, comma-ok, for-range

```go
close(ch)
```

Schließt einen Channel. Schreiben auf einen geschlossenen Channel ist ein Panic. Lesen funktioniert weiter, gibt aber den Zero-Wert zurück.

```go
v, ok := <-ch
```

`ok` ist false, wenn der Channel geschlossen ist und keine Werte mehr drin sind.

for-range über einen Channel läuft, bis dieser geschlossen ist.

```go
for v := range ch {
    fmt.Println(v)
}
```

Konvention: der Sender schließt den Channel. Bei mehreren Sendern hilft eine `sync.WaitGroup`.

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
    fmt.Println("nichts ready")
}
```

`select` wartet, bis einer der Cases möglich ist. Mehrere bereite Cases werden zufällig gewählt. Diese Randomisierung verhindert Starvation.

`select` reduziert auch Deadlock-Risiken bei mehreren Channels, weil die Reihenfolge keine Rolle spielt.

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

Loop, der auf mehreren Channels lauscht und über einen Done-Channel kontrolliert beendet wird. Ein `default` in einem for-select sorgt für eine CPU-fressende Busy-Loop und ist fast immer falsch.

## Goroutine-Leaks und Context

Eine Goroutine, die nie endet, leakt Speicher. Beispiel: ein Generator, der schreibt, aber kein Reader mehr da ist.

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

Bricht der Caller den range-Loop früh ab, hängt die Goroutine auf dem nächsten Send.

Lösung: `context.Context` plus `select`.

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

`cancel()` am Caller schließt den Done-Channel, die Goroutine beendet sich.

## Buffered für bekannte Anzahl

```go
results := make(chan int, conc)
for i := 0; i < conc; i++ {
    go func() {
        v := <-ch
        results <- process(v)
    }()
}
```

Buffer-Größe gleich Anzahl der Goroutines. Jede Goroutine kann schreiben und sofort enden.

## Backpressure mit Buffered Channel

Ein Buffered Channel als Token-Bucket.

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

Passt ein Token in den Channel, wird gearbeitet. Sonst feuert der `default` und gibt einen Fehler zurück. Damit lassen sich gleichzeitige Requests ohne explizite Locks limitieren.

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

Bei Timeout schließt der Context, `Done()` feuert, die Funktion gibt einen Error zurück. Der Buffer für `out` ist 1, damit die Worker-Goroutine nicht hängen bleibt.

## APIs ohne Concurrency-Details

Channels und Mutexes gehören nicht in die öffentliche API. Sonst muss jeder Caller wissen, ob ein Channel buffered ist, wann er geschlossen wird, in welcher Reihenfolge er gelesen werden darf.

## Zusammenfassung

- Concurrency strukturiert ein Problem, ist nicht automatisch schneller
- Goroutines sind günstig, müssen aber sauber beendet werden
- Channels sind Referenz-Typen, Richtung in Signaturen
- Unbuffered ist Default, Buffered mit Grund
- Sender schließt den Channel, Comma-ok beim Receiver
- `select` wählt zufällig zwischen Cases die ready sind
- for-select mit Done-Channel zum Beenden
- Context für Cancellation und Timeouts
- Public APIs ohne Channels oder Mutexes
