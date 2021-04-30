package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/waiter"
)

func TestAccAwsRamPrincipalAssociation_basic(t *testing.T) {
	var association ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamPrincipalAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamPrincipalAssociationExists(resourceName, &association),
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
	var association ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamPrincipalAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamPrincipalAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamPrincipalAssociationExists(resourceName, &association),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRamPrincipalAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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
			return fmt.Errorf("error parsing ID (%s): %w", rs.Primary.ID, err)
		}

		var association *ram.ResourceShareAssociation

		if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
			// AWS Account ID Principals need to be accepted to become ASSOCIATED
			association, err = finder.ResourceSharePrincipalAssociationByShareARNPrincipal(conn, resourceShareARN, principal)
		} else {
			association, err = waiter.ResourceSharePrincipalAssociated(conn, resourceShareARN, principal)
		}

		if err != nil {
			return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareARN, principal, err)
		}

		if association == nil {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) not found", resourceShareARN, principal)
		}

		if aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociating {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) status not associating or associated: %s", resourceShareARN, principal, aws.StringValue(association.Status))
		}

		*resourceShare = *association

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

		association, err := waiter.ResourceSharePrincipalDisassociated(conn, resourceShareARN, principal)

		if err != nil {
			return err
		}

		if association != nil && aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusDisassociated {
			return fmt.Errorf("RAM Resource Share (%s) Principal Association (%s) not disassociated: %s", resourceShareARN, principal, aws.StringValue(association.Status))
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
