package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_vpc_endpoint", &resource.Sweeper{
		Name: "aws_vpc_endpoint",
		F:    testSweepEc2VpcEndpoints,
		Dependencies: []string{
			"aws_route_table",
		},
	})
}

func testSweepEc2VpcEndpoints(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeVpcEndpointsInput{}

	err = conn.DescribeVpcEndpointsPages(input, func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vpcEndpoint := range page.VpcEndpoints {
			if vpcEndpoint == nil {
				continue
			}

			if aws.StringValue(vpcEndpoint.State) != "available" {
				continue
			}

			id := aws.StringValue(vpcEndpoint.VpcEndpointId)

			log.Printf("[INFO] Deleting EC2 VPC Endpoint: %s", id)

			r := resourceAwsVpcEndpoint()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 VPC Endpoint (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 VPC Endpoint sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 VPC Endpoints: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSVpcEndpoint_gatewayBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckVpcEndpointPrefixListAvailable(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
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

func TestAccAWSVpcEndpoint_gatewayWithRouteTableAndPolicy(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	var routeTable ec2.RouteTable
	resourceName := "aws_vpc_endpoint.test"
	routeTableResourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckRouteTableExists(routeTableResourceName, &routeTable),
					testAccCheckVpcEndpointPrefixListAvailable(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
				),
			},
			{
				Config: testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicyModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckRouteTableExists(routeTableResourceName, &routeTable),
					testAccCheckVpcEndpointPrefixListAvailable(resourceName),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
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

func TestAccAWSVpcEndpoint_gatewayPolicy(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	// This policy checks the DiffSuppressFunc
	policy1 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadOnly",
      "Principal": "*",
      "Action": [
        "dynamodb:DescribeTable",
        "dynamodb:ListTables"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
`

	policy2 := `
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
`

	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfigGatewayPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointConfigGatewayPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_interfaceBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
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

func TestAccAWSVpcEndpoint_interfaceWithSubnetAndSecurityGroup(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceWithSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
				),
			},
			{
				Config: testAccVpcEndpointConfig_interfaceWithSubnetModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
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

func TestAccAWSVpcEndpoint_interfaceNonAWSServiceAcceptOnCreate(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccAWSVpcEndpoint_interfaceNonAWSServiceAcceptOnUpdate(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceNonAWSService(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "state", "pendingAcceptance"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
				),
			},
			{
				Config: testAccVpcEndpointConfig_interfaceNonAWSService(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr(resourceName, "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_managed", "false"),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint/vpce-.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccAWSVpcEndpoint_disappears(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpcEndpoint_tags(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
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
				Config: testAccVpcEndpointConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVpcEndpointConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_VpcEndpointType_GatewayLoadBalancer(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	vpcEndpointServiceResourceName := "aws_vpc_endpoint_service.test"
	resourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfigVpcEndpointTypeGatewayLoadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_type", vpcEndpointServiceResourceName, "service_type"),
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

func testAccCheckVpcEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint" {
			continue
		}

		_, err := finder.VpcEndpointByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckVpcEndpointExists(n string, endpoint *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		out, err := finder.VpcEndpointByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*endpoint = *out

		return nil
	}
}

func testAccCheckVpcEndpointPrefixListAvailable(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		prefixListID := rs.Primary.Attributes["prefix_list_id"]
		if prefixListID == "" {
			return fmt.Errorf("Prefix list ID not available")
		}
		if !strings.HasPrefix(prefixListID, "pl") {
			return fmt.Errorf("Prefix list ID does not appear to be a valid value: '%s'", prefixListID)
		}

		var (
			cidrBlockSize int
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(rs.Primary.Attributes["cidr_blocks.#"]); err != nil {
			return err
		}
		if cidrBlockSize < 1 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		return nil
	}
}

func testAccVpcEndpointConfig_gatewayBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}
`, rName)
}

func testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  route_table_ids = [
    aws_route_table.test.id,
  ]

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicyModified(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  route_table_ids = []

  policy = ""

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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}
`, rName)
}

func testAccVpcEndpointConfig_interfaceBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"

  security_group_ids = [
    data.aws_security_group.test.id,
  ]
}
`, rName)
}

func testAccVpcEndpointConfigGatewayPolicy(rName, policy string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  policy       = <<POLICY
%[2]s
POLICY
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, policy)
}

func testAccVpcEndpointConfig_vpcBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVpcEndpointConfig_interfaceWithSubnet(rName string) string {
	return composeConfig(
		testAccVpcEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  subnet_ids = [
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[0].id,
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVpcEndpointConfig_interfaceWithSubnetModified(rName string) string {
	return composeConfig(
		testAccVpcEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = true

  subnet_ids = [
    aws_subnet.test[2].id,
    aws_subnet.test[1].id,
    aws_subnet.test[0].id,
  ]

  security_group_ids = [
    aws_security_group.test[1].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVpcEndpointConfig_interfaceNonAWSService(rName string, autoAccept bool) string {
	return composeConfig(
		testAccVpcEndpointConfig_vpcBase(rName),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test[0].id,
    aws_subnet.test[1].id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false
  auto_accept         = %[2]t

  security_group_ids = [
    aws_security_group.test[0].id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept))
}

func testAccVpcEndpointConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVpcEndpointConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVpcEndpointConfigVpcEndpointTypeGatewayLoadBalancer(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_caller_identity.current.arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}
