package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_connect_instance", &resource.Sweeper{
		Name: "aws_connect_instance",
		F:    testSweepConnectInstance,
	})
}

func testSweepConnectInstance(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).connectconn
	ctx := context.Background()
	// MaxResults:  Maximum value of 10. https://docs.aws.amazon.com/connect/latest/APIReference/API_ListInstances.html
	input := &connect.ListInstancesInput{MaxResults: aws.Int64(10)}
	var sweeperErrs *multierror.Error
	for {
		listOutput, err := conn.ListInstances(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping Connect Instance sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Connect Instance: %s", err)
		}
		for _, instance := range listOutput.InstanceSummaryList {
			id := aws.StringValue(instance.Id)
			r := resourceAwsConnectInstance()
			d := r.Data(nil)
			d.SetId(id)

			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}
		if aws.StringValue(listOutput.NextToken) == "" {
			break
		}
		input.NextToken = listOutput.NextToken
	}
	return sweeperErrs.ErrorOrNil()
}
func TestAccAwsConnectInstance_basic(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeConnectManaged),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_custom_tts_voices", "false"),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.InstanceStatusActive),
					testAccMatchResourceAttrGlobalARN(resourceName, "service_role", "iam", regexp.MustCompile(`role/aws-service-role/connect.amazonaws.com/.+`)),
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

func TestAccAwsConnectInstance_custom(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeConnectManaged),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices", "false"),
					resource.TestCheckResourceAttr(resourceName, "use_custom_tts_voices", "true"),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.InstanceStatusActive),
					testAccMatchResourceAttrGlobalARN(resourceName, "service_role", "iam", regexp.MustCompile(`role/aws-service-role/connect.amazonaws.com/.+`)),
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

func TestAccAwsConnectInstance_directory(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeExistingDirectory),
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"directory_id"},
			},
		},
	})
}

func testAccCheckAwsConnectInstanceExists(resourceName string, function *connect.DescribeInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect instance not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect instance ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(rs.Primary.ID),
		}

		getFunction, err := conn.DescribeInstance(input)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func TestAccAwsConnectInstance_saml(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigSAML(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeSaml),
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
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
func testAccCheckAwsConnectInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_instance" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		instanceID := rs.Primary.ID

		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceID),
		}

		_, connectErr := conn.DescribeInstance(input)
		// Verify the error is what we want
		if connectErr != nil {
			if awsErr, ok := connectErr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return connectErr
		}
	}
	return nil
}

func testAccAwsConnectInstanceConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  identity_management_type = "CONNECT_MANAGED"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccAwsConnectInstanceConfigCustom(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  identity_management_type  = "CONNECT_MANAGED"
  instance_alias            = %[1]q
  inbound_calls_enabled     = false
  outbound_calls_enabled    = true
  early_media_enabled       = false
  contact_flow_logs_enabled = true
  contact_lens_enabled      = false
  auto_resolve_best_voices  = false
  use_custom_tts_voices     = true
}
`, rName)
}

func testAccAwsConnectInstanceConfigDirectory(rName string) string {
	return fmt.Sprintf(`
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
    Name = "terraform-testacc-directory-service-directory-tags"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}

resource "aws_connect_instance" "foo" {
  directory_id             = aws_directory_service_directory.test.id
  identity_management_type = "EXISTING_DIRECTORY"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true

  depends_on = [aws_directory_service_directory.test]
}
`, rName)
}

func testAccAwsConnectInstanceConfigSAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  identity_management_type = "SAML"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName)
}
