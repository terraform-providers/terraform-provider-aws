package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_storagegateway_gateway", &resource.Sweeper{
		Name: "aws_storagegateway_gateway",
		F:    testSweepStorageGatewayGateways,
	})
}

func testSweepStorageGatewayGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).storagegatewayconn

	err = conn.ListGatewaysPages(&storagegateway.ListGatewaysInput{}, func(page *storagegateway.ListGatewaysOutput, isLast bool) bool {
		if len(page.Gateways) == 0 {
			log.Print("[DEBUG] No Storage Gateway Gateways to sweep")
			return true
		}

		for _, gateway := range page.Gateways {
			name := aws.StringValue(gateway.GatewayName)
			if !strings.HasPrefix(name, "tf-acc-test-") {
				log.Printf("[INFO] Skipping Storage Gateway Gateway: %s", name)
				continue
			}
			log.Printf("[INFO] Deleting Storage Gateway Gateway: %s", name)
			input := &storagegateway.DeleteGatewayInput{
				GatewayARN: gateway.GatewayARN,
			}

			_, err := conn.DeleteGateway(input)
			if err != nil {
				if isAWSErr(err, storagegateway.ErrorCodeGatewayNotFound, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Storage Gateway Gateway (%s): %s", name, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Storage Gateway Gateways: %s", err)
	}
	return nil
}

func TestAccAWSStorageGatewayGateway_GatewayType_Cached(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "CACHED"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_FileS3(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "FILE_S3"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_Stored(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "STORED"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_Vtl(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:storagegateway:[^:]+:[^:]+:gateway/sgw-.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "VTL"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayName(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName1),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayTimezone(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, "GMT-1:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-1:00"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, "GMT-2:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-2:00"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SmbActiveDirectorySettings(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", "terraformtesting.com"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SmbGuestPassword(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "myguestpassword1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword1"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "myguestpassword2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_guest_password"},
			},
		},
	})
}

func testAccCheckAWSStorageGatewayGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_gateway" {
			continue
		}

		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeGatewayInformation(input)

		if err != nil {
			if isAWSErrStorageGatewayGatewayNotFound(err) {
				return nil
			}
			return err
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewayGatewayExists(resourceName string, gateway *storagegateway.DescribeGatewayInformationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeGatewayInformation(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Gateway %q does not exist", rs.Primary.ID)
		}

		*gateway = *output

		return nil
	}
}

// testAccAWSStorageGateway_VPCBase provides a publicly accessible subnet
// and security group, suitable for Storage Gateway EC2 instances of any type
func testAccAWSStorageGateway_VPCBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = %q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %q
  }
}
`, rName, rName, rName, rName, rName)
}

// testAccAWSStorageGateway_FileGatewayBase uses the "thinstaller" Storage
// Gateway AMI for File Gateways
func testAccAWSStorageGateway_FileGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  # https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
  instance_type               = "m4.xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags = {
    Name = %q
  }
}
`, rName)
}

// testAccAWSStorageGateway_TapeAndVolumeGatewayBase uses the Storage Gateway
// AMI for either Tape or Volume Gateways
func testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
data "aws_ami" "aws-storage-gateway-2" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["aws-storage-gateway-2.*"]
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-storage-gateway-2.id}"
  associate_public_ip_address = true
  # https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
  instance_type               = "m4.xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags = {
    Name = %q
  }
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "STORED"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "VTL"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, gatewayTimezone string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = %q
  gateway_type       = "FILE_S3"
}
`, rName, gatewayTimezone)
}

func testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = %q
  }
}

resource "aws_route_table_association" "test" {
  count = 2

  subnet_id      = "${aws_subnet.test.*.id[count.index]}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    subnet_ids = ["${aws_subnet.test.*.id}"]
    vpc_id     = "${aws_vpc.test.id}"
  }

  tags = {
    Name = %q
  }
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name         = "${aws_directory_service_directory.test.name}"
  domain_name_servers = ["${aws_directory_service_directory.test.dns_ip_addresses}"]

  tags = {
    Name = %q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  dhcp_options_id = "${aws_vpc_dhcp_options.test.id}"
  vpc_id          = "${aws_vpc.test.id}"
}

data "aws_ami" "aws-thinstaller" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test", "aws_vpc_dhcp_options_association.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  # https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
  instance_type               = "m4.xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.*.id[0]}"

  tags = {
    Name = %q
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name = "${aws_directory_service_directory.test.name}"
    password    = "${aws_directory_service_directory.test.password}"
    username    = "Administrator"
  }
}
`, rName, rName, rName, rName, rName, rName, rName, rName, rName)
}

func testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, smbGuestPassword string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
  smb_guest_password = %q
}
`, rName, smbGuestPassword)
}
