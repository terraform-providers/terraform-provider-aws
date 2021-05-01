# AT012

The AT012 analyzer reports likely incorrect uses of multiple `TestAcc` function name prefixes up to the conventional underscore (`_`) prefix separator within the same file. Typically, Terraform acceptance tests should use the same naming prefix within one test file so testers can easily run all acceptance tests for the file and not miss associated tests.

Optional parameters:

- `ignored-filenames` Comma-separated list of file names to ignore, defaults to none.

## Flagged Code

```go
func TestAccExampleThing1_Test(t *testing.T) { /* ... */ }

func TestAccExampleThing2_Test(t *testing.T) { /* ... */ }
```

## Passing Code

```go
func TestAccExampleThing_Test1(t *testing.T) { /* ... */ }

func TestAccExampleThing_Test2(t *testing.T) { /* ... */ }
```

## Ignoring Reports

In addition to the optional parameters, reports can be ignored by adding a `//lintignore:AT012` Go code comment before any test declaration to ignore, e.g.

```go
//lintignore:AT012
func TestAccExampleThing1_Test(t *testing.T) { /* ... */ }

//lintignore:AT012
func TestAccExampleThing2_Test(t *testing.T) { /* ... */ }
```

If a file mainly uses one prefix, the code ignores can be simplified to only the non-conforming test declarations.
