package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/waf/lister"
)

func init() {
	resource.AddTestSweepers("aws_waf_regex_match_set", &resource.Sweeper{
		Name: "aws_waf_regex_match_set",
		F:    testSweepWafRegexMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafRegexMatchSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListRegexMatchSetsInput{}

	err = lister.ListRegexMatchSetsPages(conn, input, func(page *waf.ListRegexMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexMatchSet := range page.RegexMatchSets {
			id := aws.StringValue(regexMatchSet.RegexMatchSetId)

			r := resourceAwsWafRegexMatchSet()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in regex_match_tuple attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Regex Match Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			// In case it was already deleted
			if d.Id() == "" {
				continue
			}

			err = r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAF Regex Match Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regex Match Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Regex Match Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegexMatchSet_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafRegexMatchSet_basic,
		"changePatterns": testAccAWSWafRegexMatchSet_changePatterns,
		"noPatterns":     testAccAWSWafRegexMatchSet_noPatterns,
		"disappears":     testAccAWSWafRegexMatchSet_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSWafRegexMatchSet_basic(t *testing.T) {
	var matchSet waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx int

	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	fieldToMatch := waf.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: aws.String("HEADER"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckAWSWafRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`regexmatchset/.+`)),
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

func testAccAWSWafRegexMatchSet_changePatterns(t *testing.T) {
	var before, after waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx1, idx2 int

	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists(resourceName, &before),
					testAccCheckAWSWafRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
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
				Config: testAccAWSWafRegexMatchSetConfig_changePatterns(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists(resourceName, &after),
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

func testAccAWSWafRegexMatchSet_noPatterns(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists(resourceName, &matchSet),
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

func testAccAWSWafRegexMatchSet_disappears(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_match_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists(resourceName, &matchSet),
					testAccCheckAWSWafRegexMatchSetDisappears(&matchSet),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func computeWafRegexMatchSetTuple(patternSet *waf.RegexPatternSet, fieldToMatch *waf.FieldToMatch, textTransformation string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := map[string]interface{}{
			"field_to_match":       flattenFieldToMatch(fieldToMatch),
			"regex_pattern_set_id": *patternSet.RegexPatternSetId,
			"text_transformation":  textTransformation,
		}

		*idx = resourceAwsWafRegexMatchSetTupleHash(m)

		return nil
	}
}

func testAccCheckAWSWafRegexMatchSetDisappears(set *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}

			for _, tuple := range set.RegexMatchTuples {
				req.Updates = append(req.Updates, &waf.RegexMatchSetUpdate{
					Action:          aws.String("DELETE"),
					RegexMatchTuple: tuple,
				})
			}

			return conn.UpdateRegexMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regex Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}
			return conn.DeleteRegexMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regex Match Set: %s", err)
		}

		return nil
	}
}

func testAccCheckAWSWafRegexMatchSetExists(n string, v *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regex Match Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
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

		return fmt.Errorf("WAF Regex Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegexMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_regex_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RegexMatchSet.RegexMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regex Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Regex Pattern Set is already destroyed
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "User-Agent"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_waf_regex_pattern_set.test.id
    text_transformation  = "NONE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegexMatchSetConfig_changePatterns(matchSetName, patternSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"

  regex_match_tuple {
    field_to_match {
      data = "Referer"
      type = "HEADER"
    }

    regex_pattern_set_id = aws_waf_regex_pattern_set.test.id
    text_transformation  = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"
}
`, matchSetName)
}
