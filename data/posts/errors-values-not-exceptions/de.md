---
title: "Errors als Werte"
description: "Notizen zu Kapitel 9 von Learning Go: das error-Interface, Sentinel Errors, eigene Error-Typen, die nil-Interface-Falle, Error Wrapping mit %w, errors.Is, errors.As, errors.Join, panic und recover."
tags: [go, learning, errors, error-handling]
date: 2026-04-20
---

## Kapitel 9

Themen: das `error`-Interface, Konstruktion, Sentinel Errors, Wrapping, `errors.Is`, `errors.As`, `panic` und `recover`.

## Das Interface

```go
type error interface {
    Error() string
}
```

Alles mit dieser einen Methode ist ein Error. `nil` als Rückgabe heißt: kein Fehler.

```go
func calcRemainderAndMod(num, denom int) (int, int, error) {
    if denom == 0 {
        return 0, 0, errors.New("denominator is 0")
    }
    return num / denom, num % denom, nil
}
```

Konvention für Error-Strings: kleingeschrieben, kein Punkt am Ende, kein Newline. Das wirkt erst stilistisch, ergibt aber Sinn beim Wrapping. Mehrere Fragmente bleiben lesbar.

## errors.New und fmt.Errorf

```go
return 0, errors.New("only even numbers are processed")

return 0, fmt.Errorf("%d isn't an even number", i)
```

`fmt.Errorf` für Laufzeit-Werte im Error-String, `errors.New` für statische Texte.

## Sentinel Errors

Ein Sentinel Error ist ein vordefinierter Error-Wert auf Package-Level. Caller können mit `==` darauf prüfen.

```go
if err == io.EOF {
    // Ende der Daten
}
```

Konvention für den Namen: `ErrXxx`, `io.EOF` ist die bekannte Ausnahme. Sentinel Errors sind Teil der öffentlichen API und müssen mitgepflegt werden. Deshalb sparsam einsetzen, nur wenn ein spezifischer Zustand kommuniziert wird, ohne zusätzliche Daten.

## Eigene Error-Typen

```go
type Status int

const (
    InvalidLogin Status = iota + 1
    NotFound
)

type StatusErr struct {
    Status  Status
    Message string
}

func (se StatusErr) Error() string {
    return se.Message
}
```

Verwendung:

```go
return nil, StatusErr{
    Status:  InvalidLogin,
    Message: fmt.Sprintf("invalid credentials for %s", uid),
}
```

## Die nil-Interface-Falle

```go
func GenerateErrorBroken(flag bool) error {
    var genErr StatusErr // konkreter Typ, nicht error
    if flag {
        genErr = StatusErr{Status: NotFound}
    }
    return genErr
}
```

Wenn `flag` false ist, ist `genErr` der Zero-Wert von `StatusErr`. Der Rückgabe-Typ ist `error`, also wird der konkrete Typ in ein Interface gepackt. Das Interface ist nicht nil. Caller-Code mit `if err != nil` findet einen Fehler, der keiner sein sollte.

Fix: entweder explizit `return nil` oder die lokale Variable als `error` deklarieren.

```go
func GenerateErrorOK(flag bool) error {
    if flag {
        return StatusErr{Status: NotFound}
    }
    return nil
}
```

## Error Wrapping mit %w

```go
func fileChecker(name string) error {
    f, err := os.Open(name)
    if err != nil {
        return fmt.Errorf("in fileChecker: %w", err)
    }
    f.Close()
    return nil
}
```

`%w` wickelt den darunterliegenden Error ein. `errors.Unwrap` peelt ihn ab, wird aber selten direkt benutzt.

Für reines Text-Einbetten ohne Identität: `%v` statt `%w`.

## errors.Is

```go
if errors.Is(err, os.ErrNotExist) {
    fmt.Println("Datei existiert nicht")
}
```

`errors.Is` läuft den Error-Tree durch und prüft jeden Layer gegen den Sentinel-Wert.

## errors.As

```go
var statusErr StatusErr
if errors.As(err, &statusErr) {
    fmt.Println(statusErr.Status, statusErr.Message)
}
```

Pointer auf eine Variable des gesuchten Typs. Findet der Tree einen passenden Error, wird die Variable befüllt.

Regel: `errors.Is` für spezifische Werte, `errors.As` für spezifische Typen.

## Mehrere Errors zusammenfassen

```go
func ValidatePerson(p Person) error {
    var errs []error
    if len(p.FirstName) == 0 {
        errs = append(errs, errors.New("FirstName ist Pflicht"))
    }
    if len(p.LastName) == 0 {
        errs = append(errs, errors.New("LastName ist Pflicht"))
    }
    if p.Age < 0 {
        errs = append(errs, errors.New("Age darf nicht negativ sein"))
    }
    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}
```

`errors.Join` ist sinnvoll, wenn mehrere Fehler in einem Schritt entstehen und alle gemeldet werden sollen.

## defer fürs Wrapping

Wenn jeder Error in einer Funktion mit der gleichen Message gewrappt wird, lässt sich das mit `defer` und einem benannten Return zusammenfassen.

```go
func DoSomeThings(val1 int, val2 string) (_ string, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("in DoSomeThings: %w", err)
        }
    }()
    val3, err := doThing1(val1)
    if err != nil {
        return "", err
    }
    val4, err := doThing2(val2)
    if err != nil {
        return "", err
    }
    return doThing3(val3, val4)
}
```

## panic und recover

`panic` ist nicht der Weg für normale Fehler. Es signalisiert einen Zustand, in dem die Laufzeit nicht weiterarbeiten kann: Index out of range, Division durch Null, nil-Dereferenz.

`recover` fängt einen Panic ab und muss in einem `defer` stehen.

```go
func div60(i int) {
    defer func() {
        if v := recover(); v != nil {
            fmt.Println(v)
        }
    }()
    fmt.Println(60 / i)
}
```

`recover` wird typischerweise nur am Rand einer Library-API verwendet, um Panics aus dem Public Surface zu halten. Im Anwendungscode bleibt die explizite Error-Rückgabe der Standard.

## Stack Traces

Standard-Errors haben keinen Stack Trace. Third-Party-Libraries wie das von Cockroach wrappen Errors mit Trace. Sonst wird die "Spur" durch Wrapping per Hand aufgebaut.

## Zusammenfassung

- `error` ist ein Interface mit `Error() string`
- Error-Strings kleingeschrieben, kein Punkt, kein Newline
- Sentinel Errors sparsam einsetzen
- Eigene Error-Typen mit Struct plus `Error() string`
- Lokale Variablen vom konkreten Error-Typ vermeiden, immer `error`
- `%w` zum Wrappen, `%v` für Text-Einbettung
- `errors.Is` für Werte, `errors.As` für Typen
- `errors.Join` für mehrere Errors
- `defer` plus benannter Return für einheitliches Wrapping
- `panic` und `recover` nicht als Exception-Ersatz
- Stack Traces über Third-Party-Libraries
