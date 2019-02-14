package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	req := &waf.ListRegexMatchSetsInput{}
	resp, err := conn.ListRegexMatchSets(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Regex Match Set sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing WAF Regional Regex Match Sets: %s", err)
	}

	if len(resp.RegexMatchSets) == 0 {
		log.Print("[DEBUG] No AWS WAF Regional Regex Match Sets to sweep")
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
		err = updateRegexMatchSetResourceWR(*set.RegexMatchSetId, oldTuples, noTuples, conn, region)
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional Regex Match Set: %s", err)
		}

		wr := newWafRegionalRetryer(conn, region)
		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.DeleteRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}
			log.Printf("[INFO] Deleting WAF Regional Regex Match Set: %s", req)
			return conn.DeleteRegexMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("error deleting WAF Regional Regex Match Set (%s): %s", aws.StringValue(set.RegexMatchSetId), err)
		}
	}

	return nil
}

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegionalRegexMatchSet(t *testing.T) {
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

	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	fieldToMatch := waf.FieldToMatch{
		Data: aws.String("User-Agent"),
		Type: aws.String("HEADER"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists("aws_wafregional_regex_match_set.test", &matchSet),
					testAccCheckAWSWafRegionalRegexPatternSetExists("aws_wafregional_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &fieldToMatch, "NONE", &idx),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "regex_match_tuple.#", "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx, "user-agent"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx, "NONE"),
				),
			},
		},
	})
}

func testAccAWSWafRegionalRegexMatchSet_changePatterns(t *testing.T) {
	var before, after waf.RegexMatchSet
	var patternSet waf.RegexPatternSet
	var idx1, idx2 int

	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists("aws_wafregional_regex_match_set.test", &before),
					testAccCheckAWSWafRegionalRegexPatternSetExists("aws_wafregional_regex_pattern_set.test", &patternSet),
					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("User-Agent"), Type: aws.String("HEADER")}, "NONE", &idx1),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "regex_match_tuple.#", "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx1, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx1, "user-agent"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx1, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx1, "NONE"),
				),
			},
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig_changePatterns(matchSetName, patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists("aws_wafregional_regex_match_set.test", &after),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "regex_match_tuple.#", "1"),

					computeWafRegexMatchSetTuple(&patternSet, &waf.FieldToMatch{Data: aws.String("Referer"), Type: aws.String("HEADER")}, "COMPRESS_WHITE_SPACE", &idx2),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.#", &idx2, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.data", &idx2, "referer"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.field_to_match.0.type", &idx2, "HEADER"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_regex_match_set.test", "regex_match_tuple.%d.text_transformation", &idx2, "COMPRESS_WHITE_SPACE"),
				),
			},
		},
	})
}

func testAccAWSWafRegionalRegexMatchSet_noPatterns(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig_noPatterns(matchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists("aws_wafregional_regex_match_set.test", &matchSet),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "name", matchSetName),
					resource.TestCheckResourceAttr("aws_wafregional_regex_match_set.test", "regex_match_tuple.#", "0"),
				),
			},
		},
	})
}

func testAccAWSWafRegionalRegexMatchSet_disappears(t *testing.T) {
	var matchSet waf.RegexMatchSet
	matchSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexMatchSetConfig(matchSetName, patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexMatchSetExists("aws_wafregional_regex_match_set.test", &matchSet),
					testAccCheckAWSWafRegionalRegexMatchSetDisappears(&matchSet),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafRegionalRegexMatchSetDisappears(set *waf.RegexMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
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
			return fmt.Errorf("Failed updating WAF Regional Regex Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRegexMatchSetInput{
				ChangeToken:     token,
				RegexMatchSetId: set.RegexMatchSetId,
			}
			return conn.DeleteRegexMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regional Regex Match Set: %s", err)
		}

		return nil
	}
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
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
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
    regex_pattern_set_id = "${aws_wafregional_regex_pattern_set.test.id}"
    text_transformation = "NONE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name = "%s"
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
    regex_pattern_set_id = "${aws_wafregional_regex_pattern_set.test.id}"
    text_transformation = "COMPRESS_WHITE_SPACE"
  }
}

resource "aws_wafregional_regex_pattern_set" "test" {
  name = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, matchSetName, patternSetName)
}

func testAccAWSWafRegionalRegexMatchSetConfig_noPatterns(matchSetName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_match_set" "test" {
  name = "%s"
}`, matchSetName)
}
