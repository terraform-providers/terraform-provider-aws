package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_waf_regex_match_set", &resource.Sweeper{
		Name: "aws_waf_regex_match_set",
		F:    testSweepWafRegexMatchSet,
	})
}

func testSweepWafRegexMatchSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	req := &waf.ListRegexMatchSetsInput{}
	resp, err := conn.ListRegexMatchSets(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regex Match Set sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing WAF Regex Match Sets: %s", err)
	}

	if len(resp.RegexMatchSets) == 0 {
		log.Print("[DEBUG] No AWS WAF Regex Match Sets to sweep")
		return nil
	}

	for _, s := range resp.RegexMatchSets {
		if !strings.HasPrefix(*s.Name, "tfacc") {
			continue
		}

		resp, err := conn.GetRegexMatchSet(&waf.GetRegexMatchSetInput{
			RegexMatchSetId: s.RegexMatchSetId,
		})
		if err != nil {
			return err
		}
		set := resp.RegexMatchSet

		oldTuples := flattenWafRegexMatchTuples(set.RegexMatchTuples)
		noTuples := []interface{}{}
		err = updateRegexMatchSetResource(*set.RegexMatchSetId, oldTuples, noTuples, conn)
		if err != nil {
			return fmt.Errorf("Error updating WAF Regex Match Set: %s", err)
		}

		wr := newWafRetryer(conn)
		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.DeleteRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}
			log.Printf("[INFO] Deleting WAF Regex Match Set: %s", req)
			return conn.DeleteRegexMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("error deleting WAF Regex Match Set (%s): %s", aws.StringValue(set.RegexMatchSetId), err)
		}
	}

	return nil
}

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegexMatchSet(t *testing.T) {
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

	fieldToMatch := waf.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: aws.String("HEADER"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists("aws_waf_regex_match_set.test", &matchSet),
					testAccCheckAWSWafRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &fieldToMatch, "NONE", &idx),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "regex_match_tuple.#", "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx, "user-agent"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx, "NONE"),
				),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists("aws_waf_regex_match_set.test", &before),
					testAccCheckAWSWafRegexPatternSetExists("aws_waf_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("User-Agent"), Type: aws.String("HEADER")}, "NONE", &idx1),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "regex_match_tuple.#", "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx1, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx1, "user-agent"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx1, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx1, "NONE"),
				),
			},
			{
				Config: testAccAWSWafRegexMatchSetConfig_changePatterns(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists("aws_waf_regex_match_set.test", &after),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "regex_match_tuple.#", "1"),

					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("Referer"), Type: aws.String("HEADER")}, "COMPRESS_WHITE_SPACE", &idx2),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx2, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx2, "referer"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx2, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx2, "COMPRESS_WHITE_SPACE"),
				),
			},
		},
	})
}

func testAccAWSWafRegexMatchSet_noPatterns(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists("aws_waf_regex_match_set.test", &matchSet),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_waf_regex_match_set.test", "regex_match_tuple.#", "0"),
				),
			},
		},
	})
}

func testAccAWSWafRegexMatchSet_disappears(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexMatchSetExists("aws_waf_regex_match_set.test", &matchSet),
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
		if isAWSErr(err, "WAFNonexistentItemException", "") {
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
    regex_pattern_set_id = "${aws_waf_regex_pattern_set.test.id}"
    text_transformation = "NONE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"
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
    regex_pattern_set_id = "${aws_waf_regex_pattern_set.test.id}"
    text_transformation = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_match_set" "test" {
  name = "%s"
}`, matchSetName)
}
