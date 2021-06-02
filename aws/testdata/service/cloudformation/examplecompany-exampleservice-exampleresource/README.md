# ExampleCompany::ExampleService::ExampleResource

This directory contains the source for a valid CloudFormation Type. The files were generated with the [CloudFormation CLI](https://docs.aws.amazon.com/cloudformation-cli/latest/userguide/what-is-cloudformation-cli.html) `cfn init` command.

These files are for acceptance testing the Terraform `aws_cloudformation_type` managed resource and data source. The `bin/handler` file is included in the `.gitignore` since it is large. To generate it, run `make build` in this directory.

Since the deployed `.rpdk-config` and `schema.json` must match the `RegisterType` API `TypeName` parameter, the acceptance testing includes a helper function, `testAccAwsCloudformationTypeZipGenerator`, which can create valid zip files to match randomized `TypeName` and prevent race conditions in the testing.
