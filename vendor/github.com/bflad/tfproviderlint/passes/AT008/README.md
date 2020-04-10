# AT008

The AT008 analyzer reports where the `*testing.T` parameter of an acceptance test declaration is not named `t`, which is a standard convention.

## Flagged Code

```go
func TestAccExampleThing_basic(invalid *testing.T) { /* ... */}
```

## Passing Code

```go
func TestAccExampleThing_basic(t *testing.T) { /* ... */}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT008` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT008
func TestAccExampleThing_basic(invalid *testing.T) { /* ... */}
```
