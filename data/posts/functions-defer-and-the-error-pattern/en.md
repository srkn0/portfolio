---
title: "Functions, defer, and the error pattern"
description: "Notes on chapters 4 and 5 of Learning Go: blocks, shadowing, for, switch, multiple return values, the if err != nil pattern, defer, closures, call by value."
tags: [go, learning, functions, error-handling]
date: 2026-03-23
---

## Chapters 4 and 5

Topics: blocks and shadowing, control flow, functions, closures, defer, call by value.

## Blocks and shadowing

Each pair of braces creates a block. Variables declared inside a block exist only within that block.

A variable with the same name in an inner block shadows the outer one.

```go
x := 10
if x > 5 {
    x := 20  // new x
    fmt.Println(x) // 20
}
fmt.Println(x) // 10
```

Both `x` are different variables. Shadowing can introduce subtle bugs, and `go vet` does not catch every case.

## for loop

Go has a single loop construct: `for`. Four forms.

```go
// classic
for i := 0; i < 10; i++ { }

// condition only, equivalent to while
for x < 100 { }

// infinite
for { }

// for-range over slices, maps, strings
for i, v := range items { }
```

`for-range` over strings iterates over runes, not bytes. The value variable is a copy, modifying it does not change the original.

## switch

`switch` does not fall through. After the first matching case, execution leaves the statement.

```go
switch size := len(word); size {
case 1, 2, 3, 4:
    fmt.Println("short word")
case 5:
    fmt.Println("exactly right")
default:
    fmt.Println("long word")
}
```

Multiple values per case with commas. Switch on any comparable type.

The blank switch replaces an if/else chain.

```go
switch {
case len(word) < 5:
    fmt.Println("short")
case len(word) > 10:
    fmt.Println("long")
default:
    fmt.Println("medium")
}
```

## Multiple return values

```go
func divAndRemainder(num, denom int) (int, int, error) {
    if denom == 0 {
        return 0, 0, errors.New("cannot divide by zero")
    }
    return num / denom, num % denom, nil
}
```

Call site:

```go
result, remainder, err := divAndRemainder(5, 2)
if err != nil {
    fmt.Println(err)
    os.Exit(1)
}
fmt.Println(result, remainder)
```

Convention: error is the last return value. `nil` means success.

## The `if err != nil` pattern

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

Every call checks the error. Happy path and error handling sit next to each other.

## Functions as values

Functions are values. They can be assigned to variables, passed as arguments, and returned.

```go
var opMap = map[string]func(int, int) int{
    "+": func(a, b int) int { return a + b },
    "-": func(a, b int) int { return a - b },
}
```

### Closures

Functions declared inside other functions are closures. They can read and modify variables from the outer function.

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

Closures are commonly passed to `sort.Slice`.

```go
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})
```

## defer

`defer` delays a function call until the surrounding function exits. Common use case: closing resources.

```go
f, err := os.Open(filename)
if err != nil {
    log.Fatal(err)
}
defer f.Close()
```

Multiple `defer` calls run in LIFO order.

`defer` combined with named return values is useful for database transaction cleanup:

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

## Call by value

All values are passed as copies. Functions receive copies, not references. Maps and slices behave differently because they internally contain pointers.

```go
func modSlice(s []int) {
    for k, v := range s {
        s[k] = v * 2  // modifies the original
    }
    s = append(s, 10)  // does NOT modify the original
}
```

Elements of a passed slice can be modified. Growing it is unreliable because `append` may allocate a new backing array that the caller never sees.

## Summary

- One loop construct, four forms
- `switch` does not fall through, blank switch replaces if/else chains
- Multiple return values, error last, `nil` for success
- `if err != nil` is visible control flow
- `defer` guarantees cleanup, runs in LIFO order
- Functions are values, can be stored and passed
- Call by value, maps and slices are special cases due to internal pointers
