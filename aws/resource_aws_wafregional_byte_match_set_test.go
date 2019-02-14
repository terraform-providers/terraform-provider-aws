package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
)

func TestAccAWSWafRegionalByteMatchSet_basic(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "name", byteMatchSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.text_transformation", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalByteMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	byteMatchSetNewName := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "name", byteMatchSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.text_transformation", "NONE"),
				),
			},
			{
				Config: testAccAWSWafRegionalByteMatchSetConfigChangeName(byteMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "name", byteMatchSetNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.text_transformation", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalByteMatchSet_changeByteMatchTuples(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "name", byteMatchSetName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2081155357.text_transformation", "NONE"),
				),
			},
			{
				Config: testAccAWSWafRegionalByteMatchSetConfigChangeByteMatchTuples(byteMatchSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "name", byteMatchSetName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.field_to_match.0.type", "METHOD"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.positional_constraint", "EXACTLY"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.target_string", "GET"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2069994922.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.field_to_match.0.type", "URI"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.positional_constraint", "ENDS_WITH"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.target_string", "badrefer2+"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_byte_match_set.byte_set", "byte_match_tuples.2506804529.text_transformation", "LOWERCASE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalByteMatchSet_noByteMatchTuples(t *testing.T) {
	var byteMatchSet waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig_noDescriptors(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_match_set", &byteMatchSet),
					resource.TestCheckResourceAttr("aws_wafregional_byte_match_set.byte_match_set", "name", byteMatchSetName),
					resource.TestCheckResourceAttr("aws_wafregional_byte_match_set.byte_match_set", "byte_match_tuples.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalByteMatchSet_disappears(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.byte_set", &v),
					testAccCheckAWSWafRegionalByteMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafRegionalByteMatchSetDisappears(v *waf.ByteMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
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

func testAccCheckAWSWafRegionalByteMatchSetExists(n string, v *waf.ByteMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF ByteMatchSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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

func testAccCheckAWSWafRegionalByteMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_byte_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetByteMatchSet(
			&waf.GetByteMatchSetInput{
				ByteMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.ByteMatchSet.ByteMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF ByteMatchSet %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			continue
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalByteMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = "%s"
  byte_match_tuples {
    text_transformation = "NONE"
    target_string = "badrefer1"
    positional_constraint = "CONTAINS"
    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation = "NONE"
    target_string = "badrefer2"
    positional_constraint = "CONTAINS"
    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}`, name)
}

func testAccAWSWafRegionalByteMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = "%s"
  byte_match_tuples {
    text_transformation = "NONE"
    target_string = "badrefer1"
    positional_constraint = "CONTAINS"
    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }

  byte_match_tuples {
    text_transformation = "NONE"
    target_string = "badrefer2"
    positional_constraint = "CONTAINS"
    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}`, name)
}

func testAccAWSWafRegionalByteMatchSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_match_set" {
	name = "%s"
}`, name)
}

func testAccAWSWafRegionalByteMatchSetConfigChangeByteMatchTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = "%s"
  byte_match_tuples {
    text_transformation = "NONE"
    target_string = "GET"
    positional_constraint = "EXACTLY"
    field_to_match {
      type = "METHOD"
    }
  }

  byte_match_tuples {
    text_transformation = "LOWERCASE"
    target_string = "badrefer2+"
    positional_constraint = "ENDS_WITH"
    field_to_match {
      type = "URI"
    }
  }
}`, name)
}
