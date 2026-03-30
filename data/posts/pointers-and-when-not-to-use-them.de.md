---
title: "Pointer in Go, und wann sie nicht nötig sind"
description: "Notizen zu Kapitel 6 von Learning Go: Pointer-Syntax, Pointer als Signal für Mutabilität, Stack und Heap, Escape Analysis und Mechanical Sympathy."
tags: [go, learning, pointers, memory]
date: 2026-03-30
---

## Kapitel 6

Themen: Pointer-Syntax, Pointer-Konventionen, Pointer und Performance, Stack gegen Heap.

## Syntax

`&` ist der Address-Of-Operator. Er liefert die Speicheradresse eines Werts.

```go
x := 10
p := &x
fmt.Println(p)  // Adresse
fmt.Println(*p) // 10
```

`*` vor einem Pointer dereferenziert ihn. Der Zero-Wert eines Pointers ist `nil`. Eine Dereferenz von `nil` ist ein Panic.

```go
var x *int
fmt.Println(x == nil) // true
fmt.Println(*x)       // panic
```

Go kennt keine Pointer-Arithmetik. Dafür gibt es das `unsafe`-Package, das in normalem Code nicht verwendet wird.

## Andere Sprachen verstecken Pointer

In vielen Sprachen ist eine Klasseninstanz unter der Haube ein Pointer. Wird eine Instanz an eine Methode übergeben und dort ein Feld geändert, ist die Änderung im Original sichtbar. Reassign des Parameters selbst hat keinen Effekt. Genau das beschreibt Pointer-Semantik.

Go gibt die Wahl: Value-Typ oder Pointer. Default ist Value-Typ.

## Pointer signalisieren Mutabilität

Go hat kein Schlüsselwort für Immutability. `const` funktioniert nur für Compile-Zeit-Werte. Stattdessen gilt die Konvention: Wenn eine Funktion einen Pointer nimmt, ist Mutation erlaubt. Value-Typen werden als unverändert behandelt.

```go
func update(x *Thing) {
    x.field = "neu"
}
```

## Pointer als letzter Ausweg

Idiomatisches Go bevorzugt Value-Typen.

```go
// nicht so
func MakeFoo(f *Foo) error {
    f.Field1 = "val"
    f.Field2 = 20
    return nil
}

// so
func MakeFoo() (Foo, error) {
    return Foo{Field1: "val", Field2: 20}, nil
}
```

Ausnahme: Interfaces wie `json.Unmarshal` nehmen einen Pointer, damit die Funktion in den vom Caller allozierten Speicher schreibt.

## nil als "kein Wert"

Pointer können den Unterschied zwischen Zero-Wert und "nicht gesetzt" abbilden. Bei JSON-Unmarshaling mit nullable Feldern ist das idiomatisch. Sonst ist ein Value-Typ plus `bool` meist klarer.

## Maps und Slices verhalten sich pointer-ähnlich

Eine Map ist intern ein Pointer auf eine Struct. Ein Slice ist eine Struct mit Länge, Kapazität und einem Pointer auf das Backing-Array.

Beim Übergeben an eine Funktion wird der Pointer bzw. der Header kopiert. Änderungen am Inhalt wirken im Original. Reassign in der Funktion nicht.

```go
func modSlice(s []int) {
    s[0] = 100         // im Original sichtbar
    s = append(s, 999) // im Original NICHT sichtbar, wenn ein neues Backing-Array entsteht
}
```

Maps eignen sich nicht für öffentliche API-Parameter, weil nichts dokumentiert, welche Keys erwartet werden. Ein Struct ist klarer.

## Stack, Heap, Escape Analysis

Der Compiler entscheidet per Escape Analysis, ob ein Wert auf dem Stack oder Heap landet. Wenn der Compiler nicht garantieren kann, dass ein Pointer den aktuellen Stackframe nicht überlebt, landet der Wert auf dem Heap.

Stack-Allokation ist billig. Heap-Allokation kostet GC-Arbeit.

Konsequenz: weniger Pointer im Code reduziert GC-Druck.

## Mechanical Sympathy

Begriff aus dem Motorsport. Code, der mit der Hardware harmoniert, läuft schneller. RAM ist "random access", sequenzielles Lesen ist trotzdem deutlich schneller.

Ein `[]Foo` (Slice aus Structs) liegt sequenziell im Speicher. Ein `[]*Foo` (Slice aus Pointern) verteilt die Daten zufällig. Das Buch verweist auf Tests mit einem Faktor von etwa 100 beim Zugriff.

Idiomatisches Go bevorzugt automatisch das Layout mit weniger indirektem Speicherzugriff.

## Zusammenfassung

- `&` für Adresse, `*` zum Dereferenzieren
- Zero-Wert eines Pointers ist `nil`, Dereferenz davon ist ein Panic
- Pointer signalisieren Mutabilität an den Leser
- Value-Typen sind der Default, Pointer die Ausnahme
- Maps und Slices tragen interne Pointer, daher das abweichende Verhalten
- Stack-Allokation ist billig, Heap-Allokation kostet GC
- Sequenzieller Speicher ist deutlich schneller als verstreuter
