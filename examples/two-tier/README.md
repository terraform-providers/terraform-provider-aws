# Basic Two-Tier AWS Architecture

This provides a template for running a simple two-tier architecture on Amazon
Web services. The premise is that you have stateless app servers running behind
an ELB serving traffic.

To simplify the example, this intentionally ignores deploying and
getting your application onto the servers. However, you could do so either via
[provisioners](https://www.terraform.io/docs/provisioners/) and a configuration
management tool, or by pre-baking configured AMIs with
[Packer](http://www.packer.io).

Please replace all strings REPLACE_WITH_YOUR with your data in variables.tf file.

This example will also create a new EC2 Key Pair in the specified AWS Region. 
The key name and path to the public key must be specified via the  
terraform variables in variables.tf.

After you run `terraform apply` on this configuration, it will
automatically output the DNS address of the ELB. After your instance
registers, this should respond with the default nginx web page.

To run, configure your AWS provider as described in 

https://www.terraform.io/docs/providers/aws/index.html

Run with a command like this:

```
terraform init
terraform apply
terraform destroy
```
# Requirements

AWS account with AWS access and AWS secret keys and administrative permissions on AWS region.

SSH key generated on Linux.
