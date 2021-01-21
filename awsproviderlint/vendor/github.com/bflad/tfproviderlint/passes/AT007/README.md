# AT007

The AT007 analyzer reports acceptance test functions that contain multiple
`resource.ParallelTest()` invocations. Acceptance tests should be split by
invocation and multiple `resource.ParallelTest()` will cause a panic.

## Flagged Code

```go
func TestAccExampleThing_basic(t *testing.T) {
  resource.ParallelTest(/* ... */)
  resource.ParallelTest(/* ... */)
}
```

## Passing Code

```go
func TestAccExampleThing_first(t *testing.T) {
  resource.ParallelTest(/* ... */)
}

func TestAccExampleThing_second(t *testing.T) {
  resource.ParallelTest(/* ... */)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT007` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT007
func TestAccExampleThing_basic(t *testing.T) {
  resource.ParallelTest(/* ... */)
  resource.ParallelTest(/* ... */)
}
```
