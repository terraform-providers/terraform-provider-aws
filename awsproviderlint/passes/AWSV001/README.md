# AWSV001

The `AWSV001` analyzer reports when a `validation.StringInSlice()` call has the first parameter of a `[]string`, which suggests either that AWS API model constants are not available or that the usage is prior to the AWS Go SDK adding functions that return all values for the enumeration type.

If the API model constants are not available, this check can be ignored but it is recommended to submit an AWS Support case to the AWS service team for adding the constants.

If the elements of the string slice are AWS Go SDK constants, this check reports when the parameter should be switched to the newer AWS Go SDK `ENUM_Values()` function.

## Flagged Code

```go
&schema.Schema{
    ValidateFunc: validation.StringInSlice([]string{
        service.EnumTypeExample1,
        service.EnumTypeExample2,
    }, false),
}
```

## Passing Code

```go
&schema.Schema{
    ValidateFunc: validation.StringInSlice(service.EnumType_Values(), false),
}
```

## Ignoring Check

The check can be ignored for a certain line via a `//lintignore:AWSV001` comment on the previous line or at the end of the offending line, e.g.

```go
//lintignore:AWSV001
ValidateFunc: validation.StringInSlice([]string{
    service.EnumTypeExample1,
    service.EnumTypeExample2,
}, false),
```
