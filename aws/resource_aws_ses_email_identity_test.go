package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ses_email_identity", &resource.Sweeper{
		Name: "aws_ses_email_identity",
		F:    func(region string) error { return testSweepSesIdentities(region, ses.IdentityTypeEmailAddress) },
	})
}

func TestAccAWSSESEmailIdentity_basic(t *testing.T) {
	email := fmt.Sprintf(
		"%s@terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", email))),
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

func TestAccAWSSESEmailIdentity_trailingPeriod(t *testing.T) {
	email := fmt.Sprintf(
		"%s@terraformtesting.com.",
		acctest.RandString(10))
	resourceName := "aws_ses_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESEmailIdentityConfig(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ses", regexp.MustCompile(fmt.Sprintf("identity/%s$", strings.TrimSuffix(email, ".")))),
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

func testAccCheckAwsSESEmailIdentityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_email_identity" {
			continue
		}

		email := rs.Primary.ID
		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[email] != nil {
			return fmt.Errorf("SES Email Identity %s still exists. Failing!", email)
		}
	}

	return nil
}

func testAccCheckAwsSESEmailIdentityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Email Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Email Identity name not set")
		}

		email := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.GetIdentityVerificationAttributesInput{
			Identities: []*string{
				aws.String(email),
			},
		}

		response, err := conn.GetIdentityVerificationAttributes(params)
		if err != nil {
			return err
		}

		if response.VerificationAttributes[email] == nil {
			return fmt.Errorf("SES Email Identity %s not found in AWS", email)
		}

		return nil
	}
}

func testAccAwsSESEmailIdentityConfig(email string) string {
	return fmt.Sprintf(`
resource "aws_ses_email_identity" "test" {
  email = %q
}
`, email)
}
