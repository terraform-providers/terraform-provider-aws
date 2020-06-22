package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAWSWafByteMatchSet_basic(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_waf_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafByteMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSet),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.text_transformation", "NONE"),
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
		PreCheck:            func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafByteMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafByteMatchSetConfig(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.target_string", "badrefer2"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.2081155357.text_transformation", "NONE"),
				),
			},
			{
				Config: testAccAWSWafByteMatchSetConfig_changeTuples(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.data", "referer"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.target_string", "badrefer1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3483354334.text_transformation", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.field_to_match.0.type", "METHOD"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.positional_constraint", "CONTAINS_WORD"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.target_string", "blah"),
					resource.TestCheckResourceAttr(resourceName, "byte_match_tuples.3945246213.text_transformation", "URL_DECODE"),
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
