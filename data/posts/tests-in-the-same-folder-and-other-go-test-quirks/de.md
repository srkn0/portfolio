---
title: "Tests in Go: testing-Package, Tabletests, Coverage, Fuzzing"
description: "Notizen zu Kapitel 15 von Learning Go: testing-Package, Test-Funktionen, Setup und Teardown, testdata, go-cmp, Tabletests mit Subtests, parallele Tests, Code Coverage und Fuzzing."
tags: [go, learning, tests]
date: 2026-05-11
---

## Kapitel 15

Themen: testing-Package, Test-Konventionen, Setup, Tabletests, Coverage, Fuzzing.

## Wo Tests liegen

Tests liegen im selben Verzeichnis wie der Produktiv-Code, im selben Package. Dadurch können Tests auf unexportierte Funktionen zugreifen. Dateiname endet auf `_test.go`. Zu `adder.go` gehört `adder_test.go`.

## Die kleinste Test-Funktion

```go
func Test_addNumbers(t *testing.T) {
    result := addNumbers(2, 3)
    if result != 5 {
        t.Errorf("incorrect result: expected 5, got %d", result)
    }
}
```

Konventionen:

- Funktion startet mit `Test`
- Ein Parameter `t *testing.T`
- Keine Rückgabewerte
- Kein Assert-DSL, normales Go

`t.Error` und `t.Errorf` markieren den Test als fehlgeschlagen, der Test läuft trotzdem weiter. `t.Fatal` und `t.Fatalf` brechen die Test-Funktion ab.

Gestartet mit `go test ./...`, `-v` für mehr Output.

## Setup und Teardown

Für package-weites Setup gibt es `TestMain`.

```go
func TestMain(m *testing.M) {
    // Setup
    code := m.Run()
    // Teardown
    os.Exit(code)
}
```

Maximal eine `TestMain`-Funktion pro Package.

Test-spezifisches Cleanup mit `t.Cleanup`.

```go
t.Cleanup(func() {
    os.Remove(filename)
})
```

Vergleichbar mit `defer`, lässt sich aber aus Hilfsfunktionen registrieren, weil `t` als Parameter übergeben wird.

Für temporäre Verzeichnisse `t.TempDir`, für Umgebungsvariablen `t.Setenv`.

```go
func TestEnvVar(t *testing.T) {
    t.Setenv("OUTPUT_FORMAT", "JSON")
    cfg := ProcessEnvVars()
    if cfg.OutputFormat != "JSON" {
        t.Error("OutputFormat falsch gesetzt")
    }
}
```

## Sample-Daten in testdata

Fixtures kommen in einen Unterordner `testdata`. Der Compiler behandelt das als Test-Daten und ignoriert es beim Build. Im Test mit relativem Pfad lesen.

## Black-Box-Tests

Für rein öffentliche API-Tests den Package-Namen mit `_test`-Suffix verwenden.

```go
package mypkg_test
```

Trotzdem im selben Verzeichnis, aber als externes Package. Das eigene Package wird importiert.

Beide Formen lassen sich mischen, manche Test-Dateien im Package, manche im `_test`-Package.

## go-cmp für Vergleiche

```go
import "github.com/google/go-cmp/cmp"

if diff := cmp.Diff(expected, result); diff != "" {
    t.Error(diff)
}
```

`cmp.Diff` liefert bei Mismatches einen lesbaren Diff. Für Felder, die ignoriert werden sollen, lässt sich ein Comparer mitgeben.

```go
comparer := cmp.Comparer(func(x, y Person) bool {
    return x.Name == y.Name && x.Age == y.Age
})
cmp.Diff(expected, result, comparer)
```

## Table-Tests mit Subtests

```go
func TestDoMath(t *testing.T) {
    data := []struct {
        name     string
        num1, num2 int
        op       string
        expected int
        errMsg   string
    }{
        {"add", 2, 2, "+", 4, ""},
        {"sub", 2, 2, "-", 0, ""},
        {"mul", 2, 3, "*", 6, ""},
        {"div", 6, 2, "/", 3, ""},
        {"div_zero", 2, 0, "/", 0, "division by zero"},
    }
    for _, d := range data {
        t.Run(d.name, func(t *testing.T) {
            result, err := DoMath(d.num1, d.num2, d.op)
            if result != d.expected {
                t.Errorf("expected %d, got %d", d.expected, result)
            }
            var errMsg string
            if err != nil {
                errMsg = err.Error()
            }
            if errMsg != d.errMsg {
                t.Errorf("expected error %q, got %q", d.errMsg, errMsg)
            }
        })
    }
}
```

`t.Run` startet einen Subtest, der Name landet im Output. Einzeln laufen lassen mit `go test -run TestDoMath/div_zero`.

Hinweis: Error-Strings vergleichen ist fragil. Bei einer Sentinel-Variable oder einem konkreten Typ besser `errors.Is` oder `errors.As`.

## Parallele Tests

Tests laufen sequenziell. Für parallele Ausführung `t.Parallel()` als erstes aufrufen.

```go
func TestMyCode(t *testing.T) {
    t.Parallel()
    // ...
}
```

Praktisch bei IO-bound Tests. Vor Go 1.22 musste die Loop-Variable in Tabletests geshadowt werden, sonst sahen alle Subtests den letzten Wert. Seit 1.22 nicht mehr nötig.

## Coverage

```bash
go test -cover -coverprofile=c.out
go tool cover -html=c.out
```

Im Browser zeigt das Tool zeilenweise Abdeckung. 100 Prozent Coverage heißt nicht 100 Prozent bugfrei. Das Buch zeigt ein Beispiel, in dem ein Typo (`+` statt `*`) erst durch eine zusätzliche Tabellen-Zeile auffällt.

## Fuzzing

Seit Go 1.18 eingebaut. Der Fuzzer generiert zufällige Inputs und sucht nach Panics oder verletzten Invarianten.

```go
func FuzzParseData(f *testing.F) {
    seed := [][]byte{
        []byte("3\nhello\ngoodbye\ngreetings\n"),
        []byte("0\n"),
    }
    for _, s := range seed {
        f.Add(s)
    }
    f.Fuzz(func(t *testing.T, in []byte) {
        // Aufruf der Funktion, Invariante prüfen
    })
}
```

Starten mit `go test -fuzz=FuzzParseData`. Findet der Fuzzer einen Crash, speichert er den Input unter `testdata/fuzz/...` und nutzt ihn ab da als regulären Unit Test.

## Zusammenfassung

- Tests liegen im selben Package wie der Code, Dateinamen enden auf `_test.go`
- `t.Error` für sammeln, `t.Fatal` für sofort abbrechen
- `TestMain` einmal pro Package, `t.Cleanup`, `t.TempDir`, `t.Setenv` pro Test
- `testdata` für Fixtures
- `pkg_test` für Black-Box-Tests
- `go-cmp` für lesbare Vergleiche
- Table-Tests mit `t.Run` für Subtests
- `t.Parallel` für parallele Tests, ab Go 1.22 ohne Loop-Variable-Shadowing
- `go test -cover` plus HTML-Tool für Coverage
- `go test -fuzz=...` für eingebautes Fuzzing
