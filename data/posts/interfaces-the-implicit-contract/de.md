---
title: "Typen, Methoden, Interfaces"
description: "Notizen zu Kapitel 7 von Learning Go: benannte Typen, Methoden mit Receivern, Embedding statt Vererbung, implizite Interfaces, accept interfaces return structs, nil-Interface-Falle, Type Assertions und Type Switches."
tags: [go, learning, interfaces, types]
date: 2026-04-06
---

## Kapitel 7

Themen: benannte Typen, Methoden, Embedding, Interfaces, Type Assertions.

## Benannte Typen

Jeder Underlying-Typ kann einen eigenen Namen bekommen.

```go
type Score int
type Converter func(string) Score
type TeamScores map[string]Score
```

`Score` hat `int` als Underlying-Typ, ist aber ein eigenständiger Typ. Eine Zuweisung eines `int` an `Score` erfordert eine explizite Konvertierung. Der Vorteil: der Typ dokumentiert die Bedeutung des Werts.

## Methoden

Methoden sind Funktionen mit einem Receiver zwischen `func` und Name.

```go
type Person struct {
    FirstName string
    LastName  string
}

func (p Person) FullName() string {
    return p.FirstName + " " + p.LastName
}
```

Konvention für den Receiver-Namen: kurz, üblicherweise der erste Buchstabe des Typs. `this` oder `self` ist nicht idiomatisch.

Methoden müssen im selben Package wie der Typ deklariert sein. Typen aus fremden Packages lassen sich nicht erweitern.

## Value Receiver gegen Pointer Receiver

```go
func (c Counter) String() string { ... }  // value receiver
func (c *Counter) Increment()    { ... }  // pointer receiver
```

Regeln:

- Wenn die Methode den Receiver verändert, Pointer Receiver
- Wenn die Methode mit einem nil-Receiver umgehen können soll, Pointer Receiver
- Sonst Value Receiver

Hat ein Typ mindestens eine Pointer-Receiver-Methode, werden meist alle Methoden mit Pointer Receiver geschrieben.

Go macht beim Aufruf automatisch die nötige Konvertierung.

```go
var c Counter
c.Increment() // wird zu (&c).Increment()
```

Die Method Sets bleiben strikt: Eine Pointer-Receiver-Methode gehört nur zum Method Set des Pointer-Typs. Das wird bei Interfaces relevant.

## Methoden für nil-Receiver

Ein Methodenaufruf auf einem nil-Pointer-Receiver kann funktionieren, wenn die Methode den Fall behandelt. Auf einem nil-Value-Receiver gibt es einen Panic.

Beispiel aus dem Buch: ein Binary Tree, der `nil` als leeren Baum interpretiert.

```go
func (t *IntTree) Insert(val int) *IntTree {
    if t == nil {
        return &IntTree{val: val}
    }
    if val < t.val {
        t.left = t.left.Insert(val)
    } else if val > t.val {
        t.right = t.right.Insert(val)
    }
    return t
}
```

## Keine Vererbung

Go hat kein `class` und kein `extends`. `type HighScore Score` erzeugt einen neuen Typ mit dem gleichen Underlying-Typ. Methoden von `Score` gehören nicht zu `HighScore`. Eine Substitution ist nicht möglich.

## Composition durch Embedding

Anstelle von Vererbung gibt es Embedding. Ein Struct kann andere Typen als anonyme Felder enthalten.

```go
type Employee struct {
    Name string
    ID   string
}

func (e Employee) Description() string {
    return e.Name + " (" + e.ID + ")"
}

type Manager struct {
    Employee
    Reports []Employee
}
```

Felder und Methoden des eingebetteten Typs sind direkt am äußeren Typ erreichbar.

```go
m := Manager{
    Employee: Employee{Name: "Bob", ID: "12345"},
}
fmt.Println(m.Description()) // Bob (12345)
```

Embedding ist keine Vererbung. Ein `Manager` lässt sich nicht einer `Employee`-Variable zuweisen. Es gibt kein Dynamic Dispatch für konkrete Typen.

## Implizite Interfaces

```go
type Stringer interface {
    String() string
}
```

Jeder Typ mit einer `String() string`-Methode erfüllt das Interface. Es ist keine explizite Deklaration nötig, kein `implements`-Keyword.

```go
type Counter struct{ n int }

func (c Counter) String() string {
    return fmt.Sprintf("%d", c.n)
}

// Counter erfüllt fmt.Stringer.
```

Das Buch beschreibt das als Kombination aus zwei Welten. Explizite Interfaces wie in Java machen Anforderungen sichtbar. Duck Typing wie in Python erspart die Bürokratie. In Go deklariert der Caller das Interface, das er benötigt. Die Implementierung muss nur passen.

Interfaces in Go werden klein gehalten, oft mit nur einer Methode. Definiert wird beim Caller, wo das Interface gebraucht wird.

## Accept interfaces, return structs

Idiomatisches Pattern. Funktionen, die Arbeit erledigen, nehmen Interfaces. Funktionen, die Werte produzieren, geben konkrete Typen zurück.

```go
func Process(r io.Reader) error
func NewReader() *bufio.Reader
```

Vorteil: Caller kann beliebige Implementierungen reinreichen. Produzenten-Funktionen lassen sich erweitern, ohne Interfaces zu brechen.

## Die nil-Interface-Falle

Interfaces sind intern ein Paar aus Typ-Pointer und Wert-Pointer. Das Interface ist nur dann `nil`, wenn beide `nil` sind.

```go
var pc *Counter        // nil pointer
var i Incrementer      // nil interface
fmt.Println(pc == nil) // true
fmt.Println(i == nil)  // true

i = pc
fmt.Println(i == nil)  // false
```

Konsequenz: einen nil-konkreten Typ nicht in ein Interface verpacken und zurückgeben.

```go
func getCounter() Incrementer {
    var pc *Counter
    return pc // Interface ist nicht nil
}
```

## Type Assertions und Type Switches

```go
var i any = 42
n, ok := i.(int)
if !ok {
    // war kein int
}
```

Die Comma-ok-Form macht Type Assertions sicher. Ohne `ok` panict eine fehlgeschlagene Assertion.

Bei mehreren möglichen Typen ein Type Switch.

```go
switch v := i.(type) {
case int:
    fmt.Println("int:", v)
case string:
    fmt.Println("string:", v)
default:
    fmt.Println("anderer Typ")
}
```

## any (das leere Interface)

`any` ist ein Alias für `interface{}`. Variablen vom Typ `any` können beliebige Werte aufnehmen.

```go
data := map[string]any{}
json.Unmarshal(contents, &data)
```

Sinnvoll bei JSON mit unbekanntem Schema. Sonst untergräbt `any` die Typsicherheit, deshalb sparsam einsetzen.

## Zusammenfassung

- Methoden gehören über den Receiver zu einem Typ, kein `this`
- Pointer Receiver beim Mutieren oder bei nil-Receiver-Support
- Embedding ist Composition mit Promotion, keine Vererbung
- Interfaces sind implizit, der Compiler prüft die Methoden
- Interfaces gehören zum Caller, klein halten
- Accept interfaces, return structs
- nil-Interface braucht Typ und Wert auf nil
- Type Assertion mit Comma-ok, Type Switch für mehrere Möglichkeiten
- `any` als Notnagel, kein Default
