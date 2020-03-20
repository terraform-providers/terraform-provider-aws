package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_smb", &resource.Sweeper{
		Name: "aws_datasync_location_smb",
		F:    testSweepDataSyncLocationSmbs,
	})
}

func testSweepDataSyncLocationSmbs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location SMB sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location SMBs: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location SMBs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "smb://") {
				log.Printf("[INFO] Skipping DataSync Location SMB: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location SMB: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location SMB (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationSmb_basic(t *testing.T) {
	var locationSmb1 datasync.DescribeLocationSmbOutput

	resourceName := "aws_datasync_location_smb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationSmbDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationSmbConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationSmbExists(resourceName, &locationSmb1),

					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^smb://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationSmb_disappears(t *testing.T) {
	var locationSmb1 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationSmbDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationSmbConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationSmbExists(resourceName, &locationSmb1),
					testAccCheckAWSDataSyncLocationSmbDisappears(&locationSmb1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationSmb_Tags(t *testing.T) {
	var locationSmb1, locationSmb2, locationSmb3 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationSmbDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationSmbConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationSmbExists(resourceName, &locationSmb1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
			{
				Config: testAccAWSDataSyncLocationSmbConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationSmbExists(resourceName, &locationSmb2),
					testAccCheckAWSDataSyncLocationSmbNotRecreated(&locationSmb1, &locationSmb2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationSmbConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationSmbExists(resourceName, &locationSmb3),
					testAccCheckAWSDataSyncLocationSmbNotRecreated(&locationSmb2, &locationSmb3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationSmbDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_smb" {
			continue
		}

		input := &datasync.DescribeLocationSmbInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationSmb(input)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDataSyncLocationSmbExists(resourceName string, locationSmb *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
		input := &datasync.DescribeLocationSmbInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationSmb(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationSmb = *output

		return nil
	}
}

func testAccCheckAWSDataSyncLocationSmbDisappears(location *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckAWSDataSyncLocationSmbNotRecreated(i, j *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Location SMB was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncLocationSmbConfigBase() string {
	gatewayUid := acctest.RandString(5)

	return fmt.Sprintf(`
data "aws_ami" "aws_thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

data "aws_ami" "aws_datasync" {
  most_recent = true
	# I do not know why, but only in us-west-2 
	# does the datasync ami _not_ have the amazon-alias.
	# Reverting to amazon-owner id.
  owners      = ["633936118553"]

  filter {
    name   = "name"
    values = ["aws-datasync-*"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "storagegateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.name}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Effect": "Allow"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  force_destroy = true

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
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
    Name = "tf-acc-test-datasync-smb"
  }
}


resource "aws_instance" "test_gateway" {
	depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws_thinstaller.id}"
  associate_public_ip_address = true

  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = ["${aws_security_group.test.id}"]
  subnet_id              = "${aws_subnet.test.id}"

  ebs_block_device {
    device_name = "/dev/sdf"
    volume_size = "150"
  }

  tags = {
    Name = "tf-acc-test-datasync-smb"
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = "${aws_instance.test_gateway.public_ip}"
  gateway_name       = "datasyncsmb-%s"
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
  smb_guest_password = "ZaphodBeeblebroxPW"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "/dev/nvme1n1"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
	# (Shamelessly stolen from:
	# https://github.com/terraform-providers/terraform-provider-aws/
	# blob/1647a5ba13c5abaf5cf65ecdeb7c5fdf0107e56f/aws
	# /resource_aws_storagegateway_cache_test.go#L219 )
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.
  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = "${data.aws_storagegateway_local_disk.test.id}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = "${aws_storagegateway_gateway.test.arn}"
  location_arn   = "${aws_s3_bucket.test.arn}"
  role_arn       = "${aws_iam_role.test.arn}"

	# I'm not super sure why this depends_on sadness is required in
	# the test framework but not the binary so... yolo!
	depends_on = ["aws_storagegateway_cache.test"]
}

resource "aws_instance" "test_datasync" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws_datasync.id}"
  associate_public_ip_address = true

  instance_type          = "c5.large"
  vpc_security_group_ids = ["${aws_security_group.test.id}"]
  subnet_id              = "${aws_subnet.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-smb"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test_datasync.public_ip}"
  name       = "datasyncsmb-%s"
}
`, gatewayUid, gatewayUid)
}

func testAccAWSDataSyncLocationSmbConfig() string {
	return testAccAWSDataSyncLocationSmbConfigBase() + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
	user               = "Guest"
	password           = "ZaphodBeeblebroxPW"
	subdirectory       = "/${aws_s3_bucket.test.id}/"

	server_hostname  = "${aws_instance.test_datasync.public_ip}"
	agent_arns       = ["${aws_datasync_agent.test.arn}"]
}
`)
}

func testAccAWSDataSyncLocationSmbConfigTags1(key1, value1 string) string {
	return testAccAWSDataSyncLocationSmbConfigBase() + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
	user               = "Guest"
	password           = "ZaphodBeeblebroxPW"
	subdirectory       = "/${aws_s3_bucket.test.id}/"

	server_hostname  = "${aws_instance.test_datasync.public_ip}"
	agent_arns       = ["${aws_datasync_agent.test.arn}"]

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncLocationSmbConfigTags2(key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncLocationSmbConfigBase() + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
	user               = "Guest"
	password           = "ZaphodBeeblebroxPW"
	subdirectory       = "/${aws_s3_bucket.test.id}/"

	server_hostname  = "${aws_instance.test_datasync.public_ip}"
	agent_arns       = ["${aws_datasync_agent.test.arn}"]

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
