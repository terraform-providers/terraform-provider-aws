package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLicenseManagerAssociation_basic(t *testing.T) {
	var licenseSpecification licensemanager.LicenseSpecification
	resourceName := "aws_licensemanager_association.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLicenseManagerAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerAssociationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseManagerAssociationExists(resourceName, &licenseSpecification),
					resource.TestCheckResourceAttrPair(resourceName, "license_configuration_arn", "aws_licensemanager_license_configuration.example", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_instance.example", "arn"),
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

func testAccCheckLicenseManagerAssociationExists(resourceName string, licenseSpecification *licensemanager.LicenseSpecification) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).licensemanagerconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resourceArn, licenseConfigurationArn, err := resourceAwsLicenseManagerAssociationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		specification, err := resourceAwsLicenseManagerAssociationFindSpecification(conn, resourceArn, licenseConfigurationArn)
		if err != nil {
			return err
		}

		if specification == nil {
			return fmt.Errorf("Error retrieving License Manager association (%s): Not found", rs.Primary.ID)
		}

		*licenseSpecification = *specification
		return nil
	}
}

func testAccCheckLicenseManagerAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).licensemanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_licensemanager_association" {
			continue
		}

		resourceArn, licenseConfigurationArn, err := resourceAwsLicenseManagerAssociationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		specification, err := resourceAwsLicenseManagerAssociationFindSpecification(conn, resourceArn, licenseConfigurationArn)
		if err != nil {
			return err
		}

		if specification != nil {
			return fmt.Errorf("License Manager association %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccLicenseManagerAssociationConfig_basic = `
data "aws_ami" "example" {
  most_recent      = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }
}

resource "aws_instance" "example" {
  ami           = "${data.aws_ami.example.id}"
  instance_type = "t2.micro"
}

resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "vCPU"
}

resource "aws_licensemanager_association" "example" {
  license_configuration_arn = "${aws_licensemanager_license_configuration.example.id}"
  resource_arn              = "${aws_instance.example.arn}"
}
`
