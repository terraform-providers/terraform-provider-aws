package aws

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLBTargetGroupAttachment_basic(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigTargetIdInstance(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachment_disappears(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigTargetIdInstance(targetGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
					testAccCheckAWSLBTargetGroupAttachmentDisappears("aws_lb_target_group_attachment.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachment_BackwardsCompatibility(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigBackwardsCompatibility(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_alb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachment_Port(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigPort(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachment_ipAddress(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigTargetIdIpAddress(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachment_lambda(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigTargetIdLambda(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func testAccCheckAWSLBTargetGroupAttachmentDisappears(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Attachment not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn
		targetGroupArn := rs.Primary.Attributes["target_group_arn"]

		target := &elbv2.TargetDescription{
			Id: aws.String(rs.Primary.Attributes["target_id"]),
		}

		_, hasPort := rs.Primary.Attributes["port"]
		if hasPort {
			port, _ := strconv.Atoi(rs.Primary.Attributes["port"])
			target.Port = aws.Int64(int64(port))
		}

		params := &elbv2.DeregisterTargetsInput{
			TargetGroupArn: aws.String(targetGroupArn),
			Targets:        []*elbv2.TargetDescription{target},
		}

		_, err := conn.DeregisterTargets(params)
		if err != nil && !isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
			return fmt.Errorf("Error deregistering Targets: %s", err)
		}

		return err
	}
}

func testAccCheckAWSLBTargetGroupAttachmentExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Target Group Attachment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		_, hasPort := rs.Primary.Attributes["port"]
		targetGroupArn := rs.Primary.Attributes["target_group_arn"]

		target := &elbv2.TargetDescription{
			Id: aws.String(rs.Primary.Attributes["target_id"]),
		}
		if hasPort {
			port, _ := strconv.Atoi(rs.Primary.Attributes["port"])
			target.Port = aws.Int64(int64(port))
		}

		describe, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupArn),
			Targets:        []*elbv2.TargetDescription{target},
		})

		if err != nil {
			return err
		}

		if len(describe.TargetHealthDescriptions) != 1 {
			return errors.New("Target Group Attachment not found")
		}

		return nil
	}
}

func testAccCheckAWSLBTargetGroupAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_target_group_attachment" && rs.Type != "aws_alb_target_group_attachment" {
			continue
		}

		_, hasPort := rs.Primary.Attributes["port"]
		targetGroupArn := rs.Primary.Attributes["target_group_arn"]

		target := &elbv2.TargetDescription{
			Id: aws.String(rs.Primary.Attributes["target_id"]),
		}
		if hasPort {
			port, _ := strconv.Atoi(rs.Primary.Attributes["port"])
			target.Port = aws.Int64(int64(port))
		}

		describe, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupArn),
			Targets:        []*elbv2.TargetDescription{target},
		})
		if err == nil {
			if len(describe.TargetHealthDescriptions) != 0 {
				return fmt.Errorf("Target Group Attachment %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") || isAWSErr(err, elbv2.ErrCodeInvalidTargetException, "") {
			return nil
		} else {
			return fmt.Errorf("Unexpected error checking LB destroyed: %s", err)
		}
	}

	return nil
}

func testAccAWSLBTargetGroupAttachmentConfigInstanceBase() string {
	return `
data "aws_availability_zones" "available" {
  # t2.micro instance type is not available in these Availability Zones
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

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

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.test.id
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-lb-target-group-attachment"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-lb-target-group-attachment"
  }
}
`
}

func testAccAWSLBTargetGroupAttachmentConfigTargetIdInstance(rName string) string {
	return testAccAWSLBTargetGroupAttachmentConfigInstanceBase() + fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
}
`, rName)
}

func testAccAWSLBTargetGroupAttachmentConfigPort(rName string) string {
	return testAccAWSLBTargetGroupAttachmentConfigInstanceBase() + fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}
`, rName)
}

func testAccAWSLBTargetGroupAttachmentConfigBackwardsCompatibility(rName string) string {
	return testAccAWSLBTargetGroupAttachmentConfigInstanceBase() + fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 443
  protocol = "HTTPS"
  vpc_id   = aws_vpc.test.id
}

resource "aws_alb_target_group_attachment" "test" {
  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_instance.test.id
  port             = 80
}
`, rName)
}

func testAccAWSLBTargetGroupAttachmentConfigTargetIdIpAddress(rName string) string {
	return testAccAWSLBTargetGroupAttachmentConfigInstanceBase() + fmt.Sprintf(`
resource "aws_lb_target_group" "test" {
  name        = %[1]q
  port        = 443
  protocol    = "HTTPS"
  target_type = "ip"
  vpc_id      = aws_vpc.test.id
}

resource "aws_lb_target_group_attachment" "test" {
  availability_zone = aws_instance.test.availability_zone
  target_group_arn  = aws_lb_target_group.test.arn
  target_id         = aws_instance.test.private_ip
}
`, rName)
}

func testAccAWSLBTargetGroupAttachmentConfigTargetIdLambda(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "elasticloadbalancing.${data.aws_partition.current.dns_suffix}"
  qualifier     = aws_lambda_alias.test.name
  source_arn    = aws_lb_target_group.test.arn
  statement_id  = "AllowExecutionFromlb"
}

resource "aws_lb_target_group" "test" {
  name        = %[1]q
  target_type = "lambda"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda_elb.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_elb.lambda_handler"
  runtime       = "python3.7"
}

resource "aws_lambda_alias" "test" {
  name             = "test"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "$LATEST"
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
	EOF

}

resource "aws_lb_target_group_attachment" "test" {
  depends_on = [aws_lambda_permission.test]

  target_group_arn = aws_lb_target_group.test.arn
  target_id        = aws_lambda_alias.test.arn
}
`, rName)
}
