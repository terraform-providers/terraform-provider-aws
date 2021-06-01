# AWSAT002

The AWSAT002 analyzer reports hardcoded AMI IDs. AMI IDs are region dependent and tests will fail in any region or partition other than where the AMI was created.

## Flagged Code

```go
func testAccAWSSpotFleetRequestConfig(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
    }
}
`, validUntil)
}
```

## Passing Code

```go
func testAccAWSSpotFleetRequestConfig(rName string, rInt int, validUntil string) string {
    return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_spot_fleet_request" "test" {
    launch_specification {
        instance_type = "m1.small"
        ami = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    }
}
`, validUntil)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT002` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
ami = "ami-516b9131" //lintignore:AWSAT002
```
