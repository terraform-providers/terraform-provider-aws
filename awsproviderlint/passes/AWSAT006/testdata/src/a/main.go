package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
	name = "%s"

	assume_role_policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
	{
		"Effect": "Allow",
		"Principal": {
		"Service": "eks.${data.aws_partition.current.dns_suffix}"
		},
		"Action": "sts:AssumeRole"
	}
	]
}
POLICY
}	
`, "Misericordiam")

	/* Comment ignored cases */

	//lintignore:AWSAT006
	fmt.Sprintf(`service = "%v.amazonaws.com"`, "eks")

	fmt.Sprintf(`service = "%v.amazonaws.com"`, "eks") //lintignore:AWSAT006
	fmt.Sprintf(`service = "%v.amazonaws.com"`, "eks") //lintignore:AWSAT006

	/* Failing cases */

	fmt.Sprintf(`service = "%v.amazonaws.com"`, "eks") // want "avoid hardcoding AWS partition DNS suffixes"

}
