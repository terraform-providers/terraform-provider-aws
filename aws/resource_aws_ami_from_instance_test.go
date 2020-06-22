package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAMIFromInstance_basic(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_ami_from_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMIFromInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMIFromInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMIFromInstanceExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing Terraform aws_ami_from_instance resource"),
				),
			},
		},
	})
}

func TestAccAWSAMIFromInstance_tags(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_ami_from_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMIFromInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMIFromInstanceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMIFromInstanceExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSAMIFromInstanceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMIFromInstanceExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAMIFromInstanceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMIFromInstanceExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSAMIFromInstanceExists(resourceName string, image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if len(output.Images) == 0 || aws.StringValue(output.Images[0].ImageId) != rs.Primary.ID {
			return fmt.Errorf("AMI %q not found", rs.Primary.ID)
		}

		*image = *output.Images[0]

		return nil
	}
}

func testAccCheckAWSAMIFromInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami_from_instance" {
			continue
		}

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if output != nil && len(output.Images) > 0 && aws.StringValue(output.Images[0].ImageId) == rs.Primary.ID {
			return fmt.Errorf("AMI %q still exists in state: %s", rs.Primary.ID, aws.StringValue(output.Images[0].State))
		}
	}

	// Check for managed EBS snapshots
	return testAccCheckAWSEbsSnapshotDestroy(s)
}

func testAccAWSAMIFromInstanceConfigBase() string {
	return fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "available" {
  filter {
    name   = "instance-type"
    values = ["t3.micro", "t2.micro"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro"]
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
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = "testAccAWSAMIFromInstanceConfig_TestAMI"
  }
}
`)
}

func testAccAWSAMIFromInstanceConfig(rName string) string {
	return testAccAWSAMIFromInstanceConfigBase() + fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = "${aws_instance.test.id}"
}
`, rName)
}

func testAccAWSAMIFromInstanceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSAMIFromInstanceConfigBase() + fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = "${aws_instance.test.id}"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAMIFromInstanceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSAMIFromInstanceConfigBase() + fmt.Sprintf(`
resource "aws_ami_from_instance" "test" {
  name               = %[1]q
  description        = "Testing Terraform aws_ami_from_instance resource"
  source_instance_id = "${aws_instance.test.id}"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
