# AWSAT006

The AWSAT006 analyzer reports hardcoded AWS partition DNS suffixes. For tests
to work across AWS partitions, the DNS suffixes should not be hardcoded.

## Flagged Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}
`, name)
}
```

## Passing Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
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
`, name)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT006` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
service = "eks.amazonaws.com" //lintignore:AWSAT006
```
