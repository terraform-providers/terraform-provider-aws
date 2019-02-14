package aws

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLBTargetGroupAttachment_basic(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfig_basic(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSLBTargetGroupAttachmentBackwardsCompatibility(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_alb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupAttachmentDestroy,
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

func TestAccAWSLBTargetGroupAttachment_withoutPort(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigWithoutPort(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroupAttachment_ipAddress(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigWithIpAddress(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
}

func TestAccAWSALBTargetGroupAttachment_lambda(t *testing.T) {
	targetGroupName := fmt.Sprintf("test-target-group-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_target_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBTargetGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBTargetGroupAttachmentConfigWithLambda(targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBTargetGroupAttachmentExists("aws_lb_target_group_attachment.test"),
				),
			},
		},
	})
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

func testAccAWSLBTargetGroupAttachmentConfigWithoutPort(targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = "${aws_lb_target_group.test.arn}"
  target_id = "${aws_instance.test.id}"
}

resource "aws_instance" "test" {
  ami = "ami-f701cb97"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.subnet.id}"
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"
  deregistration_delay = 200
  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }
  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

resource "aws_subnet" "subnet" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-lb-target-group-attachment-without-port"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-attachment-without-port"
	}
}`, targetGroupName)
}

func testAccAWSLBTargetGroupAttachmentConfig_basic(targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = "${aws_lb_target_group.test.arn}"
  target_id = "${aws_instance.test.id}"
  port = 80
}

resource "aws_instance" "test" {
  ami = "ami-f701cb97"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.subnet.id}"
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"
  deregistration_delay = 200
  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }
  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

resource "aws_subnet" "subnet" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-lb-target-group-attachment-basic"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-attachment-basic"
	}
}`, targetGroupName)
}

func testAccAWSLBTargetGroupAttachmentConfigBackwardsCompatibility(targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_alb_target_group_attachment" "test" {
  target_group_arn = "${aws_alb_target_group.test.arn}"
  target_id = "${aws_instance.test.id}"
  port = 80
}

resource "aws_instance" "test" {
  ami = "ami-f701cb97"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.subnet.id}"
}

resource "aws_alb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"
  deregistration_delay = 200
  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }
  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

resource "aws_subnet" "subnet" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-lb-target-group-attachment-bc"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-attachment-bc"
	}
}`, targetGroupName)
}

func testAccAWSLBTargetGroupAttachmentConfigWithIpAddress(targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group_attachment" "test" {
  target_group_arn = "${aws_lb_target_group.test.arn}"
  target_id = "${aws_instance.test.private_ip}"
  availability_zone = "${aws_instance.test.availability_zone}"
}
resource "aws_instance" "test" {
  ami = "ami-f701cb97"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.subnet.id}"
}
resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 443
  protocol = "HTTPS"
  vpc_id = "${aws_vpc.test.id}"
  target_type = "ip"
  deregistration_delay = 200
  stickiness {
    type = "lb_cookie"
    cookie_duration = 10000
  }
  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}
resource "aws_subnet" "subnet" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-lb-target-group-attachment-with-ip-address"
  }
}
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-lb-target-group-attachment-with-ip-address"
	}
}`, targetGroupName)
}

func testAccAWSLBTargetGroupAttachmentConfigWithLambda(targetGroupName string) string {
	funcName := fmt.Sprintf("tf_acc_lambda_func_%s", acctest.RandString(8))

	return fmt.Sprintf(`

	resource "aws_lambda_permission" "with_lb" {
		statement_id = "AllowExecutionFromlb"
		action = "lambda:InvokeFunction"
		function_name = "${aws_lambda_function.test.arn}"
		principal = "elasticloadbalancing.amazonaws.com"
		source_arn = "${aws_lb_target_group.test.arn}"
		qualifier     = "${aws_lambda_alias.test.name}"
	}

	resource "aws_lb_target_group" "test" {
		name = "%s"
		target_type = "lambda" 
	}

	resource "aws_lambda_function" "test" {
		filename = "test-fixtures/lambda_elb.zip"
		function_name = "%s"
		role = "${aws_iam_role.iam_for_lambda.arn}"
		handler = "lambda_elb.lambda_handler"
		runtime = "python3.7"
	}

	resource "aws_lambda_alias" "test" {
		name             = "test"
		description      = "a sample description"
		function_name    = "${aws_lambda_function.test.function_name}"
		function_version = "$LATEST"
	}
	
	resource "aws_iam_role" "iam_for_lambda" {
			assume_role_policy = <<EOF
{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}
		]
	}
	EOF
}

	resource "aws_lb_target_group_attachment" "test" {
		target_group_arn = "${aws_lb_target_group.test.arn}"
		target_id = "${aws_lambda_alias.test.arn}"
		depends_on = ["aws_lambda_permission.with_lb"]
	}
`, targetGroupName, funcName)
}
