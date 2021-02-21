package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{}

	err = conn.DescribeVpcEndpointServiceConfigurationsPages(input, func(page *ec2.DescribeVpcEndpointServiceConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, serviceConfiguration := range page.ServiceConfigurations {
			if serviceConfiguration == nil {
				continue
			}

			if aws.StringValue(serviceConfiguration.ServiceState) == ec2.ServiceStateDeleted {
				continue
			}

			id := aws.StringValue(serviceConfiguration.ServiceId)

			log.Printf("[INFO] Deleting EC2 VPC Endpoint Service: %s", id)

			r := resourceAwsVpcEndpointService()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 VPC Endpoint Service (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Endpoint Service sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 VPC Endpoint Services: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSVpcEndpointService_basic(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_NetworkLoadBalancerArns(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
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

func TestAccAWSVpcEndpointService_AllowedPrincipals(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_allowedPrincipals(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointServiceConfig_allowedPrincipalsUpdated(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_load_balancer_arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "allowed_principals.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName1),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointService_disappears(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_NetworkLoadBalancerArns(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpointService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointService_GatewayLoadBalancerArns(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName := acctest.RandomWithPrefix("tfacctest") // 32 character limit

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfig_GatewayLoadBalancerArns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "gateway_load_balancer_arns.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointServiceConfig_GatewayLoadBalancerArns(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "gateway_load_balancer_arns.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointService_tags(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfigTags1(rName1, rName2, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointServiceConfigTags2(rName1, rName2, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVpcEndpointServiceConfigTags1(rName1, rName2, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointService_private_dns_name(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	resourceName := "aws_vpc_endpoint_service.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceConfigPrivateDnsName(rName1, rName2, "example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.0.type", "TXT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointServiceConfigPrivateDnsName(rName1, rName2, "changed.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists(resourceName, &svcCfg),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name", "changed.com"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_name_configuration.0.type", "TXT"),
				),
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
    aws_subnet.test1.id,
    aws_subnet.test2.id,
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
    aws_subnet.test1.id,
    aws_subnet.test2.id,
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
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "current" {}
`, rName1, rName2)
}

func testAccVpcEndpointServiceConfig_GatewayLoadBalancerArns(rName string, count int) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-load-balancer"
  }
}

resource "aws_lb" "test" {
  count = %[2]d

  load_balancer_type = "gateway"
  name               = "%[1]s-${count.index}"

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  gateway_load_balancer_arns = aws_lb.test[*].arn
}
`, rName, count))
}

func testAccVpcEndpointServiceConfig_NetworkLoadBalancerArns(rName1, rName2 string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.test1.arn,
  ]
}
`)
}

func testAccVpcEndpointServiceConfig_allowedPrincipals(rName1, rName2 string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.test1.arn,
  ]

  allowed_principals = [
    data.aws_caller_identity.current.arn,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName1))
}

func testAccVpcEndpointServiceConfig_allowedPrincipalsUpdated(rName1, rName2 string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test1.arn,
    aws_lb.test2.arn,
  ]

  allowed_principals = []

  tags = {
    Name = %[1]q
  }
}
`, rName1))
}

func testAccVpcEndpointServiceConfigTags1(rName1, rName2, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.test1.arn,
  ]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVpcEndpointServiceConfigTags2(rName1, rName2, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false

  network_load_balancer_arns = [
    aws_lb.test1.arn,
  ]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVpcEndpointServiceConfigPrivateDnsName(rName1, rName2, dnsName string) string {
	return composeConfig(
		testAccVpcEndpointServiceConfig_base(rName1, rName2),
		fmt.Sprintf(`
resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = false
  private_dns_name    = "%s"

  network_load_balancer_arns = [
    aws_lb.test1.arn,
  ]
}
`, dnsName))
}
