package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSWorkLinkFleet_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "company_code"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_DisplayName(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigDisplayName(suffix, "display1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display1"),
				),
			},
			{
				Config: testAccAWSWorkLinkFleetConfigDisplayName(suffix, "display2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_OptimizeForEndUserLocation(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigOptimizeForEndUserLocation(suffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "false"),
				),
			},
			{
				Config: testAccAWSWorkLinkFleetConfigOptimizeForEndUserLocation(suffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_AuditStreamArn(t *testing.T) {
	rName := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigAuditStreamArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "audit_stream_arn", "aws_kinesis_stream.test_stream", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_Network(t *testing.T) {
	rName := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigNetwork(rName, "192.168.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSWorkLinkFleetConfigNetwork(rName, "10.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccAWSWorkLinkFleetConfig(rName),
				ExpectError: regexp.MustCompile(`Company Network Configuration cannot be removed`),
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_DeviceCaCertificate(t *testing.T) {
	rName := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"
	fName := "test-fixtures/worklink-device-ca-certificate.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigDeviceCaCertificate(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "device_ca_certificate", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWorkLinkFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_ca_certificate", ""),
				),
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_IdentityProvider(t *testing.T) {
	rName := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"
	fName := "test-fixtures/saml-metadata.xml"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfigIdentityProvider(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.0.type", "SAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccAWSWorkLinkFleetConfig(rName),
				ExpectError: regexp.MustCompile(`Identity Provider Configuration cannot be removed`),
			},
		},
	})
}

func TestAccAWSWorkLinkFleet_Disappears(t *testing.T) {
	rName := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkFleetExists(resourceName),
					testAccCheckAWSWorkLinkFleetDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWorkLinkFleetDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn

		input := &worklink.DeleteFleetInput{
			FleetArn: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeleteFleet(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    worklinkFleetStateRefresh(conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForState()

		return err
	}
}

func testAccCheckAWSWorkLinkFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).worklinkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_worklink_fleet" {
			continue
		}

		_, err := conn.DescribeFleetMetadata(
			&worklink.DescribeFleetMetadataInput{
				FleetArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Return nil if the Worklink Fleet is already destroyed
			if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return err
		}
		return fmt.Errorf("Worklink Fleet %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSWorkLinkFleetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn
		_, err := conn.DescribeFleetMetadata(&worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccPreCheckAWSWorkLink(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).worklinkconn

	input := &worklink.ListFleetsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListFleets(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSWorkLinkFleetConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"
}
`, r)
}

func testAccAWSWorkLinkFleetConfigDisplayName(r, displayName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name         = "tf-worklink-fleet-%s"
  display_name = "%s"
}
`, r, displayName)
}

func testAccAWSWorkLinkFleetConfigOptimizeForEndUserLocation(r string, b bool) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name                           = "tf-worklink-fleet-%s"
  optimize_for_end_user_location = %t
}
`, r, b)
}

func testAccAWSWorkLinkFleetConfigNetwork_Base(rName, cidrBlock string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "%s"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "test" {
  name        = "tf_test_foo"
  description = "foo"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol  = "icmp"
    from_port = -1
    to_port   = -1
    self      = true
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %q
  }
}
`, cidrBlock, rName, rName)
}

func testAccAWSWorkLinkFleetConfigNetwork(r, cidrBlock string) string {
	return fmt.Sprintf(`
%s

resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  network {
    vpc_id             = aws_vpc.test.id
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }
}
`, testAccAWSWorkLinkFleetConfigNetwork_Base(r, cidrBlock), r)
}

func testAccAWSWorkLinkFleetConfigAuditStreamArn(r string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
  name        = "AmazonWorkLink-%s_kinesis_test"
  shard_count = 1
}

resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  audit_stream_arn = aws_kinesis_stream.test_stream.arn
}
`, r, r)
}

func testAccAWSWorkLinkFleetConfigDeviceCaCertificate(r string, fName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  device_ca_certificate = file("%s")
}
`, r, fName)
}

func testAccAWSWorkLinkFleetConfigIdentityProvider(r string, fName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  identity_provider {
    type          = "SAML"
    saml_metadata = file("%s")
  }
}
`, r, fName)
}
