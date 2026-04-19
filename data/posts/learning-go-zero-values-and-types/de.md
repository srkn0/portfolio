---
title: "Zero Values, explizite Typen und der Compiler"
description: "Notizen zu Kapitel 2 von Learning Go: eingebaute Typen, Variablendeklarationen, Zero Values, explizite Typkonvertierung, Truthiness und Konstanten."
tags: [go, learning, fundamentals]
date: 2026-03-09
---

## Kapitel 2

Themen: eingebaute Typen, Variablendeklarationen, Zero Values, Typkonvertierung, Konstanten, unbenutzte Variablen.

## Zero Values

Jede deklarierte Variable in Go bekommt automatisch einen Zero-Wert, auch ohne explizite Zuweisung.

```go
var x int
var s string
var flag bool
fmt.Println(x)    // 0
fmt.Println(s)    // "" (leerer String)
fmt.Println(flag) // false
```

Es gibt kein `undefined`, kein `null`. Der Default-Wert ist je nach Typ definiert.

## Keine automatische Typkonvertierung

Go konvertiert numerische Typen nicht automatisch. Ein `int` und ein `float64` lassen sich nur addieren, wenn explizit konvertiert wird.

```go
var x int = 10
var y float64 = 30.2
var sum float64 = float64(x) + y
```

## Kein Truthiness

Bedingungen müssen einen booleschen Wert ergeben. Ein leerer String oder `0` wird nicht implizit als `false` behandelt.

```go
if s == "" {
    // String ist leer
}

if x == 0 {
    // Zahl ist Null
}
```

`if s` allein kompiliert nicht.

## `var` gegen `:=`

Zwei Formen für Variablendeklarationen.

```go
var x int = 10
```

und

```go
x := 10
```

Bei `:=` wird der Typ abgeleitet. Innerhalb von Funktionen ist das die übliche Form. `var` ist sinnvoll, wenn der Zero-Wert explizit genutzt wird:

```go
var x int
```

oder wenn ein bestimmter Typ wichtig ist:

```go
var x byte = 20
```

Lesbarer als `x := byte(20)`.

## Konstanten

Konstanten werden zur Compile-Zeit ausgewertet. Keine Funktionsaufrufe, keine Laufzeit-Berechnungen.

```go
const x = 10        // ok
const y = 20 * 10   // ok
const z = len("hi") // ok (Built-in)

a := 5
const b = a + 1     // kompiliert nicht, a ist eine Variable
```

Es gibt keinen Mechanismus, einen zur Laufzeit berechneten Wert unveränderlich zu machen.

## Unbenutzte Variablen

Eine lokale Variable, die deklariert, aber nie gelesen wird, ist ein Compile-Fehler.

```go
func main() {
    x := 10  // compile error: x declared but not used
}
```

## Zusammenfassung

- Jede Variable hat einen Zero-Wert, kein undefined
- Typkonvertierung ist immer explizit
- Bedingungen brauchen echte Booleans
- `:=` innerhalb von Funktionen, `var` für Zero-Wert oder expliziten Typ
- Konstanten nur zur Compile-Zeit
- Unbenutzte lokale Variablen sind ein Compile-Fehler
