package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/waf/lister"
)

func init() {
	resource.AddTestSweepers("aws_waf_byte_match_set", &resource.Sweeper{
		Name: "aws_waf_byte_match_set",
		F:    testSweepWafByteMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafByteMatchSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListByteMatchSetsInput{}

	err = lister.ListByteMatchSetsPages(conn, input, func(page *waf.ListByteMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, byteMatchSet := range page.ByteMatchSets {
			id := aws.StringValue(byteMatchSet.ByteMatchSetId)

			r := resourceAwsWafByteMatchSet()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in byte_match_tuples attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Byte Match Set (%s): %w", id, err)
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
				sweeperErr := fmt.Errorf("error deleting WAF Byte Match Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Byte Match Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Byte Match Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSWafByteMatchSet_basic(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSet),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer2",
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

func TestAccAWSWafByteMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	byteMatchSetNewName := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSet),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
				),
			},
			{
				Config: testAccAWSWafByteMatchSetConfigChangeName(byteMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetNewName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
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

func TestAccAWSWafByteMatchSet_changeTuples(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer2",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccAWSWafByteMatchSetConfig_changeTuples(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"positional_constraint": "CONTAINS_WORD",
						"target_string":         "blah",
						"text_transformation":   "URL_DECODE",
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

func TestAccAWSWafByteMatchSet_noTuples(t *testing.T) {
	var byteSet waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig_noTuples(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &byteSet),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "0"),
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

func TestAccAWSWafByteMatchSet_disappears(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &v),
					testAccCheckAWSWafByteMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafByteMatchSetDisappears(v *waf.ByteMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateByteMatchSetInput{
				ChangeToken:    token,
				ByteMatchSetId: v.ByteMatchSetId,
			}

			for _, ByteMatchTuple := range v.ByteMatchTuples {
				ByteMatchUpdate := &waf.ByteMatchSetUpdate{
					Action: aws.String("DELETE"),
					ByteMatchTuple: &waf.ByteMatchTuple{
						FieldToMatch:         ByteMatchTuple.FieldToMatch,
						PositionalConstraint: ByteMatchTuple.PositionalConstraint,
						TargetString:         ByteMatchTuple.TargetString,
						TextTransformation:   ByteMatchTuple.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, ByteMatchUpdate)
			}

			return conn.UpdateByteMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating ByteMatchSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteByteMatchSetInput{
				ChangeToken:    token,
				ByteMatchSetId: v.ByteMatchSetId,
			}
			return conn.DeleteByteMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting ByteMatchSet: %s", err)
		}

		return nil
	}
}

func testAccCheckAWSWafByteMatchSetExists(n string, v *waf.ByteMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF ByteMatchSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetByteMatchSet(&waf.GetByteMatchSetInput{
			ByteMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.ByteMatchSet.ByteMatchSetId == rs.Primary.ID {
			*v = *resp.ByteMatchSet
			return nil
		}

		return fmt.Errorf("WAF ByteMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafByteMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_byte_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetByteMatchSet(
			&waf.GetByteMatchSetInput{
				ByteMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.ByteMatchSet.ByteMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF ByteMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the ByteMatchSet is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "WAFNonexistentItemException" {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccAWSWafByteMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer2"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}
`, name)
}

func testAccAWSWafByteMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer2"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}
`, name)
}

func testAccAWSWafByteMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation   = "URL_DECODE"
    target_string         = "blah"
    positional_constraint = "CONTAINS_WORD"

    field_to_match {
      type = "METHOD"
      # data field omitted as the type is neither "HEADER" nor "SINGLE_QUERY_ARG"
    }
  }
}
`, name)
}

func testAccAWSWafByteMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_byte_match_set" "byte_set" {
  name = "%s"
}
`, name)
}
