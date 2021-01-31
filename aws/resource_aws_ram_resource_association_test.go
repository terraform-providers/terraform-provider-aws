package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsRamResourceAssociation_basic(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_resource_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceAssociationExists(resourceName, &resourceShareAssociation1),
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

func TestAccAwsRamResourceAssociation_disappears(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_resource_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceAssociationExists(resourceName, &resourceShareAssociation1),
					testAccCheckAwsRamResourceAssociationDisappears(&resourceShareAssociation1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsRamResourceAssociationDisappears(resourceShareAssociation *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		input := &ram.DisassociateResourceShareInput{
			ResourceArns:     []*string{resourceShareAssociation.AssociatedEntity},
			ResourceShareArn: resourceShareAssociation.ResourceShareArn,
		}

		_, err := conn.DisassociateResourceShare(input)
		if err != nil {
			return err
		}

		return waitForRamResourceShareResourceDisassociation(conn, aws.StringValue(resourceShareAssociation.ResourceShareArn), aws.StringValue(resourceShareAssociation.AssociatedEntity))
	}
}

func testAccCheckAwsRamResourceAssociationExists(resourceName string, association *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		resourceShareARN, resourceARN, err := decodeRamResourceAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := getRamResourceShareAssociation(conn, resourceShareARN, resourceARN)

		if err != nil {
			return fmt.Errorf("error reading RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
		}

		if resourceShareAssociation == nil {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not found", resourceShareARN, resourceARN)
		}

		if aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not associated: %s", resourceShareARN, resourceARN, aws.StringValue(resourceShareAssociation.Status))
		}

		*association = *resourceShareAssociation

		return nil
	}
}

func testAccCheckAwsRamResourceAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_association" {
			continue
		}

		resourceShareARN, resourceARN, err := decodeRamResourceAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := getRamResourceShareAssociation(conn, resourceShareARN, resourceARN)

		if err != nil {
			return err
		}

		if resourceShareAssociation != nil && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusDisassociated {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not disassociated: %s", resourceShareARN, resourceARN, aws.StringValue(resourceShareAssociation.Status))
		}
	}

	return nil
}

func testAccAwsRamResourceAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ram-resource-association"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ram-resource-association"
  }
}

resource "aws_ram_resource_share" "test" {
  name = %q
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_subnet.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName)
}
