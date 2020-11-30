package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegionalRegexPatternSet_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafRegionalRegexPatternSet_basic,
		"changePatterns": testAccAWSWafRegionalRegexPatternSet_changePatterns,
		"noPatterns":     testAccAWSWafRegionalRegexPatternSet_noPatterns,
		"disappears":     testAccAWSWafRegionalRegexPatternSet_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSWafRegionalRegexPatternSet_basic(t *testing.T) {
	var patternSet waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexPatternSetExists(resourceName, &patternSet),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "one"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "two"),
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

func testAccAWSWafRegionalRegexPatternSet_changePatterns(t *testing.T) {
	var before, after waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexPatternSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "one"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "two"),
				),
			},
			{
				Config: testAccAWSWafRegionalRegexPatternSetConfig_changePatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexPatternSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "two"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "three"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "four"),
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

func testAccAWSWafRegionalRegexPatternSet_noPatterns(t *testing.T) {
	var patternSet waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexPatternSetConfig_noPatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRegexPatternSetExists(resourceName, &patternSet),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "0"),
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

func testAccAWSWafRegionalRegexPatternSet_disappears(t *testing.T) {
	var patternSet waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRegexPatternSetExists(resourceName, &patternSet),
					testAccCheckAWSWafRegionalRegexPatternSetDisappears(&patternSet),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafRegionalRegexPatternSetDisappears(set *waf.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRegexPatternSetInput{
				ChangeToken:       token,
				RegexPatternSetId: set.RegexPatternSetId,
			}

			for _, pattern := range set.RegexPatternStrings {
				update := &waf.RegexPatternSetUpdate{
					Action:             aws.String("DELETE"),
					RegexPatternString: pattern,
				}
				req.Updates = append(req.Updates, update)
			}

			return conn.UpdateRegexPatternSet(req)
		})
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regional Regex Pattern Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRegexPatternSetInput{
				ChangeToken:       token,
				RegexPatternSetId: set.RegexPatternSetId,
			}
			return conn.DeleteRegexPatternSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regional Regex Pattern Set: %s", err)
		}

		return nil
	}
}

func testAccCheckAWSWafRegionalRegexPatternSetExists(n string, patternSet *waf.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regional Regex Pattern Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRegexPatternSet(&waf.GetRegexPatternSetInput{
			RegexPatternSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.RegexPatternSet.RegexPatternSetId == rs.Primary.ID {
			*patternSet = *resp.RegexPatternSet
			return nil
		}

		return fmt.Errorf("WAF Regional Regex Pattern Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegionalRegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_regex_pattern_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRegexPatternSet(&waf.GetRegexPatternSetInput{
			RegexPatternSetId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RegexPatternSet.RegexPatternSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regional Regex Pattern Set %s still exists", rs.Primary.ID)
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

func testAccAWSWafRegionalRegexPatternSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, name)
}

func testAccAWSWafRegionalRegexPatternSetConfig_changePatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["two", "three", "four"]
}
`, name)
}

func testAccAWSWafRegionalRegexPatternSetConfig_noPatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_regex_pattern_set" "test" {
  name = "%s"
}
`, name)
}
