# Elastic IP Example

The eip example launches a web server, installs nginx. It also creates security group 

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

Running the example

run `terraform apply -var 'key_name={your_key_name}'` 

Alternatively to using `-var` with each command, the `terraform.template.tfvars` file can be copied to `terraform.tfvars` and updated.

Give couple of mins for userdata to install nginx, and then type the Elastic IP from outputs in your browser and see the nginx welcome page
