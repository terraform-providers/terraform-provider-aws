package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSESDomainDkim_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsSESDomainDkimConfig, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESDomainDkimExists("aws_ses_domain_dkim.test"),
					testAccCheckAwsSESDomainDkimTokens("aws_ses_domain_dkim.test", domain),
				),
			},
		},
	})
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
		conn := testAccProvider.Meta().(*AWSClient).sesConn

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

func testAccCheckAwsSESDomainDkimTokens(n string, domain string) resource.TestCheckFunc {
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
	domain = "${aws_ses_domain_identity.test.domain}"
}
`
