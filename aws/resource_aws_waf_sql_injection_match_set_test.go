package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAWSWafSqlInjectionMatchSet_basic(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafSqlInjectionMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.text_transformation", "URL_DECODE"),
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

func TestAccAWSWafSqlInjectionMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameNew := acctest.RandomWithPrefix("tf-acc-test-new")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfigChangeName(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameNew),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_disappears(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &v),
					testAccCheckAWSWafSqlInjectionMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_changeTuples(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSWafSqlInjectionMatchSetDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.2246477904.text_transformation", "URL_DECODE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_noTuples(t *testing.T) {
	var sqlSet waf.SqlInjectionMatchSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafSqlInjectionMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafSqlInjectionMatchSetExists(resourceName, &sqlSet),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "0"),
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

func testAccCheckAWSWafSqlInjectionMatchSetDisappears(v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}

			for _, sqlInjectionMatchTuple := range v.SqlInjectionMatchTuples {
				sqlInjectionMatchTupleUpdate := &waf.SqlInjectionMatchSetUpdate{
					Action: aws.String(waf.ChangeActionDelete),
					SqlInjectionMatchTuple: &waf.SqlInjectionMatchTuple{
						FieldToMatch:       sqlInjectionMatchTuple.FieldToMatch,
						TextTransformation: sqlInjectionMatchTuple.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, sqlInjectionMatchTupleUpdate)
			}
			return conn.UpdateSqlInjectionMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating SqlInjectionMatchSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}
			return conn.DeleteSqlInjectionMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting SqlInjectionMatchSet: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafSqlInjectionMatchSetExists(n string, v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SqlInjectionMatchSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetSqlInjectionMatchSet(&waf.GetSqlInjectionMatchSetInput{
			SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
			*v = *resp.SqlInjectionMatchSet
			return nil
		}

		return fmt.Errorf("WAF SqlInjectionMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSWafSqlInjectionMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_sql_injection_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetSqlInjectionMatchSet(
			&waf.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF SqlInjectionMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the SqlInjectionMatchSet is already destroyed
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafSqlInjectionMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccAWSWafSqlInjectionMatchSetConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccAWSWafSqlInjectionMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccAWSWafSqlInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"
}
`, name)
}
