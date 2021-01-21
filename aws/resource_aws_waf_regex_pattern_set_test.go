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
	resource.AddTestSweepers("aws_waf_regex_pattern_set", &resource.Sweeper{
		Name: "aws_waf_regex_pattern_set",
		F:    testSweepWafRegexPatternSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafRegexPatternSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListRegexPatternSetsInput{}

	err = lister.ListRegexPatternSetsPages(conn, input, func(page *waf.ListRegexPatternSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, regexPatternSet := range page.RegexPatternSets {
			id := aws.StringValue(regexPatternSet.RegexPatternSetId)

			r := resourceAwsWafRegexPatternSet()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in regex_pattern_strings attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Regex Pattern Set (%s): %w", id, err)
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
				sweeperErr := fmt.Errorf("error deleting WAF Regex Pattern Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Regex Pattern Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Regex Pattern Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

// Serialized acceptance tests due to WAF account limits
// https://docs.aws.amazon.com/waf/latest/developerguide/limits.html
func TestAccAWSWafRegexPatternSet_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":          testAccAWSWafRegexPatternSet_basic,
		"changePatterns": testAccAWSWafRegexPatternSet_changePatterns,
		"noPatterns":     testAccAWSWafRegexPatternSet_noPatterns,
		"disappears":     testAccAWSWafRegexPatternSet_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAWSWafRegexPatternSet_basic(t *testing.T) {
	var v waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexPatternSetExists(resourceName, &v),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`regexpatternset/.+`)),
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

func testAccAWSWafRegexPatternSet_changePatterns(t *testing.T) {
	var before, after waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexPatternSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", patternSetName),
					resource.TestCheckResourceAttr(resourceName, "regex_pattern_strings.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "one"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regex_pattern_strings.*", "two"),
				),
			},
			{
				Config: testAccAWSWafRegexPatternSetConfig_changePatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexPatternSetExists(resourceName, &after),
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

func testAccAWSWafRegexPatternSet_noPatterns(t *testing.T) {
	var patternSet waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexPatternSetConfig_noPatterns(patternSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegexPatternSetExists(resourceName, &patternSet),
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

func testAccAWSWafRegexPatternSet_disappears(t *testing.T) {
	var v waf.RegexPatternSet
	patternSetName := fmt.Sprintf("tfacc-%s", acctest.RandString(5))
	resourceName := "aws_waf_regex_pattern_set.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegexPatternSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegexPatternSetConfig(patternSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegexPatternSetExists(resourceName, &v),
					testAccCheckAWSWafRegexPatternSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafRegexPatternSetDisappears(set *waf.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
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
			return fmt.Errorf("Failed updating WAF Regex Pattern Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRegexPatternSetInput{
				ChangeToken:       token,
				RegexPatternSetId: set.RegexPatternSetId,
			}
			return conn.DeleteRegexPatternSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting WAF Regex Pattern Set: %s", err)
		}

		return nil
	}
}

func testAccCheckAWSWafRegexPatternSetExists(n string, v *waf.RegexPatternSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regex Pattern Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRegexPatternSet(&waf.GetRegexPatternSetInput{
			RegexPatternSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.RegexPatternSet.RegexPatternSetId == rs.Primary.ID {
			*v = *resp.RegexPatternSet
			return nil
		}

		return fmt.Errorf("WAF Regex Pattern Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegexPatternSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_regex_pattern_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRegexPatternSet(&waf.GetRegexPatternSetInput{
			RegexPatternSetId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RegexPatternSet.RegexPatternSetId == rs.Primary.ID {
				return fmt.Errorf("WAF Regex Pattern Set %s still exists", rs.Primary.ID)
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

func testAccAWSWafRegexPatternSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["one", "two"]
}
`, name)
}

func testAccAWSWafRegexPatternSetConfig_changePatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name                  = "%s"
  regex_pattern_strings = ["two", "three", "four"]
}
`, name)
}

func testAccAWSWafRegexPatternSetConfig_noPatterns(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_regex_pattern_set" "test" {
  name = "%s"
}
`, name)
}
