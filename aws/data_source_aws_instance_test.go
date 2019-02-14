package aws

import (
	"testing"

	"fmt"
	"regexp"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSInstanceDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "ami", "ami-4fccb37f"),
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "instance_type", "m1.small"),
					resource.TestMatchResourceAttr("data.aws_instance.web-instance", "arn", regexp.MustCompile(`^arn:[^:]+:ec2:[^:]+:\d{12}:instance/i-.+`)),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_tags(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_Tags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "ami", "ami-4fccb37f"),
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "tags.%", "2"),
					resource.TestCheckResourceAttr("data.aws_instance.web-instance", "instance_type", "m1.small"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_AzUserData(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_AzUserData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-4fccb37f"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_gp2IopsDevice(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_gp2IopsDevice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-55a7ea65"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m3.medium"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.#", "1"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.0.iops", "100"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_blockDevices(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_blockDevices,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-55a7ea65"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m3.medium"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.#", "1"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.0.volume_size", "11"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.0.volume_type", "gp2"),
					resource.TestCheckResourceAttr("aws_instance.foo", "ebs_block_device.#", "3"),
					resource.TestCheckResourceAttr("aws_instance.foo", "ephemeral_block_device.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_rootInstanceStore(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_rootInstanceStore,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-44c36524"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m3.medium"),
					resource.TestCheckResourceAttr("aws_instance.foo", "ebs_block_device.#", "0"),
					resource.TestCheckResourceAttr("aws_instance.foo", "ebs_optimized", "false"),
					resource.TestCheckResourceAttr("aws_instance.foo", "root_block_device.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_privateIP(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_privateIP,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-c5eabbf5"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "private_ip", "10.1.1.42"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_keyPair(t *testing.T) {
	rName := fmt.Sprintf("tf-test-key-%d", acctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_keyPair(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-408c7f28"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "t1.micro"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "key_name", rName),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_VPC(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_VPC,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-4fccb37f"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "user_data", "562a3e32810edf6ff09994f050f12e799452379d"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "tenancy", "dedicated"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_PlacementGroup(t *testing.T) {
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_PlacementGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "placement_group", fmt.Sprintf("testAccInstanceDataSourceConfig_PlacementGroup_%s", rStr)),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_SecurityGroups(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_SecurityGroups(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-408c7f28"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "vpc_security_group_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_VPCSecurityGroups(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_VPCSecurityGroups,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "ami", "ami-21f78e11"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "t1.micro"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "security_groups.#", "0"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "vpc_security_group_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_getPasswordData_trueToFalse(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "get_password_data", "true"),
					resource.TestCheckResourceAttrSet("data.aws_instance.foo", "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "get_password_data", "false"),
					resource.TestCheckNoResourceAttr("data.aws_instance.foo", "password_data"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_getPasswordData_falseToTrue(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_getPasswordData(false, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "get_password_data", "false"),
					resource.TestCheckNoResourceAttr("data.aws_instance.foo", "password_data"),
				),
			},
			{
				Config: testAccInstanceDataSourceConfig_getPasswordData(true, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "get_password_data", "true"),
					resource.TestCheckResourceAttrSet("data.aws_instance.foo", "password_data"),
				),
			},
		},
	})
}

func TestAccAWSInstanceDataSource_creditSpecification(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{

				Config: testAccInstanceDataSourceConfig_creditSpecification,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_instance.foo", "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "credit_specification.#", "1"),
					resource.TestCheckResourceAttr("data.aws_instance.foo", "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

// Lookup based on InstanceID
const testAccInstanceDataSourceConfig = `
resource "aws_instance" "web" {
  # us-west-2
  ami = "ami-4fccb37f"
  instance_type = "m1.small"
  tags = {
    Name = "HelloWorld"
  }
}

data "aws_instance" "web-instance" {
  filter {
    name = "instance-id"
    values = ["${aws_instance.web.id}"]
  }
}
`

// Use the tags attribute to filter
func testAccInstanceDataSourceConfig_Tags(rInt int) string {
	return fmt.Sprintf(`
resource "aws_instance" "web" {
  # us-west-2
  ami = "ami-4fccb37f"
  instance_type = "m1.small"
  tags = {
    Name = "HelloWorld"
    TestSeed = "%d"
  }
}

data "aws_instance" "web-instance" {
  instance_tags = {
    Name = "${aws_instance.web.tags["Name"]}"
    TestSeed = "%d"
  }
}
`, rInt, rInt)
}

// filter on tag, populate more attributes
const testAccInstanceDataSourceConfig_AzUserData = `
resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-4fccb37f"
  availability_zone = "us-west-2a"

  instance_type = "m1.small"
  user_data = "foo:-with-character's"
  tags = {
    TFAccTest = "YesThisIsATest"
  }
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

// GP2IopsDevice
const testAccInstanceDataSourceConfig_gp2IopsDevice = `
resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-55a7ea65"
  instance_type = "m3.medium"
  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

// Block Device
const testAccInstanceDataSourceConfig_blockDevices = `
resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-55a7ea65"
  instance_type = "m3.medium"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }
  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops = 100
  }

  # Encrypted ebs block device
  ebs_block_device {
    device_name = "/dev/sdd"
    volume_size = 12
    encrypted = true
  }

  ephemeral_block_device {
    device_name = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

const testAccInstanceDataSourceConfig_rootInstanceStore = `
resource "aws_instance" "foo" {
  ami = "ami-44c36524"
  instance_type = "m3.medium"
}
data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

const testAccInstanceDataSourceConfig_privateIP = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-ds-private-ip"
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "tf-acc-instance-ds-private-ip"
  }
}

resource "aws_instance" "foo" {
  ami = "ami-c5eabbf5"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.foo.id}"
  private_ip = "10.1.1.42"
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

func testAccInstanceDataSourceConfig_keyPair(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_key_pair" "debugging" {
  key_name = "%s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_instance" "foo" {
  ami = "ami-408c7f28"
  instance_type = "t1.micro"
  key_name = "${aws_key_pair.debugging.key_name}"
  tags = {
    Name = "testAccInstanceDataSourceConfigKeyPair_TestAMI"
  }
}

data "aws_instance" "foo" {
  filter {
    name = "tag:Name"
    values = ["testAccInstanceDataSourceConfigKeyPair_TestAMI"]
  }
  filter {
    name = "key-name"
    values = ["${aws_instance.foo.key_name}"]
  }
}`, rName)
}

const testAccInstanceDataSourceConfig_VPC = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-data-source-vpc"
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
   Name = "tf-acc-instance-data-source-vpc"
  }
}

resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-4fccb37f"
  instance_type = "m1.small"
  subnet_id = "${aws_subnet.foo.id}"
  associate_public_ip_address = true
  tenancy = "dedicated"
  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`

func testAccInstanceDataSourceConfig_PlacementGroup(rStr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-data-source-placement-group"
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "tf-acc-instance-data-source-placement-group"
  }
}

resource "aws_placement_group" "foo" {
  name = "testAccInstanceDataSourceConfig_PlacementGroup_%s"
  strategy = "cluster"
}

# Limitations: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html#concepts-placement-groups
resource "aws_instance" "foo" {
  # us-west-2
  ami = "ami-55a7ea65"
  instance_type = "c3.large"
  subnet_id = "${aws_subnet.foo.id}"
  associate_public_ip_address = true
  placement_group = "${aws_placement_group.foo.name}"
  # pre-encoded base64 data
  user_data = "3dc39dda39be1205215e776bad998da361a5955d"
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`, rStr)
}

func testAccInstanceDataSourceConfig_SecurityGroups(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_security_group" "tf_test_foo" {
  name = "tf_test_foo-%d"
  description = "foo"

  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    self = true
  }
}

resource "aws_instance" "foo" {
  ami = "ami-408c7f28"
  instance_type = "m1.small"
  security_groups = ["${aws_security_group.tf_test_foo.name}"]
  user_data = "foo:-with-character's"
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`, rInt)
}

const testAccInstanceDataSourceConfig_VPCSecurityGroups = `
resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "terraform-testacc-instance-data-source-vpc-sgs"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-instance-data-source-vpc-sgs"
  }
}

resource "aws_security_group" "tf_test_foo" {
  name = "tf_test_foo"
  description = "foo"
  vpc_id="${aws_vpc.foo.id}"

  ingress {
    protocol = "icmp"
    from_port = -1
    to_port = -1
    self = true
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "tf-acc-instance-data-source-vpc-sgs"
  }
}

resource "aws_instance" "foo_instance" {
  ami = "ami-21f78e11"
  instance_type = "t1.micro"
  vpc_security_group_ids = ["${aws_security_group.tf_test_foo.id}"]
  subnet_id = "${aws_subnet.foo.id}"
  depends_on = ["aws_internet_gateway.gw"]
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo_instance.id}"
}
`

func testAccInstanceDataSourceConfig_getPasswordData(val bool, rInt int) string {
	return fmt.Sprintf(`
	# Find latest Microsoft Windows Server 2016 Core image (Amazon deletes old ones)
	data "aws_ami" "win2016core" {
		most_recent = true

		filter {
			name = "owner-alias"
			values = ["amazon"]
		}

		filter {
			name = "name"
			values = ["Windows_Server-2016-English-Core-Base-*"]
		}
	}

	resource "aws_key_pair" "foo" {
		key_name = "tf-acctest-%d"
		public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== tf-acc-winpasswordtest"
	}

	resource "aws_instance" "foo" {
		ami = "${data.aws_ami.win2016core.id}"
		instance_type = "t2.medium"
		key_name = "${aws_key_pair.foo.key_name}"
	}

	data "aws_instance" "foo" {
		instance_id = "${aws_instance.foo.id}"

		get_password_data = %t
	}
	`, rInt, val)
}

const testAccInstanceDataSourceConfig_creditSpecification = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_instance" "foo" {
  ami = "ami-bf4193c7"
  instance_type = "t2.micro"
  subnet_id = "${aws_subnet.foo.id}"
  credit_specification {
    cpu_credits = "unlimited"
  }
}

data "aws_instance" "foo" {
  instance_id = "${aws_instance.foo.id}"
}
`
