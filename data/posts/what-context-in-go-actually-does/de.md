---
title: "Was Context in Go macht"
description: "Notizen zu Kapitel 14 von Learning Go: das context-Package, Werte tragen durch Middleware, Cancellation, Timeouts und Deadlines, Integration mit dem HTTP-Client und -Server."
tags: [go, learning, context, http]
date: 2026-05-04
---

## Kapitel 14

Themen: Context als Interface, Werte, Cancellation, Timeouts, HTTP-Integration.

## Was Context ist

Context ist keine Sprachfunktion, sondern ein Interface aus dem `context`-Package und eine Konvention.

```go
func logic(ctx context.Context, info string) (string, error) {
    // ...
}
```

Konvention: Context ist der erste Parameter und heißt `ctx`. Analog zu `error` als letztem Rückgabewert.

Einstiegspunkte:

```go
ctx := context.Background() // Standard-Einstieg
ctx := context.TODO()       // Platzhalter während Entwicklung
```

`TODO` markiert Stellen, an denen die Context-Herkunft noch ungeklärt ist. In Production-Code sollte das nicht stehen bleiben.

## Werte mit WithValue

Context kann Werte tragen. Gedacht ist das für Metadaten, die durch APIs laufen, nicht für reguläre Funktionsargumente.

```go
ctx = context.WithValue(parent, key, value)
v := ctx.Value(key)
```

Jeder `WithValue`-Aufruf erzeugt einen neuen Context, der den alten umwickelt. Context ist immutable.

## Eigene Key-Typen

Der Key in `WithValue` ist `any`. Ein einfacher String-Key kann mit Keys aus anderen Packages kollidieren. Lösung: ein eigener unexportierter Typ.

```go
type userKey int

const (
    _ userKey = iota
    key
)
```

oder als leere Struct:

```go
type userKey struct{}
```

Damit kann kein anderes Package einen Key mit demselben Typ erzeugen.

## Konvention für Accessor-Funktionen

Statt `WithValue` und `Value` direkt zu nutzen, werden Wrapper definiert.

```go
func ContextWithUser(ctx context.Context, user string) context.Context {
    return context.WithValue(ctx, userKey{}, user)
}

func UserFromContext(ctx context.Context) (string, bool) {
    user, ok := ctx.Value(userKey{}).(string)
    return user, ok
}
```

Namensschema: `ContextWithXxx` und `XxxFromContext`. Comma-ok ist sinnvoll, weil `Value` ohne Treffer `nil` zurückgibt.

Regel: Werte aus dem Context im Handler auspacken und als explizite Argumente an die Business-Logik übergeben. Context bleibt in der Middleware-Schicht.

## Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

`WithCancel` liefert einen neuen Context und eine `CancelFunc`. Diese Funktion muss aufgerufen werden, sonst leakt der Context Ressourcen. `defer cancel()` ist die idiomatische Form.

In einer Goroutine wird auf Cancellation über den Done-Channel geprüft.

```go
select {
case <-ctx.Done():
    return
case ch <- value:
}
```

Mehrfaches `cancel()` ist erlaubt, der Aufruf ist idempotent.

## WithCancelCause

Seit Go 1.20. Erlaubt einen Grund für die Cancellation.

```go
ctx, cancel := context.WithCancelCause(context.Background())
cancel(errors.New("bad status from upstream"))

err := context.Cause(ctx)
```

Nur der erste `cancel(err)`-Aufruf zählt.

## Timeouts und Deadlines

```go
ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel()

ctx, cancel := context.WithDeadline(parent, time.Now().Add(3*time.Second))
defer cancel()
```

`WithTimeout` nimmt eine Dauer, `WithDeadline` einen Zeitpunkt. Beide verhalten sich wie ein cancellable Context, mit automatischer Cancellation bei Fristablauf.

Child-Contexts sind durch ihren Parent gedeckelt. Ein Child mit längerem Timeout endet trotzdem, wenn der Parent vorher ausläuft.

## HTTP-Client mit Context

Der eingebaute HTTP-Client respektiert Context-Cancellation. Ein Request, der mit `http.NewRequestWithContext` gebaut wurde, wird beim Cancel automatisch abgebrochen.

```go
req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
resp, err := client.Do(req)
```

## Context im HTTP-Server

`net/http` ist älter als das context-Package, das Handler-Interface hat kein Context-Argument. Stattdessen am Request.

```go
ctx := req.Context()
req = req.WithContext(ctx)
```

Middleware-Pattern:

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

## Cancellation im eigenen Code

In lang laufenden Schleifen sollte regelmäßig auf Cancellation geprüft werden.

```go
for {
    if err := context.Cause(ctx); err != nil {
        return partialResult, err
    }
    // weiter
}
```

Bei kurzen Funktionen ist das nicht nötig.

## Err und Cause

`ctx.Err()` ist `nil`, wenn der Context aktiv ist, sonst `context.Canceled` oder `context.DeadlineExceeded`. `context.Cause(ctx)` liefert den ursprünglichen Error, wenn der Context mit `WithCancelCause` cancelt wurde.

## Zusammenfassung

- Context ist ein Interface, kein versteckter ThreadLocal
- Konvention: erstes Argument `ctx context.Context`
- `context.Background()` als Einstieg, `context.TODO()` als Platzhalter
- Werte im Context nur für Metadaten
- Keys mit eigenem unexportiertem Typ
- `ContextWithXxx` und `XxxFromContext` als Naming-Konvention
- `WithCancel`, `WithTimeout`, `WithDeadline` jeweils mit `defer cancel()`
- `WithCancelCause` und `context.Cause` für Fehler-Propagation
- Child-Contexts sind durch ihren Parent gedeckelt
- HTTP-Client und -Server haben Context-Integration
- In langen Loops `context.Cause(ctx)` periodisch prüfen
