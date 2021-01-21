# Error Handling

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

The Terraform AWS Provider codebase bridges the implementation of a [Terraform Plugin](https://www.terraform.io/docs/extend/how-terraform-works.html) and an AWS API client to support AWS operations and data types as Terraform Resources. An important aspect of performing resource and remote actions is properly handling those operations, but those operations are not guaranteed to succeed every time. Some common examples include where network connections are unreliable, necessary permissions are not properly setup, incorrect Terraform configurations, or the remote system responds unexpectedly. All these situations lead to an unexpected workflow action that must be surfaced to the Terraform user interface for operators to troubleshoot. This guide is intended to explain and show various Terraform AWS Provider code implementations that are considered best practice for surfacing these issues properly to operators and code maintainers.

- [General Guidelines and Helpers](#general-guidelines-and-helpers)
    - [Naming and Check Style](#naming-and-check-style)
    - [Wrap Errors](#wrap-errors)
    - [AWS Go SDK Errors](#aws-go-sdk-errors)
        - [AWS Go SDK Error Helpers](#aws-go-sdk-error-helpers)
        - [Use AWS Go SDK Error Code Constants](#use-aws-go-sdk-error-code-constants)
    - [Terraform Plugin SDK Types and Helpers](#terraform-plugin-sdk-types-and-helpers)
- [Resource Lifecycle Guidelines](#resource-lifecycle-guidelines)
    - [Resource Creation](#resource-creation)
        - [d.IsNewResource() Checks](#disnewresource-checks)
        - [Creation Error Message Context](#creation-error-message-context)
    - [Resource Deletion](#resource-deletion)
        - [Resource Already Deleted](#resource-already-deleted)
        - [Deletion Error Message Context](#deletion-error-message-context)
    - [Resource Read](#resource-read)
        - [Singular Data Source Errors](#singular-data-source-errors)
        - [Plural Data Source Errors](#plural-data-source-errors)
        - [Read Error Message Context](#read-error-message-context)
    - [Resource Update](#resource-update)
        - [Update Error Message Context](#update-error-message-context)

## General Guidelines and Helpers

### Naming and Check Style

Following typical Go conventions, error variables in the Terraform AWS Provider codebase should be named `err`, e.g.

```go
result, err := strconv.Itoa("oh no!")
```

The code that then checks these errors should prefer `if` conditionals that usually `return` (or in the case of looping constructs, `break`/`continue`) early, especially in the case of multiple error checks, e.g.

```go
if /* ... something checking err first ... */ {
    // ... return, break, continue, etc. ...
}

if err != nil {
    // ... return, break, continue, etc. ...
}

// all good!
```

This is in preference of some other styles of error checking, such as `switch` conditionals without a condition.

### Wrap Errors

Go implements error wrapping, which means that a deeply nested function call can return a particular error type, while each function up the stack can provide additional error message context without losing the ability to determine the original error. Additional information about this concept can be found on the [Go blog entry titled Working with Errors in Go 1.13](https://blog.golang.org/go1.13-errors).

For most use cases in this codebase, this means if code is receiving an error and needs to return it, it should implement [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf) and the `%w` verb, e.g.

```go
return fmt.Errorf("adding some additional message: %w", err)
```

### AWS Go SDK Errors

The [AWS Go SDK documentation](https://docs.aws.amazon.com/sdk-for-go/) includes a [section on handling errors](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/handling-errors.html), which is recommended reading.

For the purposes of this documentation, the most important concepts with handling these errors are:

- Each response error (which eventually implements `awserr.Error`) has a `string` error code (`Code`) and `string` error message (`Message`). When printed as a string, they format as: `Code: Message`, e.g. `InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup`.
- Error handling is almost exclusively done via those `string` fields and not other response information, such as HTTP Status Codes.
- When the error code is non-specific, the error message should also be checked. Unfortunately, AWS APIs generally do not provide documentation or API modeling with the contents of these messages and often the Terraform AWS Provider code must rely on substring matching.
- Not all errors are returned in the response error from an AWS Go SDK operation. This is service and sometimes API call specific. For example, the [EC2 `DeleteVpcEndpoints` API call](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DeleteVpcEndpoints.html) can return a "successful" response (in terms of no response error) but include information in an `Unsuccessful` field in the response body.

When working with AWS Go SDK errors, it is preferred to use the helpers outlined below and use the `%w` format verb. Code should generally avoid type assertions with the underlying `awserr.Error` type or calling its `Code()`, `Error()`, `Message()`, or `String()` receiver methods. Using the `%v`, `%#v`, or `%+v` format verbs generally provides extraneous information that is not helpful to operators or code maintainers.

#### AWS Go SDK Error Helpers

To simplify operations with AWS Go SDK error types, the following helpers are available via the `github.com/hashicorp/aws-sdk-go-base/tfawserr` Go package:

- `tfawserr.ErrCodeEquals(err, "Code")`: Preferred when the error code is specific enough for the check condition. For example, a `ResourceNotFoundError` code provides enough information that the requested API resource identifier/Amazon Resource Name does not exist.
- `tfawserr.ErrMessageContains(err, "Code", "MessageContains")`: Does simple substring matching for the error message.

The recommendation for error message checking is to be just specific enough to capture the anticipated issue, but not include _too_ much matching as the AWS API can change over time without notice. The maintainers have observed changes in wording and capitalization cause unexpected issues in the past.

For example, given this error code and message:

```
InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup
```

An error check for this might be:

```go
if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "cannot be assumed") { /* ... */ }
```

The Amazon Resource Name in the error message will be different for every environment and does not add value to the check. The AWS Backup suffix is also extraneous and could change should the service ever rename.

_The codebase also contains an older style `isAWSErr(err, "CodeEquals", "MessageContains")` helper, which has not yet been refactored out. The helpers above are preferred for clarity._

#### Use AWS Go SDK Error Code Constants

Each AWS Go SDK service API typically implements common error codes, which get exported as public constants in the AWS Go SDK. In the [AWS Go SDK API Reference](https://docs.aws.amazon.com/sdk-for-go/api/), these can be found in each of the service packages under the `Constants` section (typically named `ErrCode{ExceptionName}`).

If an AWS Go SDK service API is missing an error code constant, an AWS Support case should be submitted and a new constant can be added to `aws/internal/service/{SERVICE}/errors.go` file (created if not present), e.g.

```go
const(
    ErrCodeInvalidParameterException = "InvalidParameterException"
)
```

Then referencing code can use it via:

```go
// imports
tf{SERVICE} "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/{SERVICE}"

// logic
tfawserr.ErrCodeEquals(err, tf{SERVICE}.ErrCodeInvalidParameterException)
```

e.g.

```go
// imports
tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"

// logic
tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameterException)
```

### Terraform Plugin SDK Types and Helpers

The Terraform Plugin SDK includes some error types which are used in certain operations and typically preferred over implementing new types:

* [`resource.NotFoundError`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#NotFoundError)
* [`resource.TimeoutError`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#TimeoutError): Returned from [`resource.Retry()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#Retry), [`resource.RetryContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#RetryContext), [`(resource.StateChangeConf).WaitForState()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#StateChangeConf.WaitForState), and [`(resource.StateChangeConf).WaitForStateContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#StateChangeConf.WaitForStateContext)

The Terraform AWS Provider codebase implements some additional helpers for working with these in the `github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource` package:

- `tfresource.NotFound(err)`: Returns true if the error is a `resource.NotFoundError`.
- `tfresource.TimedOut(err)`: Returns true if the error is a `resource.TimeoutError` and contains no `LastError`. This typically signifies that the retry logic was never signaled for a retry, which can happen when AWS Go SDK operations are automatically retrying before returning.

## Resource Lifecycle Guidelines

Terraform CLI and the Terraform Plugin SDK have certain expectations and automatic behaviors depending on the lifecycle operation of a resource. This section highlights some common issues that can occur and their expected resolution.

### Resource Creation

Invoked in the resource via the `schema.Resource` type `Create`/`CreateContext` function.

#### d.IsNewResource() Checks

During resource creation, Terraform CLI expects either a properly applied state for the new resource or an error. To signal proper resource existence, the Terraform Plugin SDK uses an underlying resource identifier (set via `d.SetId(/* some value */)`). If for some reason the resource creation is returned without an error, but also without the resource identifier being set, Terraform CLI will return an error such as:

```
Error: Provider produced inconsistent result after apply

When applying changes to aws_sns_topic_subscription.sqs,
provider "registry.terraform.io/hashicorp/aws" produced an unexpected new
value: Root resource was present, but now absent.

This is a bug in the provider, which should be reported in the provider's own
issue tracker.
```

A typical pattern in resource implementations in the `Create`/`CreateContext` function is to `return` the `Read`/`ReadContext` function at the end to fill in the Terraform State for all attributes. Another typical pattern in resource implementations in the `Read`/`ReadContext` function is to remove the resource from the Terraform State if the remote system returns an error or status that indicates the remote resource no longer exists by explicitly calling `d.SetId("")` and returning no error. If the remote system is not strongly read-after-write consistent (eventually consistent), this means the resource creation can return no error and also return no resource state.

To prevent this type of Terraform CLI error, the resource implementation should also check against `d.IsNewResource()` before removing from the Terraform State and returning no error. If that check is `true`, then remote operation error (or one synthesized from the non-existent status) should be returned instead. While adding this check will not fix the resource implementation to handle the eventually consistent nature of the remote system, the error being returned will be less opaque for operators and code maintainers to troubleshoot.

In the Terraform AWS Provider, an initial fix for the Terraform CLI error will typically look like:

```go
func resourceServiceThingCreate(d *schema.ResourceData, meta interface{}) error {
    /* ... */

    return resourceServiceThingRead(d, meta)
}

func resourceServiceThingRead(d *schema.ResourceData, meta interface{}) error {
    /* ... */

    output, err := conn.DescribeServiceThing(input)

    if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
        log.Printf("[WARN] {Service} {Thing} (%s) not found, removing from state", d.Id())
        d.SetId("")
        return nil
    }

    if err != nil {
        return fmt.Errorf("error reading {Service} {Thing} (%s): %w", d.Id(), err)
    }

    /* ... */
}
```

Future documentation will show how to properly retry the remote operation for a short period of time until it is successful to remove the error completely.

#### Creation Error Message Context

Returning errors during creation should include additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf):

```go
if err != nil {
    return fmt.Errorf("error creating {SERVICE} {THING}: %w", err)
}
```

e.g.

```go
if err != nil {
    return fmt.Errorf("error creating EC2 VPC: %w", err)
}
```

Code that also uses waiters or other operations that return errors should follow a similar pattern, including the resource identifier since it has typically been set before this execution:

```go
if _, err := waiter.VpcAvailable(conn, d.Id()); err != nil {
    return fmt.Errorf("error waiting for EC2 VPC (%s) availability: %w", d.Id(), err)
}
```

### Resource Deletion

Invoked in the resource via the `schema.Resource` type `Delete`/`DeleteContext` function.

#### Resource Already Deleted

A typical pattern for resource deletion is to immediately perform the remote system deletion operation without checking existence. This is generally acceptable as operators are encouraged to always refresh their Terraform State prior to performing changes. However in certain scenarios, such as external systems modifying the remote system prior to the Terraform execution, it is certainly still possible that the remote system will return an error signifying that remote resource does not exist. In these cases, resources should implement logic that catches the error and returns no error.

_NOTE: The Terraform Plugin SDK automatically handles the equivalent of d.SetId("") on deletion, so it is not necessary to include it._

For example in the Terraform AWS Provider:

```go
func resourceServiceThingDelete(d *schema.ResourceData, meta interface{}) error {
    /* ... */

    output, err := conn.DeleteServiceThing(input)

    if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
        return nil
    }

    if err != nil {
        return fmt.Errorf("error deleting {Service} {Thing} (%s): %w", d.Id(), err)
    }

    /* ... */
}
```

#### Deletion Error Message Context

Returning errors during deletion should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf):

```go
if err != nil {
    return fmt.Errorf("error deleting {SERVICE} {THING} (%s): %w", d.Id(), err)
}
```

e.g.

```go
if err != nil {
    return fmt.Errorf("error deleting EC2 VPC (%s): %w", d.Id(), err)
}
```

Code that also uses waiters or other operations that return errors should follow a similar pattern:

```go
if _, err := waiter.VpcDeleted(conn, d.Id()); err != nil {
    return fmt.Errorf("error waiting for EC2 VPC (%s) deletion: %w", d.Id(), err)
}
```

### Resource Read

Invoked in the resource via the `schema.Resource` type `Read`/`ReadContext` function.

#### Singular Data Source Errors

A data source which is expected to return Terraform State about a single remote resource is commonly referred to as a "singular" data source. Implementation-wise, it may use any available describe or listing functionality from the remote system to retrieve the information. In addition to any remote operation and other data handling errors that should be returned, these two additional cases should be covered:

- Returning an error when zero results are found.
- Returning an error when multiple results are found.

For remote operations that are designed to return an error when the remote resource is not found, this error is typically just passed through similar to other remote operation errors. For remote operations that are designed to return a successful result whether there is zero, one, or multiple multiple results the error must be generated.

For example in pseudo-code:

```go
output, err := conn.ListServiceThings(input)

if err != nil {
    return fmt.Errorf("error listing {Service} {Thing}s: %w", err)
}

if output == nil || len(output.Results) == 0 {
    return fmt.Errorf("no {Service} {Thing} found matching criteria; try different search")
}

if len(output.Results) > 1 {
    return fmt.Errorf("multiple {Service} {Thing} found matching criteria; try different search")
}
```

#### Plural Data Source Errors

An emergent concept is a data source that returns multiple results, acting similar to any available listing functionality available from the remote system. These types of data sources should return _no_ error if zero results are returned and _no_ error if multiple results are found. Remote operation and other data handling errors should still be returned.

#### Read Error Message Context

Returning errors during read should include the resource identifier (for managed resources) and additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf):

```go
if err != nil {
    return fmt.Errorf("error reading {SERVICE} {THING} (%s): %w", d.Id(), err)
}
```

e.g.

```go
if err != nil {
    return fmt.Errorf("error reading EC2 VPC (%s): %w", d.Id(), err)
}
```

### Resource Update

Invoked in the resource via the `schema.Resource` type `Update`/`UpdateContext` function.

#### Update Error Message Context

Returning errors during update should include the resource identifier and additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://pkg.go.dev/fmt#Errorf):

```go
if err != nil {
    return fmt.Errorf("error updating {SERVICE} {THING} (%s): %w", d.Id(), err)
}
```

e.g.

```go
if err != nil {
    return fmt.Errorf("error updating EC2 VPC (%s): %w", d.Id(), err)
}
```

Code that also uses waiters or other operations that return errors should follow a similar pattern:

```go
if _, err := waiter.VpcAvailable(conn, d.Id()); err != nil {
    return fmt.Errorf("error waiting for EC2 VPC (%s) update: %w", d.Id(), err)
}
```
