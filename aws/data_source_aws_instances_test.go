package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSInstancesDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_ids,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "3"),
					resource.TestCheckResourceAttr("data.aws_instances.test", "private_ips.#", "3"),
					// Public IP values are flakey for new EC2 instances due to eventual consistency
					resource.TestCheckResourceAttrSet("data.aws_instances.test", "public_ips.#"),
				),
			},
		},
	})
}

func TestAccAWSInstancesDataSource_tags(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_tags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSInstancesDataSource_instance_state_names(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_instance_state_names(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

const testAccInstancesDataSourceConfig_ids = `
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "test" {
  count = 3
  ami = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  tags = {
    Name = "TfAccTest"
  }
}

data "aws_instances" "test" {
  filter {
    name = "instance-id"
    values = [
      "${aws_instance.test.*.id[0]}",
      "${aws_instance.test.*.id[1]}",
      "${aws_instance.test.*.id[2]}",
    ]
  }
}
`

func testAccInstancesDataSourceConfig_tags(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "test" {
  count = 2
  ami = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  tags = {
    Name      = "tf-acc-test-ec2-instance-data-source-%d"
    SecondTag = "tf-acc-test-ec2-instance-data-source-%d"
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name      = "${aws_instance.test.0.tags["Name"]}"
    SecondTag = "${aws_instance.test.1.tags["Name"]}"
  }
}
`, rInt, rInt)
}

func testAccInstancesDataSourceConfig_instance_state_names(rInt int) string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "test" {
  count = 2
  ami = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  tags = {
    Name = "tf-acc-test-ec2-instance-data-source-%d"
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name = "${aws_instance.test.0.tags["Name"]}"
  }
  
  instance_state_names = [ "pending", "running" ]
}
`, rInt)
}
