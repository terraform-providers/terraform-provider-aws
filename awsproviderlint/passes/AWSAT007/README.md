# AWSAT007

The AWSAT007 analyzer reports hardcoded instance types. For tests
to work across AWS partitions, instance types should not be hardcoded.

## Flagged Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
resource "aws_db_instance" "bar" {
  allocated_storage   = 10
  engine              = "MySQL"
  engine_version      = "5.6.35"
  instance_class      = "db.t2.micro"
  name                = %q
  password            = "barbarbarbar"
  ...
}
`, name)
}
```

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
resource "aws_launch_configuration" "data_source_aws_autoscaling_group_test" {
  name          = "%[1]s"
  image_id      = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
}
`, name)
}
```

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  instance_type = "t2.micro"

  tags = {
    Name = %q
  }
}
`, name)
}
```




## Passing Code

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro"]
}

resource "aws_instance" "test" {
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  ...
}
`, name)
}
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT007` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
func testAccAWSMisericordiamHumilitatemPulchritudo(name string) string {
    //lintignore:AWSAT007
    return fmt.Sprintf(`
resource "aws_instance" "test" {
  instance_type = "t2.micro"

  tags = {
    Name = %q
  }
}
`, name)
}
```
