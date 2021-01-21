package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/wafregional/finder"
)

func init() {
	resource.AddTestSweepers("aws_wafregional_regex_match_set", &resource.Sweeper{
		Name: "aws_wafregional_regex_match_set",
		F:    testSweepWafRegionalRegexMatchSet,
	})
}

func testSweepWafRegionalRegexMatchSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafregionalconn

	var sweeperErrs *multierror.Error

	err = ListRegexMatchSetsPages(conn, &waf.ListRegexMatchSetsInput{}, func(page *waf.ListRegexMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, r := range page.RegexMatchSets {
			id := aws.StringValue(r.RegexMatchSetId)

			set, err := finder.RegexMatchSetByID(conn, id)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving WAF Regional Regex Match Set (%s): %w", id, err))
				continue
			}

			err = deleteRegexMatchSetResource(conn, region, region, id, getRegexMatchTuplesFromAPIResource(set))
			if err != nil {
				if !tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting WAF Regional Regex Match Set (%s): %w", id, err))
				}
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regional Regex Match Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Regional Regex Match Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func ListRegexMatchSetsPages(conn *wafregional.WAFRegional, input *waf.ListRegexMatchSetsInput, fn func(*waf.ListRegexMatchSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListRegexMatchSets(input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextMarker = output.NextMarker
	}
	return nil
}

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegionalRegexMatchSet_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafRegionalRegexMatchSet_basic,
		"changePatterns": testAccAWSWafRegionalRegexMatchSet_changePatterns,
		"noPatterns":     testAccAWSWafRegionalRegexMatchSet_noPatterns,
		"disappears":     testAccAWSWafRegionalRegexMatchSet_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSWafRegionalRegexMatchSet_basic(t *testing.T) {
	var matchSet waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx int

	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	fieldToMatch := waf.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: aws.String("HEADER"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckAWSWafRegionalRegexPatternSetExists("aws_wafregional_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &fieldToMatch, "NONE", &idx),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "NONE",
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

func testAccAWSWafRegionalRegexMatchSet_changePatterns(t *testing.T) {
	var before, after waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx1, idx2 int

	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists(resourceName, &before),
					testAccCheckAWSWafRegionalRegexPatternSetExists("aws_wafregional_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("User-Agent"), Type: aws.String("HEADER")}, "NONE", &idx1),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "user-agent",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig_changePatterns(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "1"),

					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("Referer"), Type: aws.String("HEADER")}, "COMPRESS_WHITE_SPACE", &idx2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "regex_match_tuple.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"text_transformation":   "COMPRESS_WHITE_SPACE",
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

func testAccAWSWafRegionalRegexMatchSet_noPatterns(t *testing.T) {
	var matchSet waf.RegexMatchSet
	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists(resourceName, &matchSet),
					resource.TestCheckResourceAttr(resourceName, "name", matchSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_match_tuple.#", "0"),
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

func testAccAWSWafRegionalRegexMatchSet_disappears(t *testing.T) {
	var matchSet waf.RegexMatchSet
	resourceName := "aws_wafregional_regex_match_set.test"
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafRegionalRegexMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafRegionalRegexMatchSetExists(n string, v *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regional Regex Match Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.RegexMatchSet.RegexMatchSetId == rs.Primary.ID {
			*v = *resp.RegexMatchSet
			return nil
		}

		return fmt.Errorf("WAF Regional Regex Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegionalRegexMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_regex_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RegexMatchSet.RegexMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regional Regex Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Regex Pattern Set is already destroyed
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "User-Agent"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_wafregional_regex_pattern_set.test.id
    text_transformation  = "NONE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegionalRegexMatchSetConfig_changePatterns(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "Referer"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_wafregional_regex_pattern_set.test.id
    text_transformation  = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegionalRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = "%s"
}
`, matchSetName)
}
