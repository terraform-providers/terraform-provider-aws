package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSMainRouteTableAssociation_basic(t *testing.T) {
	var rta ec2.RouteTableAssociation
	resourceName := "aws_main_route_table_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMainRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMainRouteTableAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMainRouteTableAssociationExists(resourceName, &rta),
				),
			},
			{
				Config: testAccMainRouteTableAssociationConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMainRouteTableAssociationExists(resourceName, &rta),
				),
			},
		},
	})
}

func testAccCheckMainRouteTableAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_main_route_table_association" {
			continue
		}

		_, err := finder.MainRouteTableAssociationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Main route table association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckMainRouteTableAssociationExists(n string, v *ec2.RouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		association, err := finder.MainRouteTableAssociationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *association

		return nil
	}
}

func testAccMainRouteTableAssociationConfigBaseVPC(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"

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
`, rName)
}

func testAccMainRouteTableAssociationConfig(rName string) string {
	return composeConfig(testAccMainRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test.id
}
`, rName))
}

func testAccMainRouteTableAssociationConfigUpdated(rName string) string {
	return composeConfig(testAccMainRouteTableAssociationConfigBaseVPC(rName), fmt.Sprintf(`
# Need to keep the old route table around when we update the
# main_route_table_association, otherwise Terraform will try to destroy the
# route table too early, and will fail because it's still the main one
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test2" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test2.id
}
`, rName))
}
