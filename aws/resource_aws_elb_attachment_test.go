package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSELBAttachment_basic(t *testing.T) {
	var conf elb.LoadBalancerDescription

	testCheckInstanceAttached := func(count int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(conf.Instances) != count {
				return fmt.Errorf("instance count does not match")
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBAttachmentConfig1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(1),
				),
			},

			{
				Config: testAccAWSELBAttachmentConfig2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(2),
				),
			},

			{
				Config: testAccAWSELBAttachmentConfig3(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(2),
				),
			},

			{
				Config: testAccAWSELBAttachmentConfig4(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(0),
				),
			},
		},
	})
}

// remove and instance and check that it's correctly re-attached.
func TestAccAWSELBAttachment_drift(t *testing.T) {
	var conf elb.LoadBalancerDescription

	deregInstance := func() {
		conn := testAccProvider.Meta().(*AWSClient).elbconn

		deRegisterInstancesOpts := elb.DeregisterInstancesFromLoadBalancerInput{
			LoadBalancerName: conf.LoadBalancerName,
			Instances:        conf.Instances,
		}

		log.Printf("[DEBUG] deregistering instance %v from ELB", *conf.Instances[0].InstanceId)

		_, err := conn.DeregisterInstancesFromLoadBalancer(&deRegisterInstancesOpts)
		if err != nil {
			t.Fatalf("Failure deregistering instances from ELB: %s", err)
		}

	}

	testCheckInstanceAttached := func(count int) resource.TestCheckFunc {
		return func(*terraform.State) error {
			if len(conf.Instances) != count {
				return fmt.Errorf("instance count does not match")
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_elb.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSELBAttachmentConfig1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(1),
				),
			},

			// remove an instance from the ELB, and make sure it gets re-added
			{
				Config:    testAccAWSELBAttachmentConfig1(),
				PreConfig: deregInstance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSELBExists("aws_elb.bar", &conf),
					testCheckInstanceAttached(1),
				),
			},
		},
	})
}

// add one attachment
func testAccAWSELBAttachmentConfig1() string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.bar.id
  instance = aws_instance.foo1.id
}
`)
}

// add a second attachment
func testAccAWSELBAttachmentConfig2() string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_instance" "foo2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.bar.id
  instance = aws_instance.foo1.id
}

resource "aws_elb_attachment" "foo2" {
  elb      = aws_elb.bar.id
  instance = aws_instance.foo2.id
}
`)
}

// swap attachments between resources
func testAccAWSELBAttachmentConfig3() string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_instance" "foo1" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_instance" "foo2" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_elb_attachment" "foo1" {
  elb      = aws_elb.bar.id
  instance = aws_instance.foo2.id
}

resource "aws_elb_attachment" "foo2" {
  elb      = aws_elb.bar.id
  instance = aws_instance.foo1.id
}
`)
}

// destroy attachments
func testAccAWSELBAttachmentConfig4() string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elb" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  listener {
    instance_port     = 8000
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}
`)
}
