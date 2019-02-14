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
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_nfs", &resource.Sweeper{
		Name: "aws_datasync_location_nfs",
		F:    testSweepDataSyncLocationNfss,
	})
}

func testSweepDataSyncLocationNfss(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location EFS sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location EFSs: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location EFSs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "nfs://") {
				log.Printf("[INFO] Skipping DataSync Location EFS: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location EFS: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location EFS (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationNfs_basic(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationNfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationNfsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^nfs://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationNfs_disappears(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationNfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationNfsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs1),
					testAccCheckAWSDataSyncLocationNfsDisappears(&locationNfs1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationNfs_AgentARNs_Multple(t *testing.T) {
	var locationNfs1 datasync.DescribeLocationNfsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationNfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationNfsConfigAgentArnsMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "on_prem_config.0.agent_arns.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationNfs_Subdirectory(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var locationNfs1 datasync.DescribeLocationNfsOutput
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationNfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationNfsConfigSubdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationNfs_Tags(t *testing.T) {
	var locationNfs1, locationNfs2, locationNfs3 datasync.DescribeLocationNfsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_nfs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationNfsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationNfsConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"server_hostname"},
			},
			{
				Config: testAccAWSDataSyncLocationNfsConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs2),
					testAccCheckAWSDataSyncLocationNfsNotRecreated(&locationNfs1, &locationNfs2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationNfsConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationNfsExists(resourceName, &locationNfs3),
					testAccCheckAWSDataSyncLocationNfsNotRecreated(&locationNfs2, &locationNfs3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationNfsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_nfs" {
			continue
		}

		input := &datasync.DescribeLocationNfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationNfs(input)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDataSyncLocationNfsExists(resourceName string, locationNfs *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
		input := &datasync.DescribeLocationNfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationNfs(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationNfs = *output

		return nil
	}
}

func testAccCheckAWSDataSyncLocationNfsDisappears(location *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckAWSDataSyncLocationNfsNotRecreated(i, j *datasync.DescribeLocationNfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Location EFS was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncLocationNfsConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
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
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  # Default instance type from sync.sh
  instance_type               = "c5.2xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"
  name       = %q
}
`, rName)
}

func testAccAWSDataSyncLocationNfsConfig(rName string) string {
	return testAccAWSDataSyncLocationNfsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.test.arn}"]
  }
}
`)
}

func testAccAWSDataSyncLocationNfsConfigAgentArnsMultiple(rName string) string {
	return testAccAWSDataSyncLocationNfsConfigBase(rName) + fmt.Sprintf(`
resource "aws_instance" "test2" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  # Default instance type from sync.sh
  instance_type               = "c5.2xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-nfs"
  }
}

resource "aws_datasync_agent" "test2" {
  ip_address = "${aws_instance.test2.public_ip}"
  name       = "%s2"
}

resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [
      "${aws_datasync_agent.test.arn}",
      "${aws_datasync_agent.test2.arn}",
    ]
  }
}
`, rName)
}

func testAccAWSDataSyncLocationNfsConfigSubdirectory(rName, subdirectory string) string {
	return testAccAWSDataSyncLocationNfsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = %q

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.test.arn}"]
  }
}
`, subdirectory)
}

func testAccAWSDataSyncLocationNfsConfigTags1(rName, key1, value1 string) string {
	return testAccAWSDataSyncLocationNfsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.test.arn}"]
  }

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncLocationNfsConfigTags2(rName, key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncLocationNfsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_nfs" "test" {
  server_hostname = "example.com"
  subdirectory    = "/"

  on_prem_config {
    agent_arns = ["${aws_datasync_agent.test.arn}"]
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
