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

func TestAccAwsRamPrincipalAssociation_basic(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamPrincipalAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamPrincipalAssociationExists(resourceName, &resourceShareAssociation1),
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

func TestAccAwsRamPrincipalAssociation_disappears(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamPrincipalAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamPrincipalAssociationExists(resourceName, &resourceShareAssociation1),
					testAccCheckAwsRamPrincipalAssociationDisappears(&resourceShareAssociation1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsRamPrincipalAssociationDisappears(resourceShareAssociation *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		input := &ram.DisassociateResourceShareInput{
			Principals:       []*string{resourceShareAssociation.AssociatedEntity},
			ResourceShareArn: resourceShareAssociation.ResourceShareArn,
		}

		_, err := conn.DisassociateResourceShare(input)
		if err != nil {
			return err
		}

		return waitForRamResourceSharePrincipalDisassociation(conn, aws.StringValue(resourceShareAssociation.ResourceShareArn), aws.StringValue(resourceShareAssociation.AssociatedEntity))
	}
}

func testAccCheckAwsRamPrincipalAssociationExists(resourceName string, resourceShare *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resourceShareARN, principal, err := resourceAwsRamPrincipalAssociationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := getRamResourceSharePrincipalAssociation(conn, resourceShareARN, principal)

		if err != nil {
			return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareARN, principal, err)
		}

		if resourceShareAssociation == nil {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) not found", resourceShareARN, principal)
		}

		if aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociating {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) not associating or associated: %s", resourceShareARN, principal, aws.StringValue(resourceShareAssociation.Status))
		}

		*resourceShare = *resourceShareAssociation

		return nil
	}
}

func testAccCheckAwsRamPrincipalAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_principal_association" {
			continue
		}

		resourceShareARN, principal, err := decodeRamResourceAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := getRamResourceSharePrincipalAssociation(conn, resourceShareARN, principal)

		if err != nil {
			return err
		}

		if resourceShareAssociation != nil && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusDisassociated {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) not disassociated: %s", resourceShareARN, principal, aws.StringValue(resourceShareAssociation.Status))
		}
	}

	return nil
}

func testAccAwsRamPrincipalAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = true
  name                      = %[1]q
}

resource "aws_ram_principal_association" "test" {
  principal          = "111111111111"
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName)
}
