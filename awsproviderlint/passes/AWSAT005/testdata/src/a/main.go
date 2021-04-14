package a

import (
	"fmt"
)

func f() {
	/* Passing cases */
	fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:%v"
}
`, "policy/AmazonEKSClusterPolicy")

	/* Comment ignored cases */

	//lintignore:AWSAT005
	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:%v"`, "policy/AmazonEKSClusterPolicy")

	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:%v"`, "policy/AmazonEKSClusterPolicy") //lintignore:AWSAT005
	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:%v"`, "policy/AmazonEKSClusterPolicy") //lintignore:AWSAT005

	/* Failing cases */

	fmt.Sprintf(`policy_arn = "arn:aws:iam::aws:%v"`, "policy/AmazonEKSClusterPolicy")        // want "avoid hardcoded ARN AWS partitions"
	fmt.Sprintf(`policy_arn = "arn:aws-us-gov:iam::aws:%v"`, "policy/AmazonEKSClusterPolicy") // want "avoid hardcoded ARN AWS partitions"

}
