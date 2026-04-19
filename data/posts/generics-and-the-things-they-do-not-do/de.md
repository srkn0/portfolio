---
title: "Generics in Go"
description: "Notizen zu Kapitel 8 von Learning Go: Type Parameter, comparable, generische Funktionen, Type Terms mit Union-Syntax, Tilde-Constraint, cmp.Ordered und bewusste Einschränkungen der Go-Generics."
tags: [go, learning, generics]
date: 2026-04-13
---

## Kapitel 8

Themen: Type Parameter, Constraints, generische Funktionen, Type Terms, was Go bei Generics nicht hat.

## Vor Generics

Vor Go 1.18 gab es zwei Wege, eine Datenstruktur für beliebige Typen zu schreiben. Pro Typ eine Kopie oder ein `Orderable`-Interface mit `any`. Beide haben Nachteile.

```go
type Orderable interface {
    Order(any) int
}
```

Mit `any` lassen sich Mischtypen einfügen, die Compile-Zeit-Prüfung greift nicht.

```go
it.Insert(OrderableInt(5))
it.Insert(OrderableString("nope")) // kompiliert, panict zur Laufzeit
```

Generics schließen diese Lücke.

## Syntax

```go
type Stack[T any] struct {
    vals []T
}

func (s *Stack[T]) Push(val T) {
    s.vals = append(s.vals, val)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.vals) == 0 {
        var zero T
        return zero, false
    }
    top := s.vals[len(s.vals)-1]
    s.vals = s.vals[:len(s.vals)-1]
    return top, true
}
```

`[T any]` nach dem Typnamen deklariert den Type Parameter. Danach lässt sich `T` wie ein normaler Typ verwenden. Bei Methoden steht `[T]` am Receiver: `func (s *Stack[T])`.

`var zero T` liefert den Zero-Wert des generischen Typs. `nil` ist nicht möglich, weil `T` auch `int` sein kann.

## Constraints sind Interfaces

`any` erlaubt jeden Typ, schließt aber Operatoren aus. `==` braucht das eingebaute `comparable`.

```go
type Stack[T comparable] struct {
    vals []T
}

func (s Stack[T]) Contains(val T) bool {
    for _, v := range s.vals {
        if v == val {
            return true
        }
    }
    return false
}
```

`comparable` matcht alle Typen, die mit `==` und `!=` verglichen werden können. Slices und Maps gehören nicht dazu.

## Generische Funktionen

```go
func Map[T1, T2 any](s []T1, f func(T1) T2) []T2 {
    r := make([]T2, len(s))
    for i, v := range s {
        r[i] = f(v)
    }
    return r
}

func Filter[T any](s []T, f func(T) bool) []T {
    var r []T
    for _, v := range s {
        if f(v) {
            r = append(r, v)
        }
    }
    return r
}
```

Type Inference funktioniert in den meisten Fällen.

```go
words := []string{"One", "Potato", "Two", "Potato"}
filtered := Filter(words, func(s string) bool {
    return s != "Potato"
})
```

## Type Terms für Operatoren

Für arithmetische Operatoren reicht `any` nicht. Ein Interface mit Type Terms listet die erlaubten Typen auf.

```go
type Integer interface {
    int | int8 | int16 | int32 | int64 |
        uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

func divAndRemainder[T Integer](num, denom T) (T, T, error) {
    if denom == 0 {
        return 0, 0, errors.New("Division durch Null")
    }
    return num / denom, num % denom, nil
}
```

Das `|` ist eine Union. Erlaubt sind die Operatoren, die für alle Type Terms definiert sind.

Type Terms matchen exakt. Für eigene Typen mit einem Underlying-Typ aus der Liste braucht es die Tilde.

```go
type Integer interface {
    ~int | ~int8 | ...
}
```

`~int` matcht jeden Typ, dessen Underlying-Typ `int` ist.

## cmp.Ordered und cmp.Compare

Seit Go 1.21 enthält die Standard-Library das `cmp`-Package mit `cmp.Ordered` und `cmp.Compare`.

```go
import "cmp"

func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}
```

Für einen generischen Tree lässt sich `cmp.Compare[int]` als Ordering-Funktion übergeben.

## Was Generics in Go nicht haben

Bewusste Einschränkungen.

**Kein Operator Overloading.** Eigene Container-Typen unterstützen weder `range` noch `[]`. Begründung: viele Operatoren, kein Method Overloading, und Operator Overloading erschwert die Lesbarkeit.

**Keine Type Parameter an Methoden.** Method Chaining wie `xs.Map(...).Reduce(...)` funktioniert in Go nicht für generische Methoden. Methoden können nur die Type Parameter des Receivers verwenden.

**Keine variadischen Type Parameter.** Eine Funktion kann nicht beliebig viele verschiedene Type Parameter haben.

**Kein Specialization, Currying oder Metaprogramming.**

## Comparable und Laufzeit

`comparable` schließt nicht-vergleichbare Typen zur Compile-Zeit aus. Ein Interface-Wert, dessen konkreter Typ ein Slice ist, kompiliert trotzdem mit einem `comparable`-Constraint und panict bei `==` zur Laufzeit.

## Zusammenfassung

- `[T any]` deklariert einen Type Parameter
- `var zero T` für den Zero-Wert eines generischen Typs
- `comparable` als Constraint für `==`-fähige Typen
- Type Terms (`int | int8 | ...`) erlauben Operator-Support
- `~int` matcht eigene Typen mit `int` als Underlying-Typ
- Type Inference funktioniert in den meisten Fällen
- `cmp.Ordered` und `cmp.Compare` seit Go 1.21
- Kein Operator Overloading, keine Method Type Parameter, kein Method Chaining
- `comparable` schützt nicht vor Laufzeit-Panic bei Interfaces mit nicht-vergleichbarem Underlying-Typ
