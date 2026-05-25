---
title: "Go: Merksätze und Mental Model"
description: "Querschnitt-Notizen aus Learning Go. Stack gegen Heap, Call by Value, Zero Values, Pointer-Semantik bei Maps und Slices, Mechanical Sympathy, GC, Concurrency-Faustregeln."
tags: [go, learning, mental-model, performance]
date: 2026-05-25
---

## Worum es geht

Verdichtete Merksätze und Mechanik-Regeln aus den Kapiteln zu Pointern, Maps, Slices, Concurrency und Garbage Collection. Diese Sätze haben sich beim Lesen mehrfach wiederholt.

## Stack und Heap

- Stack ist kostenlos. Allokation ist nur ein Pointer-Inkrement.
- Heap kostet GC. Jede Heap-Allokation muss später wieder eingesammelt werden.
- Der Compiler entscheidet per Escape Analysis, wo ein Wert landet.
- Ein Wert escaped auf den Heap, wenn der Compiler nicht garantieren kann, dass kein Pointer den Stackframe überlebt.
- Weniger Pointer im Code bedeutet weniger Heap-Allokationen, bedeutet weniger GC-Druck.

## Memory Layout

- `[]Foo` liegt sequenziell im Speicher. Schneller Zugriff durch Pre-Fetching.
- `[]*Foo` verstreut Daten im RAM. Pointer-Chasing kostet Größenordnungen.
- Mechanical Sympathy: Code, der die Hardware respektiert, ist schneller. Sequenziell schlägt zufällig.

## Call by Value

- Alles ist Call by Value. Funktionen erhalten Kopien.
- Bei Primitiven und Structs: das Original bleibt unverändert.
- Maps und Slices verhalten sich pointer-artig, weil sie intern Pointer enthalten.
- Eine Map ist intern ein Pointer auf eine Struct.
- Ein Slice ist ein Header aus Länge, Kapazität und Pointer auf das Backing-Array.

## Slice-Verhalten

- `append` kann ein neues Backing-Array allozieren. Ergebnis immer wieder zuweisen.
- Slicing teilt Speicher. Änderungen wirken in beiden Variablen.
- Dreiteiliger Slice-Ausdruck `x[:n:n]` deckelt die Kapazität.
- `make([]T, 0, n)` vermeidet wiederholte Realloziierungen, wenn die Größe bekannt ist.

## Map-Verhalten

- Zero-Wert ist `nil`. Lesen aus nil-Map liefert Zero-Wert. Schreiben in nil-Map panict.
- Maps lassen sich nicht mit `==` vergleichen.
- Keys müssen vergleichbar sein. Slices und Maps fallen raus.
- `map[T]bool` oder `map[T]struct{}` als Set-Ersatz.
- Maps sind nicht thread-safe. Concurrency erfordert `sync.RWMutex` oder `sync.Map`.

## Zero Values

- Jede Variable hat einen Zero-Wert.
- `0`, `""`, `false`, `nil` je nach Typ.
- Datentypen werden so entworfen, dass der Zero-Wert nützlich ist.
- `sync.Mutex{}`, `bytes.Buffer{}`, `[]int(nil)` und `map[string]int(nil)` sind in vielen Operationen direkt verwendbar.

## Pointer und nil

- Zero-Wert eines Pointers ist `nil`. Dereferenz von `nil` ist ein Panic.
- Pointer-Receiver kann nil-Receiver akzeptieren, wenn die Methode den Fall behandelt.
- Eine Interface-Variable ist nur `nil`, wenn Typ und Wert `nil` sind.
- Niemals einen nil-konkreten-Typ in einem Interface zurückgeben.

## Method Sets

- Pointer-Receiver gehört zum Method Set des Pointer-Typs.
- Value-Receiver gehört zu beiden Method Sets (Value und Pointer).
- Konsequenz: nur ein `*T` erfüllt ein Interface, wenn dieses eine Pointer-Receiver-Methode verlangt.

## Concurrency

- Concurrency ist Struktur, nicht automatisch Performance.
- Goroutines sind günstig, aber jede muss am Ende sauber beendet werden.
- Channels sind synchron by default. Ein Send blockiert, bis ein Receive da ist.
- Buffered Channels werden gezielt eingesetzt: bekannte Anzahl oder Backpressure.
- Sender schließt den Channel.
- `select` wählt zufällig zwischen ready Cases. Verhindert Starvation.
- Ein `default` in einem for-select ist fast immer falsch (Busy-Loop).

## Context-Cancellation

- Context propagiert Cancellation und Deadlines durch die Aufrufkette.
- `defer cancel()` direkt nach `context.WithCancel/Timeout/Deadline`.
- Mehrfaches `cancel()` ist idempotent.
- Child-Contexts sind durch ihren Parent gedeckelt.
- Lang laufende Loops prüfen periodisch `context.Cause(ctx)`.

## Error Handling

- Errors sind Werte, keine Exceptions.
- Error ist der letzte Rückgabewert. `nil` heißt Erfolg.
- Wrappen mit `%w`. Auspacken mit `errors.Is` und `errors.As`.
- `panic` nur für nicht weiterführbare Zustände.
- `recover` nur am Rand einer Library-API.

## Tooling

- `go fmt` ist verpflichtend, nicht Geschmackssache.
- `go vet` fängt typische Bugs (Shadowing, falsche Printf-Formate).
- Unbenutzte Imports und Variablen kompilieren nicht.
- `_test.go` Dateien gehören neben den Code, im selben Package.
- Standard-Library-Schnittstellen wiederverwenden (`io.Reader`, `io.Writer`).

## Generelle Haltung

- Lesbarkeit schlägt Cleverness. Klarer Datenfluss ist wichtiger als kurze Zeilen.
- Explizit schlägt implizit. Keine versteckten Konvertierungen, kein Truthiness.
- Konventionen statt Sprachfunktionen. `if err != nil` statt Exceptions, Pointer als Mutability-Signal statt Immutability-Keyword.
- Kleine, fokussierte Interfaces. Beim Caller definiert.
- Concurrency als Strukturierungs-Werkzeug, nicht als Performance-Hammer.
