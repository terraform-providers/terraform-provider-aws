package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESReceiptFilter_basic(t *testing.T) {
	resourceName := "aws_ses_receipt_filter.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t); testAccPreCheckSESReceiptRule(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptFilterExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-filter/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cidr", "10.10.10.10"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy", "Block"),
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

func TestAccAWSSESReceiptFilter_disappears(t *testing.T) {
	resourceName := "aws_ses_receipt_filter.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t); testAccPreCheckSESReceiptRule(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptFilterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptFilterExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesReceiptFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESReceiptFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_filter" {
			continue
		}

		response, err := conn.ListReceiptFilters(&ses.ListReceiptFiltersInput{})
		if err != nil {
			return err
		}

		for _, element := range response.Filters {
			if aws.StringValue(element.Name) == rs.Primary.ID {
				return fmt.Errorf("SES Receipt Filter (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil

}

func testAccCheckAwsSESReceiptFilterExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES receipt filter not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES receipt filter ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		response, err := conn.ListReceiptFilters(&ses.ListReceiptFiltersInput{})
		if err != nil {
			return err
		}

		for _, element := range response.Filters {
			if aws.StringValue(element.Name) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("The receipt filter was not created")
	}
}

func testAccAWSSESReceiptFilterConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_filter" "test" {
  cidr   = "10.10.10.10"
  name   = %q
  policy = "Block"
}
`, rName)
}
