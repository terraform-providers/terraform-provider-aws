# Lambda Example

This examples shows how to deploy an AWS Lambda function connected with an EFS file system using Terraform only.

To run, configure your AWS provider as described in https://www.terraform.io/docs/providers/aws/index.html

### Running the example

run `terraform apply` to see it work.

### Test the lambda function

```bash
 aws lambda invoke --region us-east-1 --function-name hello_lambda response.json 
```
Invoke lambda function several times, check the content of response.json. 

On each invoke, the lambda function will append one line of data to /mnt/efs/test.txt, and return all content from the file.
