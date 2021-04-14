# AT009

The AT009 analyzer reports where `acctest.RandStringFromCharSet()` calls can be simplified to `acctest.RandString()`.

## Flagged Code

```go
rString := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
```

## Passing Code

```go
rString := acctest.RandString(8)
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT009` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT009
rString := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
```
