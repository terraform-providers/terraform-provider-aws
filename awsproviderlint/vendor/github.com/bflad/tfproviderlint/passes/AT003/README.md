# AT003

The AT003 analyzer reports where an underscore is not
present in the function name, which could make per-resource testing harder to
execute in larger providers or those with overlapping resource names.

## Flagged Code

This example is presuming there are two separate resources:

- `example_thing`
- `example_thing_association`

```go
func TestAccExampleThing(t *testing.T) { /* ... */ }

func TestAccExampleThingSomeAttribute(t *testing.T) { /* ... */ }

func TestAccExampleThingAssociation(t *testing.T) { /* ... */ }

func TestAccExampleThingAssociationSomeAttribute(t *testing.T) { /* ... */ }
```

## Passing Code

```go
func TestAccExampleThing_basic(t *testing.T) { /* ... */ }

func TestAccExampleThing_SomeAttribute(t *testing.T) { /* ... */ }

func TestAccExampleThingAssociation_basic(t *testing.T) { /* ... */ }

func TestAccExampleThingAssociation_SomeAttribute(t *testing.T) { /* ... */ }
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AT003` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:AT003
func TestAccExampleThing(t *testing.T) { /* ... */ }
```
