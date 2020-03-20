# R009

The R009 analyzer reports usage of Go panics, which should be avoided. Any errors should be surfaced to Terraform, which will display them in the user interface and ensure any necessary state actions (e.g. cleanup) are performed as expected.

## Flagged Code

```go
panic("oops")

log.Panic("eek")

log.Panicf("yikes")

log.Panicln("run away")
```

## Passing Code

```go
// Not present :)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R009` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R009
panic("oops")
```
