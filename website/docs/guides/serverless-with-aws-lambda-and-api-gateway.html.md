---
layout: "aws"
page_title: "Serverless with AWS Lambda and API Gateway"
sidebar_current: "docs-aws-guide-serverless"
description: |-
  Using Terraform to configure a serverless application with AWS Lambda and API Gateway.
---

# Serverless Applications with AWS Lambda and API Gateway

_Serverless computing_ is a cloud computing model in which a cloud provider
automatically manages the provisioning and allocation of compute resources.
This contrasts with traditional cloud computing where the user is responsible
for directly managing virtual servers.

A popular approach to running "serverless" web applications is to implement
the application functionality as one or more functions in
[AWS Lambda](https://aws.amazon.com/lambda/) and then expose these for public
consumption using [Amazon API Gateway](https://aws.amazon.com/api-gateway/).

This guide will show how to deploy such an architecture using Terraform. The
guide assumes some basic familiarity with Lambda and API Gateway but does not
assume any pre-existing deployment. It also assumes that you are familiar
with the usual Terraform plan/apply workflow; if you're new to Terraform
itself, refer first to [the Getting Started guide](/intro/getting-started/install.html).

This is a slightly-opinionated guide, which chooses to ignore the built-in
versioning and staged deployment mechanisms in AWS Lambda and API Gateway.
In many cases these features are not necessary when using Terraform because
changes can be tracked and deployed by keeping the Terraform configuration
in a version-control repository. It also uses API Gateway in a very simple
way, proxying all requests to a single AWS Lambda function that is expected
to contain its own request routing logic.

As usual, there are other valid ways to use these services that make different
tradeoffs. We encourage readers to consult the official documentation for
the respective services for additional context and best-practices. This guide
can still serve as an introduction to the main resources associated with
these services, even if you choose a different architecture.

## Preparation

In order to follow this guide you will need an AWS account and to have
Terraform installed.
[Configure your credentials](/docs/providers/aws/index.html#authentication)
so that Terraform is able to act on your behalf.

For simplicity here we will assume you are already using a set of IAM
credentials with suitable access to create Lambda functions and work with API
Gateway. If you aren't sure and are working in an AWS account used only for
development, the simplest approach to get started is to use credentials with
full administrative access to the target AWS account.

In the following section we will manually emulate an automatic build process
using the `zip` command line tool and the [AWS CLI](https://aws.amazon.com/cli/).
The latter must also have access to your AWS credentials, and the easiest way
to achieve this is to provide them via environment variables so that they can
be used by both the AWS CLI and Terraform.

~> **Warning:** Following this tutorial will create objects in your AWS account
that will cost you money against your AWS bill.

## Building the Lambda Function Package

AWS Lambda expects a function's implementation to be provided as an archive
containing the function source code and any other static files needed to
execute the function.

Terraform is not a build tool, so the zip file must be prepared using a
separate build process prior to deploying it with Terraform. For a real
application we recommend automating your build via a CI system, whose job
is to run any necessary build actions (library installation, compilation, etc),
produce the deployment zip file as a build artifact, and then upload that
artifact into an Amazon S3 bucket from which it will be read for deployment.

For the sake of this tutorial we will perform these build steps manually and
build a very simple AWS Lambda function. Start by creating a new directory
called `example` that will be used to create the archive, and place in it a
single source file. We will use the JavaScript runtime in this example, so
our file is called `main.js` and will contain the following source code:

```js
'use strict';

exports.handler = function (event, context, callback) {
    var response = {
        statusCode: 200,
        headers: {
            'Content-Type': 'text/html; charset=utf-8',
        },
        body: "<p>Hello world!</p>",
    };
    callback(null, response);
};
```

The above is the simplest possible Lambda function for use with API Gateway,
returning a hard-coded "Hello world!" response in the object structure that
API Gateway expects.

From your command prompt, change to the directory containing that file and
add it to a zip file in the parent directory:

```
$ cd example
$ zip ../example.zip main.js
  adding: main.js (deflated 33%)
$ cd ..
```

In a real build and deploy scenario we would have an S3 bucket set aside for
staging our archive and would use this to "hand off" these artifacts between
the build and deploy process. For the sake of this tutorial we will create
a temporary S3 bucket using the AWS CLI. S3 bucket names are globally unique,
so you may need to change the `--bucket=` argument in the following example
and substitute your new bucket name throughout the rest of this tutorial.

```
$ aws s3api create-bucket --bucket=terraform-serverless-example --region=us-east-1
```

You can now upload your build artifact into this S3 bucket:

```
$ aws s3 cp example.zip s3://terraform-serverless-example/v1.0.0/example.zip
```

A version number is included in the object path to identify this build. Later
we will demonstrate deploying a new version, which will create another
separate object.

## Creating the Lambda Function

With the source code artifact built and uploaded to S3, we can now write our
Terraform configuration to deploy it. In a new directory, create a file
named `lambda.tf` containing the following configuration:

```hcl
provider "aws" {
  region = "us-east-1"
}

resource "aws_lambda_function" "example" {
  function_name = "ServerlessExample"

  # The bucket name as created earlier with "aws s3api create-bucket"
  s3_bucket = "terraform-serverless-example"
  s3_key    = "v1.0.0/example.zip"

  # "main" is the filename within the zip file (main.js) and "handler"
  # is the name of the property under which the handler function was
  # exported in that file.
  handler = "main.handler"
  runtime = "nodejs6.10"

  role = "${aws_iam_role.lambda_exec.arn}"
}

# IAM role which dictates what other AWS services the Lambda function
# may access.
resource "aws_iam_role" "lambda_exec" {
  name = "serverless_example_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
```

Each Lambda function must have an associated IAM role which dictates what
access it has to other AWS services. The above configuration specifies a role
with no access policy, effectively giving the function no access to any
AWS services, since our example application requires no such access.

Before you can work with a new configuration directory, it must be initialized
using `terraform init`, which in this case will install the AWS provider:

```
$ terraform init

Initializing provider plugins...
- Checking for available provider plugins on https://releases.hashicorp.com...
- Downloading plugin for provider "aws" (1.9.0)...

# ...

Terraform has been successfully initialized!

# ...
```

Now apply the configuration as usual:

```
$ terraform apply

# ....

aws_iam_role.lambda_exec: Creating...
  arn:                   "" => "<computed>"
  assume_role_policy:    "" => "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Action\": \"sts:AssumeRole\",\n      \"Principal\": {\n        \"Service\": \"lambda.amazonaws.com\"\n      },\n      \"Effect\": \"Allow\",\n      \"Sid\": \"\"\n    }\n  ]\n}\n"
  create_date:           "" => "<computed>"
  force_detach_policies: "" => "false"
  name:                  "" => "serverless_example_lambda"
  path:                  "" => "/"
  unique_id:             "" => "<computed>"
aws_iam_role.lambda_exec: Creation complete after 1s (ID: serverless_example_lambda)
aws_lambda_function.example: Creating...
  arn:              "" => "<computed>"
  function_name:    "" => "ServerlessExample"
  handler:          "" => "main.handler"
  invoke_arn:       "" => "<computed>"
  last_modified:    "" => "<computed>"
  memory_size:      "" => "128"
  publish:          "" => "false"
  qualified_arn:    "" => "<computed>"
  role:             "" => "arn:aws:iam::123456:role/serverless_example_lambda"
  runtime:          "" => "nodejs6.10"
  s3_bucket:        "" => "terraform-serverless-example"
  s3_key:           "" => "v1.0.0/example.zip"
  source_code_hash: "" => "<computed>"
  timeout:          "" => "3"
  tracing_config.#: "" => "<computed>"
  version:          "" => "<computed>"
aws_lambda_function.example: Still creating... (10s elapsed)
aws_lambda_function.example: Creation complete after 11s (ID: ServerlessExample)

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
```

After the function is created successfully, try invoking it using the AWS
CLI:

```
$ aws lambda invoke --region=us-east-1 --function-name=ServerlessExample output.txt
{"StatusCode": 200}
$ cat output.txt
{
  "statusCode":200,
  "headers":{
    "Content-Type":"text/html; charset=utf-8"
  },
  "body":"<p>Hello world!</p>"
}
```

With the function working as expected, the next step is to create the API
Gateway REST API that will provide access to it.

## Configuring API Gateway

API Gateway's name reflects its original purpose as a public-facing frontend
for REST APIs, but it was later extended with features that make it easy to
expose an entire web application based on AWS Lambda. These later features
will be used in this tutorial. The term "REST API" is thus used loosely
here, since API Gateway is serving as a generic HTTP frontend rather than
necessarily serving an API.

Create a new file `api_gateway.tf` in the same directory as our `lambda.tf`
from the previous step. First, configure the root "REST API" object, as follows:

```hcl
resource "aws_api_gateway_rest_api" "example" {
  name        = "ServerlessExample"
  description = "Terraform Serverless Application Example"
}
```

The "REST API" is the container for all of the other API Gateway objects we will
create.

All incoming requests to API Gateway must match with a configured resource and
method in order to be handled. Append the following to the `lambda.tf` file to
define a single proxy resource:

```hcl
resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  parent_id   = "${aws_api_gateway_rest_api.example.root_resource_id}"
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "proxy" {
  rest_api_id   = "${aws_api_gateway_rest_api.example.id}"
  resource_id   = "${aws_api_gateway_resource.proxy.id}"
  http_method   = "ANY"
  authorization = "NONE"
}
```

The special `path_part` value `"{proxy+}"` activates proxy behavior, which
means that this resource will match _any_ request path. Similarly, the
`aws_api_gateway_method` block uses a `http_method` of `"ANY"`, which allows
any request method to be used. Taken together, this means that all incoming
requests will match this resource.

Each method on an API gateway resource has an _integration_ which specifies
where incoming requests are routed. Add the following configuration to specify
that requests to this method should be sent to the Lambda function defined
earlier:

```hcl
resource "aws_api_gateway_integration" "lambda" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  resource_id = "${aws_api_gateway_method.proxy.resource_id}"
  http_method = "${aws_api_gateway_method.proxy.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.example.invoke_arn}"
}
```

The `AWS_PROXY` integration type causes API gateway to call into the API of
another AWS service. In this case, it will call the AWS Lambda API to create
an "invocation" of the Lambda function.

Unfortunately the proxy resource cannot match an _empty_ path at the root of
the API. To handle that, a similar configuration must be applied to the
_root resource_ that is built in to the REST API object:

```hcl
resource "aws_api_gateway_method" "proxy_root" {
  rest_api_id   = "${aws_api_gateway_rest_api.example.id}"
  resource_id   = "${aws_api_gateway_rest_api.example.root_resource_id}"
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_root" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  resource_id = "${aws_api_gateway_method.proxy_root.resource_id}"
  http_method = "${aws_api_gateway_method.proxy_root.http_method}"

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "${aws_lambda_function.example.invoke_arn}"
}
```

Finally, you need to create an API Gateway "deployment" in order to activate
the configuration and expose the API at a URL that can be used for testing:

```hcl
resource "aws_api_gateway_deployment" "example" {
  depends_on = [
    "aws_api_gateway_integration.lambda",
    "aws_api_gateway_integration.lambda_root",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  stage_name  = "test"
}
```

With all of the above configuration changes in place, run `terraform apply`
again to create these new objects:

```
$ terraform apply

# ...

aws_api_gateway_rest_api.example: Creating...
  created_date:     "" => "<computed>"
  description:      "" => "Terraform Serverless Application Example"
  name:             "" => "ServerlessExample"
  root_resource_id: "" => "<computed>"
aws_api_gateway_rest_api.example: Creation complete after 1s (ID: bkqhuuz8r8)

# ...etc, etc...

Apply complete! Resources: 5 added, 0 changed, 0 destroyed.
```

After the creation steps are complete, the new objects will be visible in
[the API Gateway console](https://console.aws.amazon.com/apigateway/home?region=us-east-1).

The integration with the Lambda function is not functional yet because
API Gateway does not have the necessary access to invoke the function.
The next step will address this, making the application fully-functional.

## Allowing API Gateway to Access Lambda

By default any two AWS services have no access to one another, until access
is explicitly granted. For Lambda functions, access is granted using the
`aws_lambda_permission` resource, which should be added to the `lambda.tf`
file created in an earlier step:

```hcl
resource "aws_lambda_permission" "apigw" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.example.arn}"
  principal     = "apigateway.amazonaws.com"

  # The /*/* portion grants access from any method on any resource
  # within the API Gateway "REST API".
  source_arn = "${aws_api_gateway_deployment.example.execution_arn}/*/*"
}
```

In order to test the created API you will need to access its test URL. To
make this easier to access, add the following output to `api_gateway.tf`:

```
output "base_url" {
  value = "${aws_api_gateway_deployment.example.invoke_url}"
}
```

Apply the latest changes with `terraform apply`:

```
$ terraform apply

# ...

aws_lambda_permission.apigw: Creating...
  statement_id:  "" => "AllowAPIGatewayInvoke"
  action:        "" => "lambda:InvokeFunction"
  function_name: "" => "ServerlessExample"
# ...
aws_lambda_permission.apigw: Creation complete after 1s

Apply complete! Resources: 1 added, 0 changed, 1 destroyed.

Outputs:

base_url = https://bkqhuuz8r8.execute-api.us-east-1.amazonaws.com/test
```

Load the URL given in the output from _your_ run in your favorite web browser.
If everything has worked, you will see the text "Hello world!". This message
is being returned from the Lambda function code uploaded earlier, via the
API Gateway endpoint.

This is a good milestone! The first version of the application is deployed and
accessible. Next we will see how to deploy a new version of the application.

## A New Version of the Lambda Function

For any real application there will inevitably be changes to the application
code over time, which must then be deployed to AWS Lambda in place of the
previous version.

Returning to the `example` directory containing the `main.js` from earlier,
update the source code to change the message. For example:

```js
'use strict';

exports.handler = function (event, context, callback) {
    var response = {
        statusCode: 200,
        headers: {
            'Content-Type': 'text/html; charset=utf-8',
        },
        body: "<p>Bonjour au monde!</p>",
    };
    callback(null, response);
};
```

Update the zip file and upload a new version to the artifact S3 bucket:

```
$ cd example
$ zip ../example.zip main.js
updating: main.js (deflated 33%)
$ cd ..
$ aws s3 cp example.zip s3://terraform-serverless-example/v1.0.1/example.zip
```

Notice that a different version number was used in the S3 object path, so
the previous archive is retained. In order to allow easy switching between
versions you can define a variable to allow the version number to be chosen
dynamically. Add the following to `lambda.tf`:

```hcl
variable "app_version" {
}
```

Then locate the `aws_lambda_function` resource defined earlier and change
its `s3_key` argument to include the version variable:

```hcl
resource "aws_lambda_function" "example" {
  function_name = "ServerlessExample"

  # The bucket name as created earlier with "aws s3api create-bucket"
  s3_bucket = "terraform-serverless-example"
  s3_key    = "v${var.app_version}/example.zip"

  # (leave the remainder unchanged)
}
```

The `terraform apply` command now requires a version number to be provided:

```
$ terraform apply -var="app_version=1.0.1"

# ...

Terraform will perform the following actions:

  ~ aws_lambda_function.example
      s3_key: "v1.0.0/example.zip" => "v1.0.1/example.zip"

Plan: 0 to add, 1 to change, 0 to destroy.

# ...
```

After the change has been applied, visit again the test URL and you should
see the updated greeting message.

## Rolling Back to an Older Version

Sometimes new code doesn't work as expected and the simplest path is to
return to the previous version. Because all of the historical versions of
the artifact are preserved on S3, the original version can be restored with
a single command:

```
$ terraform apply -var="app_version=1.0.0"
```

After this apply completes, the test URL will return the original message
again.

## Conclusion

In this guide you created an AWS Lambda function that produces a result
compatible with Amazon API Gateway _proxy resources_ and then configured
API Gateway.

Although the AWS Lambda function used in this guide is very simple, in more
practical applications it is possible to use helper libraries to map
API Gateway proxy requests to standard HTTP application APIs in various
languages, such as [Python's WSGI](https://pypi.python.org/pypi/aws-wsgi/0.0.6)
or [the NodeJS Express Framework](https://github.com/awslabs/aws-serverless-express).

When combined with an automated build process running in a CI system, Terraform
can help to deploy applications as AWS Lambda functions, with suitable IAM
policies to connect with other AWS services for persistent storage, access to
secrets, etc.

## Cleaning Up

Once you are finished with this guide, you can destroy the example objects
with Terraform. Since our configuration requires a version number as an
input variable, provide a placeholder value to destroy:

```
$ terraform destroy -var="app_version=0.0.0"
```

Since the artifact zip files and the S3 bucket itself were created
outside of Terraform, they must also be cleaned up outside of Terraform. This
can be done via [the S3 console](https://s3.console.aws.amazon.com/s3/home).
Note that all of the objects in the bucket must be deleted before the bucket
itself can be deleted.

## Further Reading

The following Terraform resource types are used in this tutorial:

* [`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html)
* [`aws_lambda_permission`](/docs/providers/aws/r/lambda_permission.html)
* [`aws_api_gateway_rest_api`](/docs/providers/aws/r/api_gateway_rest_api.html)
* [`aws_api_gateway_resource`](/docs/providers/aws/r/api_gateway_resource.html)
* [`aws_api_gateway_method`](/docs/providers/aws/r/api_gateway_method.html)
* [`aws_api_gateway_integration`](/docs/providers/aws/r/api_gateway_integration.html)
* [`aws_iam_role`](/docs/providers/aws/r/iam_role.html)

The reference page for each resource type provides full details on all of its
supported arguments and exported attributes.

### Custom Domain Names and TLS Certificates

For the sake of example, this guide uses the test URLs offered by default
by API Gateway. In practice, most applications will be deployed at a custom
hostname.

To use a custom domain name you must first register that domain and configure
DNS hosting for it. You must also either create an
[Amazon Certificate Manager](https://aws.amazon.com/certificate-manager/)
certificate or register a TLS certificate with a third-party certificate
authority.

Configuring the domain name is beyond the scope of this tutorial, but if
you already have a hostname and TLS certificate you wish to use then you can
register it with API Gateway using the
[`aws_api_gateway_domain_name`](/docs/providers/aws/r/api_gateway_domain_name.html)
resource type.

A registered domain name is then mapped to a particular "REST API" object using
[`aws_api_gateway_base_path_mapping`](/docs/providers/aws/r/api_gateway_base_path_mapping.html).
The configured domain name then becomes an alias for a particular deployment
stage.

### Making Changes to the API Gateway Configuration

This guide creates a very simple API Gateway Configuration with a single
resource that passes through all requests to a single destination. The upgrade
steps then modify only the AWS Lambda function, leaving the API Gateway
configuration unchanged.

Due to API Gateway's staged deployment model, if you _do_ need to make changes
to the API Gateway configuration you must explicitly request that it be
re-deployed by "tainting" the deployment resource:

```
$ terraform taint aws_api_gateway_deployment.example
```

This command flags that this object must be re-created in the next Terraform
plan, so a subsequent `terraform apply` will then replace the deployment and
thus activate the latest configuration changes.

Please note that this "re-deployment" will cause some downtime, since Terraform
will need to delete the stage and associated deployment before re-creating it.
Downtime can be avoided by triggering the deployment action via the API Gateway
console, outside of Terraform. The approach covered in this guide intentionally
minimizes the need to amend the API Gateway configuration over time to
mitigate this limitation. Better support for this workflow will be added
to Terraform's AWS provider in a future release.
