package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func testAccAwsGuardDutyFilter_basic(t *testing.T) {
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"
	detectorResourceName := "aws_guardduty_detector.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "action", "ARCHIVE"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "guardduty", regexp.MustCompile("detector/[a-z0-9]{32}/filter/test-filter$")),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":    "region",
						"equals.#": "1",
						"equals.0": "eu-west-1",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":        "service.additionalInfo.threatListName",
						"not_equals.#": "2",
						"not_equals.0": "some-threat",
						"not_equals.1": "another-threat",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":                 "updatedAt",
						"greater_than_or_equal": startDate,
						"less_than":             endDate,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGuardDutyFilterConfigNoop_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", detectorResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "action", "NOOP"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a NOOP"),
					resource.TestCheckResourceAttr(resourceName, "rank", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
		},
	})
}

func testAccAwsGuardDutyFilter_update(t *testing.T) {
	var v1, v2 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "3"),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "finding_criteria.0.criterion.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":    "region",
						"equals.#": "1",
						"equals.0": testAccGetRegion(),
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "finding_criteria.0.criterion.*", map[string]string{
						"field":        "service.additionalInfo.threatListName",
						"not_equals.#": "2",
						"not_equals.0": "some-threat",
						"not_equals.1": "yet-another-threat",
					}),
				),
			},
		},
	})
}

func testAccAwsGuardDutyFilter_tags(t *testing.T) {
	var v1, v2, v3 guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_multipleTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-filter"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Value"),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_updateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "Updated"),
				),
			},
			{
				Config: testAccGuardDutyFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAwsGuardDutyFilter_disappears(t *testing.T) {
	var v guardduty.GetFilterOutput
	resourceName := "aws_guardduty_filter.test"

	startDate := "2020-01-01T00:00:00Z"
	endDate := "2020-02-01T00:00:00Z"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyFilterConfig_full(startDate, endDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyFilterExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsGuardDutyFilter(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsGuardDutyFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_filter" {
			continue
		}

		detectorID, filterName, err := guardDutyFilterParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(filterName),
		}

		_, err = conn.GetFilter(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected GuardDuty Filter to be destroyed, %s found", rs.Primary.Attributes["filter_name"])
	}

	return nil
}

func testAccCheckAwsGuardDutyFilterExists(name string, filter *guardduty.GetFilterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No GuardDuty filter is set")
		}

		detectorID, name, err := guardDutyFilterParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		input := guardduty.GetFilterInput{
			DetectorId: aws.String(detectorID),
			FilterName: aws.String(name),
		}
		filter, err = conn.GetFilter(&input)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccGuardDutyFilterConfig_full(startDate, endDate string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = ["eu-west-1"]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "another-threat"]
    }

    criterion {
      field                 = "updatedAt"
      greater_than_or_equal = %[1]q
      less_than             = %[2]q
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`, startDate, endDate)
}

func testAccGuardDutyFilterConfigNoop_full(startDate, endDate string) string {
	return fmt.Sprintf(`
resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "NOOP"
  description = "This is a NOOP"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = ["eu-west-1"]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "another-threat"]
    }

    criterion {
      field                 = "updatedAt"
      greater_than_or_equal = %[1]q
      less_than             = %[2]q
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`, startDate, endDate)
}

func testAccGuardDutyFilterConfig_multipleTags() string {
	return `
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.name]
    }
  }

  tags = {
    Name = "test-filter"
    Key  = "Value"
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccGuardDutyFilterConfig_update() string {
	return `
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.name]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "yet-another-threat"]
    }
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}

func testAccGuardDutyFilterConfig_updateTags() string {
	return `
data "aws_region" "current" {}

resource "aws_guardduty_filter" "test" {
  detector_id = aws_guardduty_detector.test.id
  name        = "test-filter"
  action      = "ARCHIVE"
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = [data.aws_region.current.name]
    }
  }

  tags = {
    Key = "Updated"
  }
}

resource "aws_guardduty_detector" "test" {
  enable = true
}
`
}
