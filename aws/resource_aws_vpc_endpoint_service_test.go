package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpc_endpoint_service", &resource.Sweeper{
		Name: "aws_vpc_endpoint_service",
		F:    testSweepEc2VpcEndpointServices,
		Dependencies: []string{
			"aws_vpc_endpoint",
		},
	})
}

func testSweepEc2VpcEndpointServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}

	for {
		output, err := conn.DescribeVpcEndpointServiceConfigurations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Endpoint Service sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 VPC Endpoint Services: %s", err)
		}

		for _, serviceConfiguration := range output.ServiceConfigurations {
			if aws.StringValue(serviceConfiguration.ServiceState) == ec2.ServiceStateDeleted {
				continue
			}

			id := aws.StringValue(serviceConfiguration.ServiceId)
			input := &ec2.DeleteVpcEndpointServiceConfigurationsInput{
				ServiceIds: []*string{serviceConfiguration.ServiceId},
			}

			log.Printf("[INFO] Deleting EC2 VPC Endpoint Service: %s", id)
			_, err := conn.DeleteVpcEndpointServiceConfigurations(input)

			if isAWSErr(err, "InvalidVpcEndpointServiceId.NotFound", "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 VPC Endpoint Service (%s): %s", id, err)
			}

			if err := waitForVpcEndpointServiceDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for VPC Endpoint Service (%s) to delete: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSVpcEndpointService_basic(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	rName2 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSVpcEndpointService_AllowedPrincipalsAndTags(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	rName2 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_allowedPrincipalsAndTags(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointServiceConfig_allowedPrincipalsAndTagsUpdated(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName1),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointService_removed(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	rName2 := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

	testDestroy := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DeleteVpcEndpointServiceConfigurations(&ec2.DeleteVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{svcCfg.ServiceId},
		})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVpcEndpointServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_service" {
			continue
		}

		resp, err := conn.DescribeVpcEndpointServiceConfigurations(&ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if isAWSErr(err, "InvalidVpcEndpointServiceId.NotFound", "") {
				continue
			}
			return err
		}
		if len(resp.ServiceConfigurations) > 0 {
			return fmt.Errorf("VPC Endpoint Services still exist.")
		}

		return err
	}

	return nil
}

func testAccCheckVpcEndpointServiceExists(n string, svcCfg *ec2.ServiceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeVpcEndpointServiceConfigurations(&ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.ServiceConfigurations) == 0 {
			return fmt.Errorf("VPC Endpoint Service not found")
		}

		*svcCfg = *resp.ServiceConfigurations[0]

		return nil
	}
}

func testAccVpcEndpointServiceConfig_base(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test1" {
  name = %[1]q

  subnets = [
    "${aws_subnet.test1.id}",
    "${aws_subnet.test2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test2" {
  name = %[2]q

  subnets = [
    "${aws_subnet.test1.id}",
    "${aws_subnet.test2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[2]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "current" {}
`, rName1, rName2)
}

func testAccVpcEndpointServiceConfig_basic(rName1, rName2 string) string {
	return testAccVpcEndpointServiceConfig_base(rName1, rName2) + fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    "${aws_lb.test1.arn}",
  ]
}
`)
}

func testAccVpcEndpointServiceConfig_allowedPrincipalsAndTags(rName1, rName2 string) string {
	return testAccVpcEndpointServiceConfig_base(rName1, rName2) + fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    "${aws_lb.test1.arn}",
  ]

  allowed_principals = [
    "${data.aws_caller_identity.current.arn}",
  ]

  tags = {
    Environment = "test"
    Usage       = "original"
    Name        = %[1]q
  }
}
`, rName1)
}

func testAccVpcEndpointServiceConfig_allowedPrincipalsAndTagsUpdated(rName1, rName2 string) string {
	return testAccVpcEndpointServiceConfig_base(rName1, rName2) + fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    "${aws_lb.test1.arn}",
    "${aws_lb.test2.arn}",
  ]

  allowed_principals = []

  tags = {
    Usage = "changed"
    Name  = %[1]q
  }
}
`, rName1)
}
