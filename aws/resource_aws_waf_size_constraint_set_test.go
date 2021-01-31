package aws

import (
	"fmt"
	"log"
	"regexp"
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
	resource.AddTestSweepers("aws_waf_size_constraint_set", &resource.Sweeper{
		Name: "aws_waf_size_constraint_set",
		F:    testSweepWafSizeConstraintSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func testSweepWafSizeConstraintSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListSizeConstraintSetsInput{}

	err = lister.ListSizeConstraintSetsPages(conn, input, func(page *waf.ListSizeConstraintSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sizeConstraintSet := range page.SizeConstraintSets {
			id := aws.StringValue(sizeConstraintSet.SizeConstraintSetId)

			r := resourceAwsWafSizeConstraintSet()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in size_constraints attribute
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Size Constraint Set (%s): %w", id, err)
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
				sweeperErr := fmt.Errorf("error deleting WAF Size Constraint Set (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Size Constraint Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Size Constraint Sets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSWafSizeConstraintSet_basic(t *testing.T) {
	var v waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &v),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`sizeconstraintset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						"size":                "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
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

func TestAccAWSWafSizeConstraintSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	sizeConstraintSetNewName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
				),
			},
			{
				Config: testAccAWSWafSizeConstraintSetConfigChangeName(sizeConstraintSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", sizeConstraintSetNewName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
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

func TestAccAWSWafSizeConstraintSet_disappears(t *testing.T) {
	var v waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &v),
					testAccCheckAWSWafSizeConstraintSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafSizeConstraintSet_changeConstraints(t *testing.T) {
	var before, after waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSizeConstraintSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "EQ",
						"field_to_match.#":    "1",
						"size":                "4096",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
					}),
				),
			},
			{
				Config: testAccAWSWafSizeConstraintSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*", map[string]string{
						"comparison_operator": "GE",
						"field_to_match.#":    "1",
						"size":                "1024",
						"text_transformation": "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "size_constraints.*.field_to_match.*", map[string]string{
						"data": "",
						"type": "BODY",
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

func TestAccAWSWafSizeConstraintSet_noConstraints(t *testing.T) {
	var contraints waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_size_constraint_set.size_constraint_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSizeConstraintSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSizeConstraintSetExists(resourceName, &contraints),
					resource.TestCheckResourceAttr(resourceName, "name", setName),
					resource.TestCheckResourceAttr(resourceName, "size_constraints.#", "0"),
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

func testAccCheckAWSWafSizeConstraintSetDisappears(v *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSizeConstraintSetInput{
				ChangeToken:         token,
				SizeConstraintSetId: v.SizeConstraintSetId,
			}

			for _, sizeConstraint := range v.SizeConstraints {
				sizeConstraintUpdate := &waf.SizeConstraintSetUpdate{
					Action: aws.String("DELETE"),
					SizeConstraint: &waf.SizeConstraint{
						FieldToMatch:       sizeConstraint.FieldToMatch,
						ComparisonOperator: sizeConstraint.ComparisonOperator,
						Size:               sizeConstraint.Size,
						TextTransformation: sizeConstraint.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, sizeConstraintUpdate)
			}
			return conn.UpdateSizeConstraintSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating SizeConstraintSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSizeConstraintSetInput{
				ChangeToken:         token,
				SizeConstraintSetId: v.SizeConstraintSetId,
			}
			return conn.DeleteSizeConstraintSet(opts)
		})

		return err
	}
}

func testAccCheckAWSWafSizeConstraintSetExists(n string, v *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SizeConstraintSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetSizeConstraintSet(&waf.GetSizeConstraintSetInput{
			SizeConstraintSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.SizeConstraintSet.SizeConstraintSetId == rs.Primary.ID {
			*v = *resp.SizeConstraintSet
			return nil
		}

		return fmt.Errorf("WAF SizeConstraintSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafSizeConstraintSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_size_contraint_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetSizeConstraintSet(
			&waf.GetSizeConstraintSetInput{
				SizeConstraintSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.SizeConstraintSet.SizeConstraintSetId == rs.Primary.ID {
				return fmt.Errorf("WAF SizeConstraintSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the SizeConstraintSet is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == waf.ErrCodeNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccAWSWafSizeConstraintSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccAWSWafSizeConstraintSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size                = "4096"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccAWSWafSizeConstraintSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"

  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "GE"
    size                = "1024"

    field_to_match {
      type = "BODY"
    }
  }
}
`, name)
}

func testAccAWSWafSizeConstraintSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_size_constraint_set" "size_constraint_set" {
  name = "%s"
}
`, name)
}
