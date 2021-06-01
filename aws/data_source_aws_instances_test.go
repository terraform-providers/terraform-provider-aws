package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSInstancesDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_ids(rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSInstancesDataSource_instanceStateNames(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstancesDataSourceConfig_instanceStateNames(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instances.test", "ids.#", "2"),
				),
			},
		},
	})
}

func testAccInstancesDataSourceConfig_ids(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 3
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %q
  }
}

data "aws_instances" "test" {
  filter {
    name   = "instance-id"
    values = aws_instance.test[*].id
  }
}
`, rName))
}

func testAccInstancesDataSourceConfig_tags(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name      = %[1]q
    SecondTag = "%[1]s-2"
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name      = aws_instance.test[0].tags["Name"]
    SecondTag = aws_instance.test[0].tags["SecondTag"]
  }

  depends_on = [aws_instance.test]
}
`, rName))
}

func testAccInstancesDataSourceConfig_instanceStateNames(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  count         = 2
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %q
  }
}

data "aws_instances" "test" {
  instance_tags = {
    Name = aws_instance.test[0].tags["Name"]
  }

  instance_state_names = ["pending", "running"]
  depends_on           = [aws_instance.test]
}
`, rName))
}
