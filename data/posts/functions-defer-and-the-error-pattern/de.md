---
title: "Funktionen, defer und das Error-Pattern"
description: "Notizen zu Kapitel 4 und 5 von Learning Go: Blöcke, Shadowing, for, switch, Funktionen mit mehreren Rückgabewerten, das if err != nil Pattern, defer, Closures, call by value."
tags: [go, learning, functions, error-handling]
date: 2026-03-23
---

## Kapitel 4 und 5

Themen: Blöcke und Shadowing, Kontrollfluss, Funktionen, Closures, defer, Call by Value.

## Blöcke und Shadowing

Jedes Paar geschweifter Klammern erzeugt einen Block. Variablen aus einem Block existieren nur innerhalb dieses Blocks.

Eine Variable mit demselben Namen in einem inneren Block verdeckt die äußere.

```go
x := 10
if x > 5 {
    x := 20  // neues x
    fmt.Println(x) // 20
}
fmt.Println(x) // 10
```

Beide `x` sind verschiedene Variablen. Shadowing kann subtile Bugs erzeugen, `go vet` erkennt nicht alle Fälle.

## for-Schleife

Go hat ein einziges Schleifenkonstrukt: `for`. Vier Formen.

```go
// klassisch
for i := 0; i < 10; i++ { }

// nur Bedingung, entspricht while
for x < 100 { }

// endlos
for { }

// for-range über Slices, Maps, Strings
for i, v := range items { }
```

`for-range` über Strings iteriert über Runen, nicht über Bytes. Die Wert-Variable ist eine Kopie, eine Änderung beeinflusst das Original nicht.

## switch

`switch` fällt nicht durch. Nach dem ersten passenden Case bricht die Anweisung automatisch ab.

```go
switch size := len(word); size {
case 1, 2, 3, 4:
    fmt.Println("kurzes Wort")
case 5:
    fmt.Println("genau richtig")
default:
    fmt.Println("langes Wort")
}
```

Mehrere Werte pro Case mit Komma. Switch auf jeden vergleichbaren Typ.

Der Blank-Switch ersetzt eine if/else-Kette.

```go
switch {
case len(word) < 5:
    fmt.Println("kurz")
case len(word) > 10:
    fmt.Println("lang")
default:
    fmt.Println("mittel")
}
```

## Mehrere Rückgabewerte

```go
func divAndRemainder(num, denom int) (int, int, error) {
    if denom == 0 {
        return 0, 0, errors.New("Division durch Null")
    }
    return num / denom, num % denom, nil
}
```

Aufruf:

```go
result, remainder, err := divAndRemainder(5, 2)
if err != nil {
    fmt.Println(err)
    os.Exit(1)
}
fmt.Println(result, remainder)
```

Konvention: Fehler ist der letzte Rückgabewert. `nil` heißt Erfolg.

## Das `if err != nil`-Pattern

```go
f, err := os.Open(filename)
if err != nil {
    log.Fatal(err)
}
defer f.Close()

data, err := os.ReadFile(filename)
if err != nil {
    log.Fatal(err)
}
result, err := strconv.Atoi(string(data))
if err != nil {
    log.Fatal(err)
}
```

Jeder Funktionsaufruf prüft den Fehler. Happy Path und Fehlerbehandlung stehen direkt nebeneinander.

## Funktionen als Werte

Funktionen sind Werte. Sie können Variablen zugewiesen, als Argumente übergeben und zurückgegeben werden.

```go
var opMap = map[string]func(int, int) int{
    "+": func(a, b int) int { return a + b },
    "-": func(a, b int) int { return a - b },
}
```

### Closures

Funktionen, die innerhalb anderer Funktionen deklariert sind, sind Closures. Sie können Variablen der äußeren Funktion lesen und verändern.

```go
func main() {
    a := 20
    f := func() {
        fmt.Println(a)
        a = 30
    }
    f()
    fmt.Println(a) // 30
}
```

Closures werden zum Beispiel an `sort.Slice` übergeben.

```go
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})
```

## defer

`defer` verzögert einen Funktionsaufruf bis die umgebende Funktion endet. Häufiger Use Case: Ressourcen schließen.

```go
f, err := os.Open(filename)
if err != nil {
    log.Fatal(err)
}
defer f.Close()
```

Mehrere `defer`-Aufrufe laufen in LIFO-Reihenfolge.

`defer` lässt sich mit benannten Rückgabewerten kombinieren, etwa für das Cleanup von Datenbanktransaktionen:

```go
func DoSomeInserts(ctx context.Context, db *sql.DB) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() {
        if err == nil {
            err = tx.Commit()
        }
        if err != nil {
            tx.Rollback()
        }
    }()
    _, err = tx.ExecContext(ctx, "INSERT INTO FOO ...")
    return err
}
```

## Call by Value

Alle Werte werden als Kopie übergeben. Funktionen erhalten Kopien, keine Referenzen. Maps und Slices verhalten sich anders, weil sie intern Pointer enthalten.

```go
func modSlice(s []int) {
    for k, v := range s {
        s[k] = v * 2  // verändert das Original
    }
    s = append(s, 10)  // verändert das Original NICHT
}
```

Elemente eines übergebenen Slices lassen sich verändern. Vergrößern geht nicht zuverlässig, weil `append` ein neues Backing-Array allozieren kann, das der Caller nie sieht.

## Zusammenfassung

- Ein Schleifenkonstrukt, vier Formen
- `switch` fällt nicht durch, Blank-Switch ersetzt if/else-Ketten
- Mehrere Rückgabewerte, Fehler zuletzt, `nil` für Erfolg
- `if err != nil` als sichtbarer Kontrollfluss
- `defer` garantiert Cleanup, läuft in LIFO-Reihenfolge
- Funktionen sind Werte, können gespeichert und übergeben werden
- Call by Value, Maps und Slices sind Sonderfälle wegen interner Pointer
