---
title: "Idiomatic Go in a Nutshell"
description: "Eine Sammlung der wichtigsten Konventionen aus Learning Go. Naming, Receiver, Interfaces, Error Handling, Context, defer, Slices und Maps, Tests. Querverweis-Post zu allen vorherigen Kapiteln."
tags: [go, learning, conventions, idiomatic]
date: 2026-05-18
---

## Worum es geht

Sammlung der Konventionen aus den Kapiteln 2 bis 15 von Learning Go, kompakt nebeneinander. Kein neues Material, sondern eine Übersicht für später.

## Formatierung und Tooling

```bash
go fmt ./...
go vet ./...
go test ./...
```

Tabs zum Einrücken, Klammern auf derselben Zeile. `gofmt` ist verpflichtend, nicht Geschmackssache. Semicolon Insertion Rule erzwingt das Brace-Layout.

## Naming

- Packages: kurz, lowercase, kein Underscore, kein CamelCase
- Exportiert: erstes Zeichen großgeschrieben
- Receiver: einbuchstabig, nicht `this` oder `self`
- Interfaces: einzelne Methode oft mit `er`-Suffix (`Reader`, `Writer`, `Closer`)
- Error-Variablen: Prefix `Err`, etwa `ErrNotFound`
- Konstanten in `iota`-Blöcken: typisierter Underlying-Typ

## Variablendeklaration

```go
x := 10           // typische Form innerhalb von Funktionen
var x int         // explizit Zero-Wert
var x byte = 20   // expliziter Typ
const x = 10      // Compile-Zeit-Wert
```

Unbenutzte lokale Variablen kompilieren nicht.

## Pointer gegen Value

| Situation | Empfehlung |
|---|---|
| Funktion verändert Receiver | Pointer Receiver |
| Funktion muss `nil` annehmen | Pointer Receiver |
| Sonst | Value Receiver |
| Großes Struct als Parameter | Pointer wenn Performance messbar |
| Sonst Parameter / Return | Value |

Konsistenz innerhalb eines Typs: alle Methoden mit der gleichen Receiver-Art.

## Interfaces

Werden beim Caller definiert, klein gehalten, oft eine Methode.

```go
func Process(r io.Reader) error
func NewReader() *bufio.Reader
```

Accept interfaces, return structs. Standard-Interfaces wie `io.Reader`, `io.Writer`, `io.Closer` wiederverwenden.

## Error Handling

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething: %w", err)
}
```

- Error ist letzter Rückgabewert
- `errors.New` für statische Strings, `fmt.Errorf` mit `%w` zum Wrappen
- `errors.Is` für Werte, `errors.As` für Typen
- `errors.Join` für mehrere Errors
- Sentinel Errors sparsam
- `panic` nur für nicht weiterführbare Zustände

Error-Strings sind kleingeschrieben, ohne Punkt, ohne Newline.

## Context

```go
func DoWork(ctx context.Context, in string) (string, error) {
    // ...
}
```

- Erster Parameter, immer `ctx context.Context`
- `context.Background()` als Wurzel, `context.TODO()` als Platzhalter
- Werte im Context nur für Metadaten
- Key-Typ unexportiert, oft `struct{}`
- `WithCancel`, `WithTimeout`, `WithDeadline` mit `defer cancel()`
- HTTP-Client und -Server respektieren Context-Cancellation

## defer

Cleanup direkt nach der Ressourcen-Beschaffung.

```go
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
```

Mehrere `defer` laufen in LIFO-Reihenfolge. `defer` plus benannter Return ergibt ein Wrapping-Pattern für Errors.

## Slices und Maps

```go
// Slice mit Vorab-Kapazität
xs := make([]int, 0, n)

// Map als Set
seen := map[string]struct{}{}
seen["foo"] = struct{}{}

// Comma-ok für fehlende Keys
v, ok := m["key"]
```

`append` immer wieder zuweisen. Dreiteiliger Slice-Ausdruck `x[:n:n]` deckelt die Kapazität, wenn Unabhängigkeit nötig ist.

## Goroutines und Channels

```go
go func() {
    for v := range in {
        out <- process(v)
    }
}()
```

- Unbuffered Channel als Default
- Sender schließt, Receiver liest mit Comma-ok
- `select` für mehrere Channels
- for-select mit Done-Channel zum Beenden
- Channels und Mutexes nicht in öffentlichen APIs

## Tests

```go
func TestThing(t *testing.T) {
    t.Run("subtest", func(t *testing.T) {
        // ...
    })
}
```

- Datei endet auf `_test.go`, im selben Package
- `t.Error` für sammeln, `t.Fatal` für sofort abbrechen
- `t.Cleanup`, `t.TempDir`, `t.Setenv` für Test-Ressourcen
- Tabletests mit Slice und `t.Run`
- `testdata/` für Fixtures
- `go test -cover` und `go test -fuzz=...`

## Typen als Dokumentation

```go
type Percentage int
type UserID string

func ApplyDiscount(p Percentage, id UserID) ...
```

Benannte Typen statt nackter `int` und `string` machen Funktions-Signaturen aussagekräftig.

## Zero-Value-Design

Datentypen werden so entworfen, dass der Zero-Wert nützlich ist. `sync.Mutex{}` ist verwendbar ohne `New...`-Funktion. `bytes.Buffer{}` ebenso.

## Kein

- Keine Vererbung, stattdessen Embedding
- Kein `try/catch`, stattdessen `error`-Rückgabe
- Kein `class`, stattdessen Structs mit Methoden
- Kein Operator Overloading
- Kein impliziter Konstruktor
- Keine Method-Type-Parameter, kein Method Chaining für generische Methoden
- Kein implicit conversion zwischen Typen
- Kein Truthiness, kein `null` (nur `nil` für Pointer-artige Typen)
