# AWSAT003

The AWSAT003 analyzer reports hardcoded AWS regions. Tests that are hardcoded
to work in a region specific to a partition (eg, the AWS standard/commercial
partition) will fail in other partitions where the region does not exist (eg,
GovCloud).

## Flagged Code

The `us-west-2` region does not exist in non-standard partitions (eg, the
GovCloud partition).

```go
fmt.Sprintf(`
resource "aws_config_configuration_aggregator" "example" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = ["us-west-2"]
  }
}

data "aws_caller_identity" "current" {}
`, rName)
```

Hardcoded regions (eg, `us-west-2`) that are part of an availability zone (AZ)
designation (eg., `us-west-2a`) are also flagged.

```go
fmt.Sprintf(`
resource "aws_subnet" "test" {
  availability_zone = "us-west-2a"
  cidr_block        = %q
}
`, "10.0.0.0/24")
```

## Passing Code

```go
fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = %q
}
`, "10.0.0.0/24")
```

## Ignoring Reports

Singular reports can be ignored by adding the a `//lintignore:AWSAT003` Go code comment at the end of the offending line or on the line immediately proceding, e.g.

```go
fmt.Sprintf(`"af-south-1":     %q,`, "525921808201") //lintignore:AWSAT003
```
