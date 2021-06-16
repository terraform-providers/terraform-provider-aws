package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsMacie2Account_basic(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func testAccAwsMacie2Account_FindingPublishingFrequency(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieAccountConfigWithFinding(macie2.FindingPublishingFrequencyFifteenMinutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAwsMacieAccountConfigWithFinding(macie2.FindingPublishingFrequencyOneHour),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func testAccAwsMacie2Account_WithStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieAccountConfigWithstatus(macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAwsMacieAccountConfigWithstatus(macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusPaused),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func testAccAwsMacie2Account_WithFindingAndStatus(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieAccountConfigWithfindingandstatus(macie2.FindingPublishingFrequencyFifteenMinutes, macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusEnabled),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAwsMacieAccountConfigWithfindingandstatus(macie2.FindingPublishingFrequencyOneHour, macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, "status", macie2.MacieStatusPaused),
					testAccCheckResourceAttrGlobalARN(resourceName, "service_role", "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					testAccCheckResourceAttrRfc3339(resourceName, "created_at"),
					testAccCheckResourceAttrRfc3339(resourceName, "updated_at"),
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

func testAccAwsMacie2Account_disappears(t *testing.T) {
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2AccountDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieAccountConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2AccountExists(resourceName, &macie2Output),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsMacie2AccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_account" {
			continue
		}

		input := &macie2.GetMacieSessionInput{}
		resp, err := conn.GetMacieSession(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie account %q still enabled", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsMacie2AccountExists(resourceName string, macie2Session *macie2.GetMacieSessionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetMacieSessionInput{}

		resp, err := conn.GetMacieSession(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie account %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccAwsMacieAccountConfigBasic() string {
	return `
resource "aws_macie2_account" "test" {}
`
}

func testAccAwsMacieAccountConfigWithFinding(finding string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
}
`, finding)
}

func testAccAwsMacieAccountConfigWithstatus(status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  status = "%s"
}
`, status)
}

func testAccAwsMacieAccountConfigWithfindingandstatus(finding, status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
  status                       = "%s"
}
`, finding, status)
}
