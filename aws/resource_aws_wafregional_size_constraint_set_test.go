package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
)

func TestAccAWSWafRegionalSizeConstraintSet_basic(t *testing.T) {
	var constraints waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &constraints),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.comparison_operator", "EQ"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.281401076.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.281401076.type", "BODY"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.size", "4096"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.text_transformation", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSizeConstraintSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))
	sizeConstraintSetNewName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", sizeConstraintSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "1"),
				),
			},
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfigChangeName(sizeConstraintSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", sizeConstraintSetNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSizeConstraintSet_disappears(t *testing.T) {
	var constraints waf.SizeConstraintSet
	sizeConstraintSet := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig(sizeConstraintSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &constraints),
					testAccCheckAWSWafRegionalSizeConstraintSetDisappears(&constraints),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalSizeConstraintSet_changeConstraints(t *testing.T) {
	var before, after waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.comparison_operator", "EQ"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.281401076.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.field_to_match.281401076.type", "BODY"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.size", "4096"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.2029852522.text_transformation", "NONE"),
				),
			},
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig_changeConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.comparison_operator", "GE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.field_to_match.281401076.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.field_to_match.281401076.type", "BODY"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.size", "1024"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.3222308386.text_transformation", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSizeConstraintSet_noConstraints(t *testing.T) {
	var constraints waf.SizeConstraintSet
	setName := fmt.Sprintf("sizeConstraintSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSizeConstraintSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSizeConstraintSetConfig_noConstraints(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSizeConstraintSetExists("aws_wafregional_size_constraint_set.size_constraint_set", &constraints),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_size_constraint_set.size_constraint_set", "size_constraints.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRegionalSizeConstraintSetDisappears(constraints *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSizeConstraintSetInput{
				ChangeToken:         token,
				SizeConstraintSetId: constraints.SizeConstraintSetId,
			}

			for _, sizeConstraint := range constraints.SizeConstraints {
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
				SizeConstraintSetId: constraints.SizeConstraintSetId,
			}
			return conn.DeleteSizeConstraintSet(opts)
		})

		return err
	}
}

func testAccCheckAWSWafRegionalSizeConstraintSetExists(n string, constraints *waf.SizeConstraintSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SizeConstraintSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetSizeConstraintSet(&waf.GetSizeConstraintSetInput{
			SizeConstraintSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.SizeConstraintSet.SizeConstraintSetId == rs.Primary.ID {
			*constraints = *resp.SizeConstraintSet
			return nil
		}

		return fmt.Errorf("WAF SizeConstraintSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafRegionalSizeConstraintSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_size_contraint_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalSizeConstraintSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = "%s"
  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size = "4096"
    field_to_match {
      type = "BODY"
    }
  }
}`, name)
}

func testAccAWSWafRegionalSizeConstraintSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = "%s"
  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "EQ"
    size = "4096"
    field_to_match {
      type = "BODY"
    }
  }
}`, name)
}

func testAccAWSWafRegionalSizeConstraintSetConfig_changeConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = "%s"
  size_constraints {
    text_transformation = "NONE"
    comparison_operator = "GE"
    size = "1024"
    field_to_match {
      type = "BODY"
    }
  }
}`, name)
}

func testAccAWSWafRegionalSizeConstraintSetConfig_noConstraints(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_size_constraint_set" "size_constraint_set" {
  name = "%s"
}`, name)
}
