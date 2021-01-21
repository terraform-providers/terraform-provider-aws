# AWSAT005

The AWSAT005 analyzer reports hardcoded AWS partitions in ARNs. For tests to
work across AWS partitions, the partitions should not be hardcoded.

## Flagged Code

```go
func testAccAWSSpotFleetRequestConfig(role string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = %q
}
`, role)
}
```

## Passing Code

```go
func testAccAWSSpotFleetRequestConfig(role string) string {
    return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = %q
}
`, role)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT005` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy" //lintignore:AWSAT005
```
