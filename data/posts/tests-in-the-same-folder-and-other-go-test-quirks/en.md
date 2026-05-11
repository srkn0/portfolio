---
title: "Tests in Go: testing package, table tests, coverage, fuzzing"
description: "Notes on chapter 15 of Learning Go: the testing package, test functions, setup and teardown, testdata, go-cmp, table tests with subtests, parallel tests, code coverage, and fuzzing."
tags: [go, learning, tests]
date: 2026-05-11
---

## Chapter 15

Topics: the testing package, test conventions, setup, table tests, coverage, fuzzing.

## Where tests live

Tests live in the same directory as the production code, in the same package. This allows tests to reach unexported functions. The file name ends in `_test.go`. `adder.go` gets `adder_test.go`.

## The smallest test function

```go
func Test_addNumbers(t *testing.T) {
    result := addNumbers(2, 3)
    if result != 5 {
        t.Errorf("incorrect result: expected 5, got %d", result)
    }
}
```

Conventions:

- Function starts with `Test`
- One parameter `t *testing.T`
- No return values
- No assert DSL, just regular Go

`t.Error` and `t.Errorf` mark the test as failed but keep running. `t.Fatal` and `t.Fatalf` stop the test function.

Run with `go test ./...`, add `-v` for more output.

## Setup and teardown

For package-wide setup, `TestMain`.

```go
func TestMain(m *testing.M) {
    // setup
    code := m.Run()
    // teardown
    os.Exit(code)
}
```

At most one `TestMain` per package.

Per-test cleanup with `t.Cleanup`.

```go
t.Cleanup(func() {
    os.Remove(filename)
})
```

Comparable to `defer`, but can also be registered from helper functions because `t` is passed in.

For temporary directories, `t.TempDir`. For environment variables, `t.Setenv`.

```go
func TestEnvVar(t *testing.T) {
    t.Setenv("OUTPUT_FORMAT", "JSON")
    cfg := ProcessEnvVars()
    if cfg.OutputFormat != "JSON" {
        t.Error("OutputFormat not set correctly")
    }
}
```

## Sample data in testdata

Fixtures go into a subfolder called `testdata`. The compiler treats it as test data and ignores it during builds. Inside tests, read with a relative path.

## Black box tests

For tests against the public API only, use the package name with a `_test` suffix.

```go
package mypkg_test
```

Same directory, but a separate package. The own package is imported.

Both styles can be mixed: some test files in the package, some in `_test`.

## go-cmp for comparisons

```go
import "github.com/google/go-cmp/cmp"

if diff := cmp.Diff(expected, result); diff != "" {
    t.Error(diff)
}
```

`cmp.Diff` returns a readable diff on mismatches. To ignore fields, pass a comparer.

```go
comparer := cmp.Comparer(func(x, y Person) bool {
    return x.Name == y.Name && x.Age == y.Age
})
cmp.Diff(expected, result, comparer)
```

## Table tests with subtests

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

`t.Run` starts a subtest, its name shows up in the output. Running a single subtest: `go test -run TestDoMath/div_zero`.

Note: comparing error strings is fragile. For sentinel variables or concrete types, prefer `errors.Is` or `errors.As`.

## Parallel tests

Tests run sequentially. For parallel execution, call `t.Parallel()` first.

```go
func TestMyCode(t *testing.T) {
    t.Parallel()
    // ...
}
```

Useful for IO-bound tests. Before Go 1.22 the loop variable in table tests had to be shadowed, otherwise all subtests saw the last value. Since 1.22 this is no longer required.

## Coverage

```bash
go test -cover -coverprofile=c.out
go tool cover -html=c.out
```

The tool shows per-line coverage in the browser. 100 percent coverage does not mean bug-free. The book has an example where a typo (`+` instead of `*`) only surfaces after adding another table row.

## Fuzzing

Built in since Go 1.18. The fuzzer generates random inputs and looks for panics or broken invariants.

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
        // call the function, check the invariant
    })
}
```

Run with `go test -fuzz=FuzzParseData`. When the fuzzer finds a crash, it stores the input under `testdata/fuzz/...` and uses it as a regular unit test from then on.

## Summary

- Tests live in the same package as the code, file names end in `_test.go`
- `t.Error` to accumulate, `t.Fatal` to stop the test immediately
- `TestMain` once per package, `t.Cleanup`, `t.TempDir`, `t.Setenv` per test
- `testdata` for fixtures
- `pkg_test` for black box tests
- `go-cmp` for readable comparisons
- Table tests with `t.Run` for subtests
- `t.Parallel` for parallel tests, no loop-variable shadowing since Go 1.22
- `go test -cover` plus the HTML tool for coverage
- `go test -fuzz=...` for built-in fuzzing
