package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSWafRegionalXssMatchSet_basic(t *testing.T) {
	var v waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafRegionalXssMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.0.type", "URI"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.text_transformation", "NONE"),
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

func TestAccAWSWafRegionalXssMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.XssMatchSet
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafRegionalXssMatchSetConfigChangeName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalXssMatchSet_disappears(t *testing.T) {
	var v waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &v),
					testAccCheckAWSWafRegionalXssMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalXssMatchSet_changeTuples(t *testing.T) {
	var before, after waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafRegionalXssMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.599421078.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.field_to_match.0.type", "URI"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.41660541.text_transformation", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.42378128.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.42378128.field_to_match.0.data", "GET"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.42378128.field_to_match.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.42378128.text_transformation", "HTML_ENTITY_DECODE"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.3815294338.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.3815294338.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.3815294338.field_to_match.0.type", "BODY"),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.3815294338.text_transformation", "CMD_LINE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalXssMatchSet_noTuples(t *testing.T) {
	var ipset waf.XssMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_xss_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalXssMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalXssMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists(resourceName, &ipset),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "xss_match_tuple.#", "0"),
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

func testAccCheckAWSWafRegionalXssMatchSetDisappears(v *waf.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateXssMatchSetInput{
				ChangeToken:   token,
				XssMatchSetId: v.XssMatchSetId,
			}

			for _, xssMatchTuple := range v.XssMatchTuples {
				xssMatchTupleUpdate := &waf.XssMatchSetUpdate{
					Action: aws.String(wafregional.ChangeActionDelete),
					XssMatchTuple: &waf.XssMatchTuple{
						FieldToMatch:       xssMatchTuple.FieldToMatch,
						TextTransformation: xssMatchTuple.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, xssMatchTupleUpdate)
			}
			return conn.UpdateXssMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating regional WAF XSS Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteXssMatchSetInput{
				ChangeToken:   token,
				XssMatchSetId: v.XssMatchSetId,
			}
			return conn.DeleteXssMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting regional WAF XSS Match Set: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalXssMatchSetExists(n string, v *waf.XssMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No regional WAF XSS Match Set ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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

		return fmt.Errorf("Regional WAF XSS Match Set (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegionalXssMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_xss_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetXssMatchSet(
			&waf.GetXssMatchSetInput{
				XssMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.XssMatchSet.XssMatchSetId == rs.Primary.ID {
				return fmt.Errorf("Regional WAF XSS Match Set %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the regional WAF XSS Match Set is already destroyed
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalXssMatchSetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccAWSWafRegionalXssMatchSetConfigChangeName(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, rName)
}

func testAccAWSWafRegionalXssMatchSetConfig_changeTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "CMD_LINE"

    field_to_match {
      type = "BODY"
    }
  }

  xss_match_tuple {
    text_transformation = "HTML_ENTITY_DECODE"

    field_to_match {
      type = "METHOD"
      data = "GET"
    }
  }
}
`, rName)
}

func testAccAWSWafRegionalXssMatchSetConfig_noTuples(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_xss_match_set" "test" {
  name = %[1]q
}
`, rName)
}
