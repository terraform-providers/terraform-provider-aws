# Cognito User Pool example

This example creates a Cognito User Pool, IAM roles and lambdas.

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

Running the example

For planning phase

```
terraform plan
```

For apply phase

```
terraform apply
```

To remove the stack

```
 terraform destroy -var 'key_name={your_key_name}'
```
