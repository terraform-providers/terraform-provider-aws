# AWS WorkSpaces Example

This example demonstrates how to create a WorkSpace and WorkSpace directory using AWS WorkSpaces.

## Note

The AWS WorkSpaces service requires an IAM role named `workspaces_DefaultRole`. By default, this example assumes the role exists and is configured. If the role does not exist or is not configured appropriately, the example will not successfully deploy.

The IAM resources are defined in the Terraform source file [iam.tf](./iam.tf), but are commented out. If the role does not exist in your environment, uncomment the contents of the file. The resources `aws_workspaces_directory.example` and `aws_workspaces_workspace.example` have dependencies on the IAM resources. The `depends_on` meta-arguments should also be uncommented on both.

## Running this example

This example can be run by calling

```shell
terraform init
terraform apply
```

By default, resources are created in the `us-west-2` region. To override the region, set the variable `aws_region` to a different value.
