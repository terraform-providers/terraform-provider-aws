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

func TestAccAWSWafRegionalByteMatchSet_basic(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", byteMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", "2"),
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

func TestAccAWSWafRegionalByteMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	byteMatchSetNewName := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", byteMatchSet),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", "2"),
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
				Config: testAccAWSWafRegionalByteMatchSetConfigChangeName(byteMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", byteMatchSetNewName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", "2"),
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

func TestAccAWSWafRegionalByteMatchSet_changeByteMatchTuples(t *testing.T) {
	var before, after waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(
						resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer1",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.0.data": "referer",
						"field_to_match.0.type": "HEADER",
						"positional_constraint": "CONTAINS",
						"target_string":         "badrefer2",
						"text_transformation":   "NONE",
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalByteMatchSetConfigChangeByteMatchTuples(byteMatchSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(
						resourceName, "name", byteMatchSetName),
					resource.TestCheckResourceAttr(
						resourceName, "byte_match_tuples.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "METHOD",
						"positional_constraint": "EXACTLY",
						"target_string":         "GET",
						"text_transformation":   "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "byte_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "URI",
						"positional_constraint": "ENDS_WITH",
						"target_string":         "badrefer2+",
						"text_transformation":   "LOWERCASE",
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

func TestAccAWSWafRegionalByteMatchSet_noByteMatchTuples(t *testing.T) {
	var byteMatchSet waf.ByteMatchSet
	byteMatchSetName := fmt.Sprintf("byte-batch-set-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_match_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig_noDescriptors(byteMatchSetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &byteMatchSet),
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

func TestAccAWSWafRegionalByteMatchSet_disappears(t *testing.T) {
	var v waf.ByteMatchSet
	byteMatchSet := fmt.Sprintf("byteMatchSet-%s", acctest.RandString(5))
	resourceName := "aws_wafregional_byte_match_set.byte_set"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalByteMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalByteMatchSetConfig(byteMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists(resourceName, &v),
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

func testAccAWSWafRegionalByteMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
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

func testAccAWSWafRegionalByteMatchSetConfig_noDescriptors(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_match_set" {
  name = "%s"
}
`, name)
}

func testAccAWSWafRegionalByteMatchSetConfigChangeByteMatchTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_byte_match_set" "byte_set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "GET"
    positional_constraint = "EXACTLY"

    field_to_match {
      type = "METHOD"
    }
  }

  byte_match_tuples {
    text_transformation   = "LOWERCASE"
    target_string         = "badrefer2+"
    positional_constraint = "ENDS_WITH"

    field_to_match {
      type = "URI"
    }
  }
}
`, name)
}
