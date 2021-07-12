package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSVpcEndpointSecurityGroupAssociation_basic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_security_group_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSecurityGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSecurityGroupAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName, &endpoint),
					testAccCheckVpcEndpointSecurityGroupAssociationNumAssociations(&endpoint, 2),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointSecurityGroupAssociation_disappears(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_security_group_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSecurityGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSecurityGroupAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName, &endpoint),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpointSecurityGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointSecurityGroupAssociation_multiple(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName0 := "aws_vpc_endpoint_security_group_association.test.0"
	resourceName1 := "aws_vpc_endpoint_security_group_association.test.1"
	resourceName2 := "aws_vpc_endpoint_security_group_association.test.2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSecurityGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSecurityGroupAssociationConfigMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName0, &endpoint),
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName1, &endpoint),
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName2, &endpoint),
					testAccCheckVpcEndpointSecurityGroupAssociationNumAssociations(&endpoint, 4),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointSecurityGroupAssociation_ReplaceDefaultAssociation(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_security_group_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSecurityGroupAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSecurityGroupAssociationConfigReplaceDefaultAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSecurityGroupAssociationExists(resourceName, &endpoint),
					testAccCheckVpcEndpointSecurityGroupAssociationNumAssociations(&endpoint, 1),
				),
			},
		},
	})
}

func testAccCheckVpcEndpointSecurityGroupAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_security_group_association" {
			continue
		}

		out, err := finder.VpcEndpointByID(conn, rs.Primary.Attributes["vpc_endpoint_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		// VPC Endpoint will always have 1 SG.
		if len(out.Groups) > 1 {
			return fmt.Errorf("VPC endpoint %s has security groups", aws.StringValue(out.VpcEndpointId))
		}
	}

	return nil
}

func testAccCheckVpcEndpointSecurityGroupAssociationExists(n string, vpce *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		out, err := finder.VpcEndpointByID(conn, rs.Primary.Attributes["vpc_endpoint_id"])

		if err != nil {
			return err
		}

		err = finder.VpcEndpointSecurityGroupAssociationExists(conn, rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["security_group_id"])

		if err != nil {
			return err
		}

		*vpce = *out

		return nil
	}
}

func testAccCheckVpcEndpointSecurityGroupAssociationNumAssociations(vpce *ec2.VpcEndpoint, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len := len(vpce.Groups); len != n {
			return fmt.Errorf("got %d associations; wanted %d", len, n)
		}

		return nil
	}
}

func testAccVpcEndpointSecurityGroupAssociationConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_security_group" "test" {
  count = 3

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVpcEndpointSecurityGroupAssociationConfigBasic(rName string) string {
	return composeConfig(
		testAccVpcEndpointSecurityGroupAssociationConfigBase(rName),
		`
resource "aws_vpc_endpoint_security_group_association" "test" {
  vpc_endpoint_id   = aws_vpc_endpoint.test.id
  security_group_id = aws_security_group.test[0].id
}
`)
}

func testAccVpcEndpointSecurityGroupAssociationConfigMultiple(rName string) string {
	return composeConfig(
		testAccVpcEndpointSecurityGroupAssociationConfigBase(rName),
		`
resource "aws_vpc_endpoint_security_group_association" "test" {
  count = length(aws_security_group.test)

  vpc_endpoint_id   = aws_vpc_endpoint.test.id
  security_group_id = aws_security_group.test[count.index].id
}
`)
}

func testAccVpcEndpointSecurityGroupAssociationConfigReplaceDefaultAssociation(rName string) string {
	return composeConfig(
		testAccVpcEndpointSecurityGroupAssociationConfigBase(rName),
		`
resource "aws_vpc_endpoint_security_group_association" "test" {
  vpc_endpoint_id   = aws_vpc_endpoint.test.id
  security_group_id = aws_security_group.test[0].id

  replace_default_association = true
}
`)
}
