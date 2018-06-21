package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSInstanceState_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAWSInstanceStateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAWSInstanceStateConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAWSInstanceStateCheckStopped("aws_instance_state.stopped"),
				),
			},
		},
	})
}
func testAWSInstanceStateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_instance" {
			continue
		}

		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
			continue
		}
		if err != nil {
			return err
		}

		for _, r := range resp.Reservations {
			for _, i := range r.Instances {
				if i.State != nil && *i.State.Name != "terminated" {
					return fmt.Errorf("Found unterminated instance: %s", i)
				}
			}
		}
	}

	return nil
}

func testAWSInstanceStateCheckStopped(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		state := rs.Primary.Attributes["state"]
		if state != "stopped" {
			return fmt.Errorf("Bad: instance in invalid state. Should be stopped is %s", state)
		}
		return nil
	}
}

func testAWSInstanceStateConfig() string {
	return fmt.Sprintf(`
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.66.0.0/16"
  tags {
    Name = "tf-acctest"
  }
}

resource "aws_subnet" "public_subnet" {
  vpc_id                  = "${aws_vpc.my_vpc.id}"
  cidr_block              = "172.66.0.0/24"
  availability_zone       = "us-west-2a"
  map_public_ip_on_launch = true
}

resource "aws_instance" "state_test" {
  ami                         = "ami-22b9a343"
  instance_type               = "t2.micro"
  subnet_id                   = "${aws_subnet.public_subnet.id}"
  associate_public_ip_address = false
}

resource "aws_instance_state" "stopped" {
  instance_id   = "${aws_instance.state_test.id}"
  state	        = "stopped"
  depends_on    = ["aws_instance.state_test"]
}
`)
}
