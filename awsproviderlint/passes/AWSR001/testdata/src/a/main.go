package a

import (
	"fmt"
)

func f() {
	/* Passing cases */

	fmt.Sprintf("%s.notamazonaws.com", "test")

	fmt.Sprintf(`
resource "aws_iam_role" "test" {
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
  name = %q
}`, "test")

	/* Comment ignored cases */

	//lintignore:AWSR001
	fmt.Sprintf("%s.amazonaws.com", "test")

	fmt.Sprintf("%s.amazonaws.com", "test") //lintignore:AWSR001

	/* Failing cases */

	fmt.Sprintf("%s.amazonaws.com", "test") // want "prefer \\(\\*AWSClient\\).PartitionHostname\\(\\) or \\(\\*AWSClient\\).RegionalHostname\\(\\)"
}
