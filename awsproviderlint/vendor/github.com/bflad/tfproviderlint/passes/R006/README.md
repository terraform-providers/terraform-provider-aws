# R006

The R006 analyzer reports when `RetryFunc` declarations are missing retryable errors (e.g. `RetryableError()` calls) and should not be used as `RetryFunc`.

Optional parameters:

- `-package-aliases` Comma-separated list of additional Go import paths to consider as aliases for helper/resource, defaults to none.

## Flagged Code

```go
err := resource.Retry(1 * time.Minute, func() *RetryError {
  // Calling API logic, e.g.
  _, err := conn.DoSomething(input)

  if err != nil {
    return resource.NonRetryableError(err)
  }

  return nil
})
```

## Passing Code

```go
_, err := conn.DoSomething(input)

if err != nil {
  return err
}

// or

err := resource.Retry(1 * time.Minute, func() *RetryError {
  // Calling API logic, e.g.
  _, err := conn.DoSomething(input)

  if /* check err for retryable condition */ {
    return resource.RetryableError(err)
  }

  if err != nil {
    return resource.NonRetryableError(err)
  }

  return nil
})
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:R006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
//lintignore:R006
err := resource.Retry(1 * time.Minute, func() *RetryError {
  // Calling API logic, e.g.
  _, err := conn.DoSomething(input)

  if err != nil {
    return resource.NonRetryableError(err)
  }

  return nil
})
```
