package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func testAccAwsMacie2FindingsFilter_basic(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
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

func testAccAwsMacie2FindingsFilter_Name_Generated(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func testAccAwsMacie2FindingsFilter_NamePrefix(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	namePrefix := "tf-acc-test-prefix-"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNamePrefix(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
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

func testAccAwsMacie2FindingsFilter_disappears(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigNameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMacie2Account(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsMacie2FindingsFilter_complete(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(descriptionUpdated, macie2.FindingsFilterActionNoop, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
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

func testAccAwsMacie2FindingsFilter_WithDate(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigCompleteMultipleCriterion(descriptionUpdated, macie2.FindingsFilterActionNoop, startDate, endDate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.0.field", "region"),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "updatedAt",
						"gte":   startDate,
						"lt":    endDate,
					}),
					testAccCheckResourceAttrRfc3339(resourceName, "finding_criteria.0.criterion.2.gte"),
					testAccCheckResourceAttrRfc3339(resourceName, "finding_criteria.0.criterion.2.lt"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "sample",
						"neq.#": "2",
						"neq.0": "another-sample",
						"neq.1": "some-sample",
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

func testAccAwsMacie2FindingsFilter_WithNumber(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	dataSourceRegion := "data.aws_region.current"
	description := "this is a description"
	descriptionUpdated := "this is a description updated"
	firstNumber := "12"
	secondNumber := "13"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigComplete(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttrPair(resourceName, "finding_criteria.0.criterion.0.eq.0", dataSourceRegion, "name"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
				),
			},
			{
				Config: testAccAwsMacieFindingsFilterconfigCompleteMultipleCriterionNumber(descriptionUpdated, macie2.FindingsFilterActionNoop, firstNumber, secondNumber, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionNoop),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "count",
						"gte":   firstNumber,
						"lt":    secondNumber,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "sample",
						"neq.#": "2",
						"neq.0": "another-sample",
						"neq.1": "some-sample",
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

func testAccAwsMacie2FindingsFilter_withTags(t *testing.T) {
	var macie2Output macie2.GetFindingsFilterOutput
	resourceName := "aws_macie2_findings_filter.test"
	description := "this is a description"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsMacie2FindingsFilterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieFindingsFilterconfigWithTags(description, macie2.FindingsFilterActionArchive, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2FindingsFilterExists(resourceName, &macie2Output),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "action", macie2.FindingsFilterActionArchive),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "macie2", regexp.MustCompile(`findings-filter/.+`)),
					resource.TestCheckResourceAttr(resourceName, "position", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field": "region",
						"eq.#":  "1",
						"eq.0":  testAccGetRegion(),
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

func testAccCheckAwsMacie2FindingsFilterExists(resourceName string, macie2Session *macie2.GetFindingsFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}

		resp, err := conn.GetFindingsFilter(input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie FindingsFilter %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccCheckAwsMacie2FindingsFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_findings_filter" {
			continue
		}

		input := &macie2.GetFindingsFilterInput{Id: aws.String(rs.Primary.ID)}
		resp, err := conn.GetFindingsFilter(input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil {
			return fmt.Errorf("macie FindingsFilter %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsMacieFindingsFilterconfigNameGenerated() string {
	return `
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  action = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }
  depends_on = [aws_macie2_account.test]
}
`
}

func testAccAwsMacieFindingsFilterconfigNamePrefix(name string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  name_prefix = %[1]q
  action      = "ARCHIVE"
  finding_criteria {
    criterion {
      field = "region"
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, name)
}

func testAccAwsMacieFindingsFilterconfigComplete(description, action string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position)
}

func testAccAwsMacieFindingsFilterconfigCompleteMultipleCriterion(description, action, startDate, endDate string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
    criterion {
      field = "sample"
      neq   = ["some-sample", "another-sample"]
    }
    criterion {
      field = "updatedAt"
      gte   = %[4]q
      lt    = %[5]q
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position, startDate, endDate)
}

func testAccAwsMacieFindingsFilterconfigCompleteMultipleCriterionNumber(description, action, firstNum, secondNum string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
    criterion {
      field = "sample"
      neq   = ["some-sample", "another-sample"]
    }
    criterion {
      field = "count"
      gte   = %[4]q
      lt    = %[5]q
    }
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position, firstNum, secondNum)
}

func testAccAwsMacieFindingsFilterconfigWithTags(description, action string, position int) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_findings_filter" "test" {
  description = %[1]q
  action      = %[2]q
  position    = %[3]d
  finding_criteria {
    criterion {
      field = "region"
      eq    = [data.aws_region.current.name]
    }
  }
  tags = {
    Key = "value"
  }
  depends_on = [aws_macie2_account.test]
}
`, description, action, position)
}
