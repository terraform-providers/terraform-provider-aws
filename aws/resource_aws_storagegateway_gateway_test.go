package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
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
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
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
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
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
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
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

func TestAccAWSStorageGatewayGateway_tags(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

func TestAccAWSStorageGatewayGateway_CloudWatchLogs(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"
	resourceName2 := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_Log_Group(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", resourceName2, "arn"),
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

func TestAccAWSStorageGatewayGateway_GatewayVpcEndpoint(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_vpc_endpoint", vpcEndpointResourceName, "dns_entry.0.dns_name"),
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
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

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
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSStorageGateway_FileGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
data "aws_ec2_instance_type_offering" "storagegateway" {
  filter {
    name   = "instance-type"
    values = ["m5.xlarge", "m4.xlarge"]
  }

  filter {
    name   = "location"
    values = [aws_subnet.test.availability_zone]
  }
  
  location_type            = "availability-zone"
  preferred_instance_types = ["m5.xlarge", "m4.xlarge"]
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = ["aws_route.test"]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.storagegateway.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName string) string {
	return testAccAWSStorageGateway_VPCBase(rName) + fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
data "aws_ec2_instance_type_offering" "storagegateway" {
  filter {
    name   = "instance-type"
    values = ["m5.xlarge", "m4.xlarge"]
  }

  filter {
    name   = "location"
    values = [aws_subnet.test.availability_zone]
  }
  
  location_type            = "availability-zone"
  preferred_instance_types = ["m5.xlarge", "m4.xlarge"]
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-common.html
# NOTE: CACHED, STORED, and VTL Gateway Types share the same AMI
data "aws_ssm_parameter" "aws_service_storagegateway_ami_CACHED_latest" {
  name = "/aws/service/storagegateway/ami/CACHED/latest"
}

resource "aws_instance" "test" {
  depends_on = ["aws_route.test"]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_CACHED_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.storagegateway.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
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

func testAccAWSStorageGatewayGatewayConfig_Log_Group(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address 		= "${aws_instance.test.public_ip}"
  gateway_name       		= %[1]q
  gateway_timezone   		= "GMT"
  gateway_type       		= "FILE_S3"
  cloudwatch_log_group_arn	= "${aws_cloudwatch_log_group.test.arn}"
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

func testAccAWSStorageGatewayGatewayConfig_GatewayVpcEndpoint(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
data "aws_vpc_endpoint_service" "storagegateway" {
  service = "storagegateway"
}

resource "aws_vpc_endpoint" "test" {
  security_group_ids = [aws_security_group.test.id]
  service_name       = data.aws_vpc_endpoint_service.storagegateway.service_name
  subnet_ids         = [aws_subnet.test.id]
  vpc_endpoint_type  = data.aws_vpc_endpoint_service.storagegateway.service_type
  vpc_id             = aws_vpc.test.id
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address   = aws_instance.test.public_ip
  gateway_name         = %[1]q
  gateway_timezone     = "GMT"
  gateway_type         = "CACHED"
  gateway_vpc_endpoint = aws_vpc_endpoint.test.dns_entry[0].dns_name
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName string) string {
	return fmt.Sprintf(`
# Directory Service Directories must be deployed across multiple EC2 Availability Zones
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

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
    Name = %[1]q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name         = aws_directory_service_directory.test.name
  domain_name_servers = aws_directory_service_directory.test.dns_ip_addresses

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  dhcp_options_id = aws_vpc_dhcp_options.test.id
  vpc_id          = aws_vpc.test.id
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
data "aws_ec2_instance_type_offering" "storagegateway" {
  filter {
    name   = "instance-type"
    values = ["m5.xlarge", "m4.xlarge"]
  }

  filter {
    name   = "location"
    values = [aws_subnet.test[0].availability_zone]
  }

  location_type            = "availability-zone"
  preferred_instance_types = ["m5.xlarge", "m4.xlarge"]
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test, aws_vpc_dhcp_options_association.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.storagegateway.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Administrator"
  }
}
`, rName)
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

func testAccAWSStorageGatewayGatewayConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
	%q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSStorageGatewayGatewayConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test.public_ip}"
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
	%q = %q
	%q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
