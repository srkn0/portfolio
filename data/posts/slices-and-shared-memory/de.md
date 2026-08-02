---
title: "Slices, Maps, Structs und geteilter Speicher"
description: "Notizen zu Kapitel 3 von Learning Go: Arrays, Slices, Maps und Structs. Mit Fokus auf shared backing storage bei Slices und das Komma-OK-Idiom bei Maps."
tags: [go, learning, data-structures]
date: 2026-03-16
---

## Kapitel 3

Themen: Arrays, Slices, Maps, Structs.

## Arrays

In Go ist die Größe Teil des Typs. `[3]int` und `[4]int` sind verschiedene Typen. Eine Funktion, die `[3]int` akzeptiert, akzeptiert kein `[4]int`.

```go
var x [3]int
var y [4]int
x = y // Compile-Fehler
```

Arrays werden im Alltag selten direkt verwendet. Sie dienen als Backing-Storage für Slices.

## Slices

Slices werden ohne Größe im Typ deklariert.

```go
x := []int{10, 20, 30}
```

Drei interne Werte: Länge, Kapazität, Pointer auf das Backing-Array.

### `append` gibt ein neues Slice zurück

```go
var x []int
x = append(x, 10)
```

Das Ergebnis muss zugewiesen werden. Go ist call-by-value: `append` bekommt eine Kopie des Slice-Headers, hängt an und gibt die neue Version zurück.

### Länge gegen Kapazität

Länge ist die Anzahl belegter Elemente. Kapazität ist die Größe des reservierten Backing-Arrays.

```go
x := make([]int, 0, 10) // Länge 0, Kapazität 10
x = append(x, 1, 2, 3)  // Länge 3, Kapazität 10
```

Geht ein `append` über die Kapazität hinaus, alloziert Go ein neues, größeres Backing-Array und kopiert um. Vorab-Allokation mit `make` reduziert Allokationen und GC-Druck.

### Slicing teilt Speicher

```go
x := []string{"a", "b", "c", "d"}
y := x[:2] // y = ["a", "b"]
x[1] = "z"
fmt.Println(y) // ["a", "z"]
```

Slicing kopiert nicht. Beide Variablen zeigen auf dasselbe Backing-Array.

Mit dem dreiteiligen Ausdruck lässt sich die Kapazität deckeln, sodass ein späteres `append` ein neues Array alloziert.

```go
y := x[:2:2] // Kapazität bei 2 gedeckelt
```

## Maps

```go
teams := map[string]int{
    "Orcas": 1,
    "Lions": 2,
}
```

Wichtige Eigenschaften:

- Zero-Wert ist `nil`. Lesen aus einer nil-Map gibt den Zero-Wert des Wertetyps. Schreiben in eine nil-Map ist ein Panic.
- Maps lassen sich nicht mit `==` vergleichen.
- Keys müssen vergleichbar sein. Slices und Maps sind nicht als Key erlaubt.

### Komma-ok-Idiom

Ein fehlender Key liefert den Zero-Wert ohne Fehler. Um fehlenden Key von echtem Zero-Wert zu unterscheiden, gibt es die Komma-ok-Form.

```go
v, ok := m["key"]
if ok {
    // Key existiert, v hat den Wert
}
```

### Map als Set

Go hat keinen eingebauten Set-Typ. Üblich ist `map[T]bool` oder `map[T]struct{}` für minimalen Speicherverbrauch.

```go
intSet := map[int]bool{}
intSet[5] = true
if intSet[5] {
    fmt.Println("5 ist im Set")
}
```

## Structs

Zum Gruppieren zusammengehöriger Felder.

```go
type Person struct {
    Name string
    Age  int
    Pet  string
}
```

Eigenschaften:

- Struct-Literale können Feldnamen verwenden (empfohlen) oder positional sein.
- Structs sind mit `==` vergleichbar, wenn alle Felder vergleichbar sind.
- Anonyme Structs sind möglich, etwa für einmalige Shapes in Tests.

Go kennt keine Klassen und keine Vererbung. Structs mit Methoden ersetzen diese Konzepte.

## Zusammenfassung

- Arrays existieren, werden im Alltag kaum verwendet
- `append` gibt ein neues Slice zurück, immer wieder zuweisen
- Länge und Kapazität sind zwei verschiedene Werte
- Slicing teilt Speicher, dreiteiliger Ausdruck deckelt die Kapazität
- Komma-ok-Idiom unterscheidet fehlenden Key von Zero-Wert
- Structs gruppieren Daten, keine Vererbung
