# Retries and Waiters

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

Terraform plugins may run into situations where calling the remote system after an operation may be necessary. These typically fall under three classes where:

- The request never reaches the remote system.
- The request reaches the remote system and responds that it cannot handle the request temporarily.
- The implementation of the remote system requires additional requests to ensure success.

This guide describes the behavior of the Terraform AWS Provider and provides code implementations that help ensure success in each of these situations.

- [Terraform Plugin SDK Functionality](#terraform-plugin-sdk-functionality)
    - [State Change Configuration and Functions](#state-change-configuration-and-functions)
    - [Retry Functions](#retry-functions)
- [AWS Request Handling](#aws-request-handling)
    - [Default AWS Go SDK Retries](#default-aws-go-sdk-retries)
    - [Lower Network Error Retries](#lower-network-error-retries)
    - [Terraform AWS Provider Service Retries](#terraform-aws-provider-service-retries)
- [Eventual Consistency](#eventual-consistency)
    - [Operation Specific Error Retries](#operation-specific-error-retries)
        - [IAM Error Retries](#iam-error-retries)
    - [Resource Lifecycle Retries](#resource-lifecycle-retries)
    - [Resource Attribute Value Waiters](#resource-attribute-value-waiters)
- [Asynchronous Operations](#asynchronous-operations)
    - [AWS Go SDK Waiters](#aws-go-sdk-waiters)
    - [Resource Lifecycle Waiters](#resource-lifecycle-waiters)

## Terraform Plugin SDK Functionality

The [Terraform Plugin SDK](https://github.com/hashicorp/terraform-plugin-sdk/), which the AWS Provider uses, provides vital tools for handling consistency: the `resource.StateChangeConf{}` struct, and the retry functions, `resource.Retry()` and `resource.RetryContext()`. We will discuss these throughout the rest of this guide. Since they help keep the AWS Provider code consistent, we heavily prefer them over custom implementations.

This guide goes beyond the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/retries-and-customizable-timeouts.html) by providing additional context and emergent implementations specific to the Terraform AWS Provider.

### State Change Configuration and Functions

The [`resource.StateChangeConf` type](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#StateChangeConf) along with its receiver methods [`WaitForState()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#StateChangeConf.WaitForState) and [`WaitForStateContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#StateChangeConf.WaitForStateContext) is a generic primitive for repeating operations in Terraform resource logic until desired value(s) are received. The "state change" in this case is generic to any value and not specific to the Terraform State. Among other functionality, it supports some of these desirable optional properties:

- Expecting specific value(s) while waiting for the target value(s) to be reached. Unexpected values are returned as an error which can be augmented with additional details.
- Expecting the target value(s) to be returned multiple times in succession.
- Allowing various polling configurations such as delaying the initial request and setting the time between polls.

### Retry Functions

The [`resource.Retry()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#Retry) and [`resource.RetryContext()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource#RetryContext) functions provide a simplified retry implementation around `resource.StateChangeConf`. Their most common use is for simple error-based retries.

## AWS Request Handling

The Terraform AWS Provider's requests to AWS service APIs happen on top of Hypertext Transfer Protocol (HTTP). The following is a simplified description of the layers and handling that requests pass through:

- A Terraform resource calls an AWS Go SDK function.
- The AWS Go SDK generates an AWS-compatible HTTP request using the [Go standard library `net/http` package](https://pkg.go.dev/net/http/). This includes the following:
    - Adding HTTP headers for authentication and signing of requests to ensure authenticity.
    - Converting operation inputs into required HTTP URI parameters and/or request body type (XML or JSON).
    - If debug logging is enabled, logging of the HTTP request.
- The AWS Go SDK transmits the `net/http` request using Go's standard handling of the Operating System (OS) and Domain Name System (DNS) configuration.
- The AWS service potentially receives the request and responds, typically adding a request identifier HTTP header which can be used for AWS Support cases.
- The OS and Go `net/http` receive the response and pass it to the AWS Go SDK.
- The AWS Go SDK attempts to handle the response. This may include:
    - Parsing output
    - Converting errors into operation errors (Go `error` type of wrapped [`awserr.Error` type](https://docs.aws.amazon.com/sdk-for-go/api/aws/awserr/#Error)).
    - Converting response elements into operation outputs (AWS Go SDK operation-specific types).
    - Triggering automatic request retries based on default and custom logic.
- The Terraform resource receives the response, including any output and errors, from the AWS Go SDK.

The Terraform AWS Provider specific configuration for AWS Go SDK operation handling can be found in `aws/config.go` in this codebase and the [`hashicorp/aws-sdk-go-base` codebase](https://github.com/hashicorp/aws-sdk-go-base).

_NOTE: The section descibes the current handling with version 1 of the AWS Go SDK. In the future, this codebase will be migrated to version 2 of the AWS Go SDK. The newer version implements a very similar request flow but uses a simpler credential and request handling configuration. As such, the `aws-sdk-go-base` dependency will likely not receive further updates and will be removed after that migration._

### Default AWS Go SDK Retries

In some situations, while handling a response, the AWS Go SDK automatically retries a request before returning the output and error. The retry mechanism implements an exponential backoff algorithm. The default conditions triggering automatic retries (implemented through [`client.DefaultRetryer`](https://docs.aws.amazon.com/sdk-for-go/api/aws/client/#DefaultRetryer)) include:

- Certain network errors. A common exception to this is connection reset errors.
- HTTP status codes 429 and 5xx.
- Certain API error codes, which are common across various AWS services (e.g. `ThrottledException`). However, not all AWS services implement these error codes consistently. A common exception to this is certain expired credentials errors.

By default, the Terraform AWS Provider sets the maximum number of AWS Go SDK retries based on the [`max_retries` provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#max_retries). The provider configuration defaults to 25 and the exponential backoff roughly equates to one hour of retries. This very high default value was present before the Terraform AWS Provider codebase was split from Terraform CLI in version 0.10.

_NOTE: The section describes the current handling with version 1 of the AWS Go SDK. In the future, this codebase will be migrated to version 2 of the AWS Go SDK. The newer version implements additional retry conditions by default, such as consistently retrying all common network errors._

_NOTE: The section describes the current handling with Terraform Plugin SDK resource signatures without `context.Context`. In the future, this codebase will be migrated to the context-aware resource signatures which currently enforce a 20-minute default timeout that conflicts with the timeout with the default `max_retries` value. The Terraform Plugin SDK may be updated to support removing this default 20-minute timeout or the default retry mechanism described here will be updated to prevent context cancellation errors where possible._

### Lower Network Error Retries

Given the very high default number of AWS Go SDK retries configured in the Terraform AWS Provider and the excessive wait that practitioners would face, the [`hashicorp/aws-sdk-go-base` codebase](https://github.com/hashicorp/aws-sdk-go-base/blob/57529b4c2d2f8f3b5299d66a829b01259fa800d7/session.go#L108-L130) lowers retries to 10 for certain network errors that typically cannot be remediated via retries. This roughly equates to 30 seconds of retries.

### Terraform AWS Provider Service Retries

The AWS Go SDK provides hooks for injecting custom logic into the service client handlers. We prefer this handling in situations where contributors would need to apply the retry behavior to many resources. For example, in cases where the AWS service API does not mark an error code as automatically retriable. The AWS Provider includes other retry-changing behaviors using this method. You can find them in the `aws/config.go` file. For example:

```go
client.kafkaconn.Handlers.Retry.PushBack(func(r *request.Request) {
	if tfawserr.ErrMessageContains(r.Error, kafka.ErrCodeTooManyRequestsException, "Too Many Requests") {
		r.Retryable = aws.Bool(true)
	}
})
```

## Eventual Consistency

Eventual consistency is a temporary condition where the remote system can return outdated information or errors due to not being strongly read-after-write consistent. This is a pattern found in remote systems that must be highly scaled for broad usage.

Terraform expects any planned resource lifecycle change (create, update, destroy of the resource itself) and planned resource attribute value change to match after being applied. Conversely, operators typically expect that Terraform resources also implement the concept of drift detection for resources and their attributes, which requires reading information back from the remote system after an operation. A common implementation is refreshing the Terraform State information (`d.Set()`) during the `Read` function of a resource after `Create` and `Update`.

These two concepts conflict with each other and require additional handling in Terraform resource logic as shown in the following sections. These issues are _not_ reliably reproducible, especially in the case of writing acceptance testing, so they can be elusive with false positives to verify fixes.

### Operation Specific Error Retries

Even given a properly ordered Terraform configuration, eventual consistency can unexpectedly prevent downstream operations from succeeding. A simple retry after a few seconds resolves many of these issues. To reduce frustrating behavior for operators, wrap AWS Go SDK operations with the `resource.Retry()` or `resource.RetryContext()` functions. These retries should have a reasonably low timeout (typically two minutes but up to five minutes). Save them in a constant for reusability. These functions are preferably in line with the associated resource logic to remove any indirection with the code.

Do not use this type of logic to overcome improperly ordered Terraform configurations. The approach may not work in larger environments.

```go
// aws/internal/service/example/waiter/waiter.go (created if does not exist)

const (
	// Maximum amount of time to wait for Thing operation eventual consistency
	ThingOperationTimeout = 2 * time.Minute
)
```

```go
// aws/resource_example_thing.go

import (
	// ... other imports ...
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/example/waiter"
)

// ... Create, Read, Update, or Delete function ...
	err := resource.Retry(waiter.ThingOperationTimeout, func() *resource.RetryError {
		_, err := conn./* ... AWS Go SDK operation with eventual consistency errors ... */

		// Retryable conditions which can be checked.
		// These must be updated to match the AWS service API error code and message.
		if tfawserr.ErrMessageContains(err, /* error code */, /* error message */) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// This check is important - it handles when the AWS Go SDK operation retries without returning.
	// e.g. any automatic retries due to network or throttling errors.
	if tfresource.TimedOut(err) {
		// The use of equals assignment (over colon equals) is also important here.
		// This overwrites the error variable to simplify logic.
		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
	}

	if err != nil {
		return fmt.Errorf("... error message context ... : %w", err)
	}
```

_NOTE: The section descibes the current handling with version 1 of the AWS Go SDK. In the future, this codebase will be migrated to version 2 of the AWS Go SDK. The newer version natively supports operation-specific retries in a more friendly manner, which may replace this type of implementation._

#### IAM Error Retries

A common eventual consistency issue is an error returned due to IAM permissions. The IAM service itself is eventually consistent along with the propagation of its components and permissions to other AWS services. For example, if the following operations occur in quick succession:

- Create an IAM Role
- Attach an IAM Policy to the IAM Role
- Reference the new IAM Role in another AWS service, such as creating a Lambda Function

The last operation can receive varied API errors ranging from:

- IAM Role being reported as not existing
- IAM Role being reported as not having permissions for the other service to use it (assume role permissions)
- IAM Role being reported as not having sufficient permissions (inline or attached role permissions)

Each AWS service API (and sometimes even operations within the same API) varies in the implementation of these errors. To handle them, it is recommended to use the [Operation Specific Error Retries](#operation-specific-error-retries) pattern. The Terraform AWS Provider implements a standard timeout constant of two minutes in the `aws/internal/service/iam/waiter` package which should be used for all retry timeouts associated with IAM errors. This timeout was derived from years of Terraform operational experience with all AWS APIs.

```go
// aws/resource_example_thing.go

import (
	// ... other imports ...
	// By convention, cross-service waiter imports are aliased as {SERVICE}waiter
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

// ... Create and typically Update function ...
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */

		// Example retryable condition
		// This must be updated to match the AWS service API error code and message.
		if tfawserr.ErrMessageContains(err, /* error code */, /* error message */) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn./* ... AWS Go SDK operation with IAM eventual consistency errors ... */
	}

	if err != nil {
		return fmt.Errorf("... error message context ... : %w", err)
	}
```

### Resource Lifecycle Retries

Resource lifecycle eventual consistency is a type of consistency issue that relates to the existence or state of an AWS infrastructure component. For example, if you create a resource and immediately try to get information about it, some AWS services and operations will return a "not found" error. Depending on the service and general AWS load, these errors can be frequent or rare.

In order to avoid this issue, identify operations that make changes. Then, when calling any other operations that rely on the changes, account for the possibility that the AWS service has not yet fully realized them.

A typical example is creating an AWS component. After creation, when attempting to read the component's information, provide logic to retry the read if the AWS service returns a "not found" error.

The pattern that most resources should follow is to have the `Create` function return calling the `Read` function. This fills in computed attributes and ensures that the AWS service applied the configuration correctly. Add retry logic to the `Read` function to overcome the temporary condition on resource creation.

Note that for eventually consistent resources, "not found" errors can still occur in the `Read` function even after implementing [Resource Lifecycle Waiters](#resource-lifecycle-waiters) for the Create function.

```go
// aws/internal/service/example/waiter/waiter.go (created if does not exist)

const (
	// Maximum amount of time to wait for Thing eventual consistency on creation
	ThingCreationTimeout = 2 * time.Minute
)
```

```go
// aws/resource_example_thing.go

function ExampleThingCreate(d *schema.ResourceData, meta interface{}) error {
	// ...
	return ExampleThingRead(d, meta)
}

function ExampleThingRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).exampleconn

	input := &example.OperationInput{/* ... */}

	var output *example.OperationOutput
	err := resource.Retry(waiter.ThingCreationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.Operation(input)

		// Retry on any API "not found" errors, but only on new resources.
		if d.IsNewResource() && tfawserr.ErrorCodeEquals(err, example.ErrCodeResourceNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Retry AWS Go SDK operation if no response from automatic retries.
	if tfresource.TimedOut(err) {
		output, err = exampleconn.Operation(input)
	}

	// Prevent confusing Terraform error messaging to operators by
	// Only ignoring API "not found" errors if not a new resource.
	if !d.IsNewResource() && tfawserr.ErrorCodeEquals(err, example.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] Example Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Example Thing (%s): %w", d.Id(), err)
	}

	// Prevent panics.
	if output == nil {
		return fmt.Errorf("error reading Example Thing (%s): empty response", d.Id())
	}

	// ... refresh Terraform state as normal ...
	d.Set("arn", output.Arn)
}
```

Some other general guidelines are:

- If the `Create` function uses `resource.StateChangeConf`, the underlying `resource.RefreshStateFunc` should `return nil, "", nil` instead of the API "not found" error. This way the `StateChangeConf` logic will automatically retry.
- If the `Create` function uses `resource.Retry()`, the API "not found" error should be caught and `return resource.RetryableError(err)` to automatically retry.

In rare cases, it may be easier to duplicate all `Read` function logic in the `Create` function to handle all retries in one place.

### Resource Attribute Value Waiters

An emergent solution for handling eventual consistency with attribute values on updates is to introduce a custom `resource.StateChangeConf` and `resource.RefreshStateFunc` handlers. For example:

```go
// aws/internal/service/example/waiter/status.go (created if does not exist)

// ThingAttribute fetches the Thing and its Attribute
func ThingAttribute(conn *example.Example, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := /* ... AWS Go SDK operation to fetch resource/value ... */

		if tfawserr.ErrCodeEquals(err, example.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Attribute), nil
	}
}
```

```go
// aws/internal/service/example/waiter/waiter.go (created if does not exist)

const (
	ThingAttributePropagationTimeout = 2 * time.Minute
)

// ThingAttributeUpdated is an attribute waiter for ThingAttribute
func ThingAttributeUpdated(conn *example.Example, id string, expectedValue string) (*example.Thing, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{expectedValue},
		Refresh: ThingAttribute(conn, id),
		Timeout: ThingAttributePropagationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*example.Thing); ok {
		return output, err
	}

	return nil, err
}
```

```go
// aws/resource_example_thing.go

function ExampleThingUpdate(d *schema.ResourceData, meta interface{}) error {
	// ...

	d.HasChange("attribute") {
		// ... AWS Go SDK logic to update attribute ...

		if _, err := waiter.ThingAttributeUpdated(conn, d.Id(), d.Get("attribute").(string)); err != nil {
			return fmt.Errorf("error waiting for Example Thing (%s) attribute update: %w", d.Id(), err)
		}
	}

	// ...
}
```

## Asynchronous Operations

When you initiate a long-running operation, an AWS service may return a successful response immediately and continue working on the request asynchronously. A resource can track the status with a component-level field (e.g. `CREATING`, `UPDATING`, etc.) or an explicit tracking identifier.

Terraform resources should wait for these background operations to complete. Failing to do so can introduce incomplete state information and downstream errors in other resources. In rare scenarios involving very long-running operations, operators may request a flag to skip the waiting. However, these should only be implemented case-by-case to prevent those previously mentioned confusing issues.

### AWS Go SDK Waiters

In limited cases, the AWS service API model includes the information to automatically generate a waiter function in the AWS Go SDK for an operation. These are typically named with the prefix `WaitUntil...`. If available, these functions can be used for an initial resource implementation. For example:

```go
if err := conn.WaitUntilEndpointInService(input); err != nil {
	return fmt.Errorf("error waiting for Example Thing (%s) ...: %w", d.Id(), err)
}
```

If it is necessary to customize the timeouts and polling, we generally prefer using [Resource Lifecycle Waiters](#resource-lifecycle-waiters) instead since they are more commonly used throughout the codebase.

### Resource Lifecycle Waiters

Most of the codebase uses `resource.StateChangeConf` and `resource.RefreshStateFunc` handlers for tracking either component level status fields or explicit tracking identifiers. These should be placed in the `aws/internal/service/{SERVICE}/waiter` package and split into separate functions. For example:

```go
// aws/internal/service/example/waiter/status.go (created if does not exist)

// ThingStatus fetches the Thing and its Status
func ThingStatus(conn *example.Example, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := /* ... AWS Go SDK operation to fetch resource/status ... */

		if tfawserr.ErrCodeEquals(err, example.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
```

```go
// aws/internal/service/example/waiter/waiter.go (created if does not exist)

const (
	ThingCreationTimeout = 2 * time.Minute
	ThingDeletionTimeout = 5 * time.Minute
)

// ThingCreated is a resource waiter for Thing creation
func ThingCreated(conn *example.Example, id string) (*example.Thing, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{example.StatusCreating},
		Target:  []string{example.StatusCreated},
		Refresh: ThingStatus(conn, id),
		Timeout: ThingCreationTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*example.Thing); ok {
		return output, err
	}

	return nil, err
}

// ThingDeleted is a resource waiter for Thing deletion
func ThingDeleted(conn *example.Example, id string) (*example.Thing, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{example.StatusDeleting},
		Target:  []string{}, // Use empty list if the resource disappears and does not have "deleted" status
		Refresh: ThingStatus(conn, id),
		Timeout: ThingDeletionTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*example.Thing); ok {
		return output, err
	}

	return nil, err
}
```

```go
// aws/resource_example_thing.go

function ExampleThingCreate(d *schema.ResourceData, meta interface{}) error {
	// ... AWS Go SDK logic to create resource ...

	if _, err := waiter.ThingCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Example Thing (%s) creation: %w", d.Id(), err)
	}

	return ExampleThingRead(d, meta)
}

function ExampleThingDelete(d *schema.ResourceData, meta interface{}) error {
	// ... AWS Go SDK logic to delete resource ...

	if _, err := waiter.ThingDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Example Thing (%s) deletion: %w", d.Id(), err)
	}

	return ExampleThingRead(d, meta)
}
```

Typically, the AWS Go SDK should include constants for various status field values (e.g. `StatusCreating` for `CREATING`). If not, create them in a file named `aws/internal/service/{SERVICE}/consts.go`.
