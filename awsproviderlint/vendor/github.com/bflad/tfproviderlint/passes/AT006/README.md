# AT006

The AT006 analyzer reports acceptance test functions that contain multiple
`resource.Test()` invocations. Acceptance tests should be split by invocation.

## Flagged Code

```go
func TestAccExampleThing_basic(t *testing.T) {
  resource.Test(/* ... */)
  resource.Test(/* ... */)
}
```

## Passing Code

```go
func TestAccExampleThing_first(t *testing.T) {
  resource.Test(/* ... */)
}

func TestAccExampleThing_second(t *testing.T) {
  resource.Test(/* ... */)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT006
func TestAccExampleThing_basic(t *testing.T) {
  resource.Test(/* ... */)
  resource.Test(/* ... */)
}
```
