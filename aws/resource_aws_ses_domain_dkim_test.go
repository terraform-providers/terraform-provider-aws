package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESDomainDkim_basic(t *testing.T) {
	resourceName := "aws_ses_domain_dkim.test"
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESDomainDkimDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsSESDomainDkimConfig, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainDkimExists(resourceName),
					testAccCheckAwsSESDomainDkimTokens(resourceName),
				),
			},
		},
	})
}

func testAccCheckAwsSESDomainDkimDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_domain_dkim" {
			continue
		}

		domain := rs.Primary.ID
		params := &ses.GetIdentityDkimAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		resp, err := conn.GetIdentityDkimAttributes(params)

		if err != nil {
			return err
		}

		if resp.DkimAttributes[domain] != nil {
			return fmt.Errorf("SES Domain Dkim %s still exists.", domain)
		}
	}

	return nil
}

func testAccCheckAwsSESDomainDkimExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		domain := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.GetIdentityDkimAttributesInput{
			Identities: []*string{
				aws.String(domain),
			},
		}

		response, err := conn.GetIdentityDkimAttributes(params)
		if err != nil {
			return err
		}

		if response.DkimAttributes[domain] == nil {
			return fmt.Errorf("SES Domain DKIM %s not found in AWS", domain)
		}

		return nil
	}
}

func testAccCheckAwsSESDomainDkimTokens(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		expectedNum := 3
		expectedFormat := regexp.MustCompile("[a-z0-9]{32}")

		tokenNum, _ := strconv.Atoi(rs.Primary.Attributes["dkim_tokens.#"])
		if expectedNum != tokenNum {
			return fmt.Errorf("Incorrect number of DKIM tokens, expected: %d, got: %d", expectedNum, tokenNum)
		}
		for i := 0; i < expectedNum; i++ {
			key := fmt.Sprintf("dkim_tokens.%d", i)
			token := rs.Primary.Attributes[key]
			if !expectedFormat.MatchString(token) {
				return fmt.Errorf("Incorrect format of DKIM token: %v", token)
			}
		}

		return nil
	}
}

const testAccAwsSESDomainDkimConfig = `
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_domain_dkim" "test" {
  domain = aws_ses_domain_identity.test.domain
}
`
