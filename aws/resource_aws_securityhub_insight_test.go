package aws

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
)

func testAccAwsSecurityHubInsight_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					testAccCheckAwsSecurityHubInsightArn(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison": securityhub.StringFilterComparisonEquals,
						"value":      "1234567890",
					}),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "AwsAccountId"),
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

func testAccAwsSecurityHubInsight_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSecurityHubInsight(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsSecurityHubInsight_DateFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	endDate := time.Now().Add(5 * time.Minute).Format(time.RFC1123)
	startDate := time.Now().Format(time.RFC1123)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_DateFilters_DateRange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.created_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.created_at.*", map[string]string{
						"date_range.#":       "1",
						"date_range.0.unit":  securityhub.DateRangeUnitDays,
						"date_range.0.value": "5",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsSecurityHubInsightConfig_DateFilters_StartEnd(rName, startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.created_at.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.created_at.*", map[string]string{
						"start": startDate,
						"end":   endDate,
					}),
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

func testAccAwsSecurityHubInsight_IpFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_IpFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.network_destination_ipv4.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.network_destination_ipv4.*", map[string]string{
						"cidr": "10.0.0.0/16",
					}),
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

func testAccAwsSecurityHubInsight_KeywordFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_KeywordFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.keyword.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.keyword.*", map[string]string{
						"value": rName,
					}),
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

func testAccAwsSecurityHubInsight_MapFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_MapFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.product_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison": securityhub.MapFilterComparisonEquals,
						"key":        "key1",
						"value":      "value1",
					}),
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

func testAccAwsSecurityHubInsight_MultipleFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_MultipleFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison": securityhub.StringFilterComparisonEquals,
						"value":      "1234567890",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison": securityhub.StringFilterComparisonEquals,
						"value":      "09876543210",
					}),
					resource.TestCheckResourceAttr(resourceName, "filters.0.product_fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison": securityhub.MapFilterComparisonEquals,
						"key":        "key1",
						"value":      "value1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.product_fields.*", map[string]string{
						"comparison": securityhub.MapFilterComparisonEquals,
						"key":        "key2",
						"value":      "value2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsSecurityHubInsightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.aws_account_id.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.aws_account_id.*", map[string]string{
						"comparison": securityhub.StringFilterComparisonEquals,
						"value":      "1234567890",
					}),
				),
			},
		},
	})
}

func testAccAwsSecurityHubInsight_Name(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-update")

	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					testAccCheckAwsSecurityHubInsightArn(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				Config: testAccAwsSecurityHubInsightConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					testAccCheckAwsSecurityHubInsightArn(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
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

func testAccAwsSecurityHubInsight_NumberFilters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_NumberFilters(rName, "eq = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"eq": "50.5",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsSecurityHubInsightConfig_NumberFilters(rName, "gte = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"gte": "50.5",
					}),
				),
			},
			{
				Config: testAccAwsSecurityHubInsightConfig_NumberFilters(rName, "lte = 50.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.confidence.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.confidence.*", map[string]string{
						"lte": "50.5",
					}),
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

func testAccAwsSecurityHubInsight_GroupByAttribute(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "AwsAccountId"),
				),
			},
			{
				Config: testAccAwsSecurityHubInsightConfig_UpdateGroupByAttribute(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "CompanyName"),
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

func testAccAwsSecurityHubInsight_WorkflowStatus(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_securityhub_insight.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubInsightDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubInsightConfig_WorkflowStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubInsightExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filters.0.workflow_status.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "filters.0.workflow_status.*", map[string]string{
						"comparison": securityhub.StringFilterComparisonEquals,
						"value":      securityhub.WorkflowStatusNew,
					}),
					resource.TestCheckResourceAttr(resourceName, "group_by_attribute", "WorkflowStatus"),
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

func testAccCheckAwsSecurityHubInsightDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_insight" {
			continue
		}

		insight, err := finder.Insight(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
				continue
			}
			if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
				continue
			}
			return fmt.Errorf("error deleting Security Hub Insight (%s): %w", rs.Primary.ID, err)
		}

		if insight != nil {
			return fmt.Errorf("Security Hub Insight (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsSecurityHubInsightExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		insight, err := finder.Insight(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Security Hub Insight (%s): %w", rs.Primary.ID, err)
		}

		if insight == nil {
			return fmt.Errorf("error reading Security Hub Insight (%s): not found", rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckAwsSecurityHubInsightArn checks the computed ARN value
// and accounts for differences in SecurityHub on GovCloud where the partition portion
// of the ARN is still "aws" while other services utilize the "aws-us-gov" partition
func testAccCheckAwsSecurityHubInsightArn(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expectedArn := fmt.Sprintf(`^arn:aws[^:]*:securityhub:%s:%s:insight/%s/custom/.+$`, testAccGetRegion(), testAccGetAccountID(), testAccGetAccountID())
		//lintignore:AWSAT001
		return resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(expectedArn))(s)
	}
}

func testAccAwsSecurityHubInsightConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_DateFilters_DateRange(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    created_at {
      date_range {
        unit  = "DAYS"
        value = 5
      }
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_DateFilters_StartEnd(rName, startDate, endDate string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    created_at {
      start = %q
      end   = %q
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, startDate, endDate, rName)
}

func testAccAwsSecurityHubInsightConfig_IpFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    network_destination_ipv4 {
      cidr = "10.0.0.0/16"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_KeywordFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    keyword {
      value = %[1]q
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_MapFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    product_fields {
      comparison = "EQUALS"
      key        = "key1"
      value      = "value1"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_MultipleFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }

    aws_account_id {
      comparison = "EQUALS"
      value      = "09876543210"
    }

    product_fields {
      comparison = "EQUALS"
      key        = "key1"
      value      = "value1"
    }

    product_fields {
      comparison = "EQUALS"
      key        = "key2"
      value      = "value2"
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %[1]q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_NumberFilters(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    confidence {
      %s
    }
  }

  group_by_attribute = "AwsAccountId"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, value, rName)
}

func testAccAwsSecurityHubInsightConfig_UpdateGroupByAttribute(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    aws_account_id {
      comparison = "EQUALS"
      value      = "1234567890"
    }
  }

  group_by_attribute = "CompanyName"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}

func testAccAwsSecurityHubInsightConfig_WorkflowStatus(rName string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_insight" "test" {
  filters {
    workflow_status {
      comparison = "EQUALS"
      value      = "NEW"
    }
  }

  group_by_attribute = "WorkflowStatus"

  name = %q

  depends_on = [aws_securityhub_account.test]
}
`, rName)
}
