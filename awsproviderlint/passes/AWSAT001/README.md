# AWSAT001

The `AWSAT001` analyzer reports when a `resource.TestMatchResourceAttr()` call references an Amazon
Resource Name (ARN) attribute. It is preferred to use `resource.TestCheckResourceAttrPair()` or one
one of the available Terraform AWS Provider ARN testing check functions instead building full ARN
regular expressions. These testing helper functions consider the value of the AWS Account ID,
Partition, and Region of the acceptance test runner.

The `resource.TestCheckResourceAttrPair()` call can be used when the Terraform state has the ARN
value already available, such as when the current resource is referencing an ARN attribute of
another resource.

Otherwise, available ARN testing check functions include:

- `testAccCheckResourceAttrGlobalARN`
- `testAccCheckResourceAttrGlobalARNNoAccount`
- `testAccCheckResourceAttrRegionalARN`
- `testAccMatchResourceAttrGlobalARN`
- `testAccMatchResourceAttrRegionalARN`
- `testAccMatchResourceAttrRegionalARNNoAccount`

## Flagged Code

```go
resource.TestMatchResourceAttr("aws_lb_listener.test", "certificate_arn", regexp.MustCompile(`^arn:[^:]+:acm:[^:]+:[^:]+:certificate/.+$`))
```

## Passing Code

```go
resource.TestCheckResourceAttrPair("aws_lb_listener.test", "certificate_arn", "aws_acm_certificate.test", "arn")

testAccMatchResourceAttrRegionalARN("aws_lb_listener.test", "certificate_arn", "acm", regexp.MustCompile(`certificate/.+`))
```

## Ignoring Check

The check can be ignored for a certain line via a `//lintignore:AWSAT001` comment on the previous line or at the end of the offending line, e.g.

```go
//lintignore:AWSAT001
resource.TestMatchResourceAttr("aws_lb_listener.test", "certificate_arn", regexp.MustCompile(`^arn:[^:]+:acm:[^:]+:[^:]+:certificate/.+$`))
```
