package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_licensemanager_license_configuration", &resource.Sweeper{
		Name: "aws_licensemanager_license_configuration",
		F:    testSweepLicenseManagerLicenseConfigurations,
	})
}

func testSweepLicenseManagerLicenseConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).licensemanagerconn

	resp, err := conn.ListLicenseConfigurations(&licensemanager.ListLicenseConfigurationsInput{})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping License Manager License Configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving License Manager license configurations: %s", err)
	}

	if len(resp.LicenseConfigurations) == 0 {
		log.Print("[DEBUG] No License Manager license configurations to sweep")
		return nil
	}

	for _, lc := range resp.LicenseConfigurations {
		id := aws.StringValue(lc.LicenseConfigurationArn)

		log.Printf("[INFO] Deleting License Manager license configuration: %s", id)

		opts := &licensemanager.DeleteLicenseConfigurationInput{
			LicenseConfigurationArn: aws.String(id),
		}

		_, err := conn.DeleteLicenseConfiguration(opts)

		if err != nil {
			log.Printf("[ERROR] Error deleting License Manager license configuration (%s): %s", id, err)
		}
	}

	return nil
}

func TestAccAWSLicenseManagerLicenseConfiguration_basic(t *testing.T) {
	var licenseConfiguration licensemanager.LicenseConfiguration
	resourceName := "aws_licensemanager_license_configuration.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLicenseManagerLicenseConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseManagerLicenseConfigurationExists(resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", "Example"),
					resource.TestCheckResourceAttr(resourceName, "description", "Example"),
					resource.TestCheckResourceAttr(resourceName, "license_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "true"),
					resource.TestCheckResourceAttr(resourceName, "license_counting_type", "Socket"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "license_rules.0", "#minimumSockets=3"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "barr"),
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

func TestAccAWSLicenseManagerLicenseConfiguration_update(t *testing.T) {
	var licenseConfiguration licensemanager.LicenseConfiguration
	resourceName := "aws_licensemanager_license_configuration.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLicenseManagerLicenseConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseManagerLicenseConfigurationExists(resourceName, &licenseConfiguration),
				),
			},
			{
				Config: testAccLicenseManagerLicenseConfigurationConfig_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLicenseManagerLicenseConfigurationExists(resourceName, &licenseConfiguration),
					resource.TestCheckResourceAttr(resourceName, "name", "NewName"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "license_count", "123"),
					resource.TestCheckResourceAttr(resourceName, "license_count_hard_limit", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.abc", "def"),
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

func testAccCheckLicenseManagerLicenseConfigurationExists(resourceName string, licenseConfiguration *licensemanager.LicenseConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).licensemanagerconn
		resp, err := conn.ListLicenseConfigurations(&licensemanager.ListLicenseConfigurationsInput{
			LicenseConfigurationArns: [](*string){aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error retrieving License Manager license configuration (%s): %s", rs.Primary.ID, err)
		}

		if len(resp.LicenseConfigurations) == 0 {
			return fmt.Errorf("Error retrieving License Manager license configuration (%s): Not found", rs.Primary.ID)
		}

		*licenseConfiguration = *resp.LicenseConfigurations[0]
		return nil
	}
}

func testAccCheckLicenseManagerLicenseConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).licensemanagerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_licensemanager_license_configuration" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetLicenseConfiguration(&licensemanager.GetLicenseConfigurationInput{
			LicenseConfigurationArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, licensemanager.ErrCodeInvalidParameterValueException, "") {
				continue
			}
			return err
		}
	}

	return nil

}

const testAccLicenseManagerLicenseConfigurationConfig_basic = `
resource "aws_licensemanager_license_configuration" "example" {
  name                     = "Example"
  description              = "Example"
  license_count            = 10
  license_count_hard_limit = true
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]

  tags = {
    foo = "barr"
  }
}
`

const testAccLicenseManagerLicenseConfigurationConfig_update = `
resource "aws_licensemanager_license_configuration" "example" {
  name                     = "NewName"
  license_count            = 123
  license_counting_type    = "Socket"

  license_rules = [
    "#minimumSockets=3"
  ]

  tags = {
    test = "test"
    abc = "def"
  }
}
`
