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
	resource.AddTestSweepers("aws_waf_xss_match_set", &resource.Sweeper{
		Name: "aws_waf_xss_match_set",
		F:    testSweepWafXssMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafXssMatchSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListXssMatchSetsInput{}

	err = lister.ListXssMatchSetsPages(conn, input, func(page *waf.ListXssMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, xssMatchSet := range page.XssMatchSets {
			id := aws.StringValue(xssMatchSet.XssMatchSetId)

			r := resourceAwsWafXssMatchSet()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in xss_match_tuples attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF XSS Match Set (%s): %w", id, err)
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
				sweeperErr := fmt.Errorf("error deleting WAF XSS Match Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF XSS Match Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF XSS Match Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSWafXssMatchSet_basic(t *testing.T) {
	var v waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafXssMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &v),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`xssmatchset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
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

func TestAccAWSWafXssMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	xssMatchSetNewName := fmt.Sprintf("xssMatchSetNewName-%s", acctest.RandString(5))
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafXssMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
				),
			},
			{
				Config: testAccAWSWafXssMatchSetConfigChangeName(xssMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", xssMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
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

func TestAccAWSWafXssMatchSet_disappears(t *testing.T) {
	var v waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafXssMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafXssMatchSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafXssMatchSet_changeTuples(t *testing.T) {
	var before, after waf.XssMatchSet
	setName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafXssMatchSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccAWSWafXssMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"text_transformation":   "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "xss_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "BODY",
						"text_transformation":   "CMD_LINE",
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

func TestAccAWSWafXssMatchSet_noTuples(t *testing.T) {
	var ipset waf.XssMatchSet
	setName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafXssMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafXssMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuples.#", "0"),
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

func testAccCheckAWSWafXssMatchSetExists(n string, v *waf.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF XSS Match Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetXssMatchSet(&waf.GetXssMatchSetInput{
			XssMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
			*v = *resp.XssMatchSet
			return nil
		}

		return fmt.Errorf("WAF XssMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafXssMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_xss_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetXssMatchSet(
			&waf.GetXssMatchSetInput{
				XssMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF XssMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the XssMatchSet is already destroyed
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafXssMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccAWSWafXssMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccAWSWafXssMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuples {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuples {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccAWSWafXssMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_xss_match_set" "test" {
  name = %[1]q
}
`, name)
}
