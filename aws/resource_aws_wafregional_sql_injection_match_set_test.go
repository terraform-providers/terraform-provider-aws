package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSWafRegionalSqlInjectionMatchSet_basic(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", sqlInjectionMatchSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.text_transformation", "URL_DECODE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSqlInjectionMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", acctest.RandString(5))
	sqlInjectionMatchSetNewName := fmt.Sprintf("sqlInjectionMatchSetNewName-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", sqlInjectionMatchSet),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "1"),
				),
			},
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig(sqlInjectionMatchSetNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", sqlInjectionMatchSetNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSqlInjectionMatchSet_disappears(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	sqlInjectionMatchSet := fmt.Sprintf("sqlInjectionMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig(sqlInjectionMatchSet),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &v),
					testAccCheckAWSWafRegionalSqlInjectionMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalSqlInjectionMatchSet_changeTuples(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.0.data", ""),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.field_to_match.0.type", "QUERY_STRING"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.1913782288.text_transformation", "URL_DECODE"),
				),
			},
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig_changeTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.3961339938.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.3961339938.field_to_match.0.data", "user-agent"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.3961339938.field_to_match.0.type", "HEADER"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.3961339938.text_transformation", "NONE"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalSqlInjectionMatchSet_noTuples(t *testing.T) {
	var ipset waf.SqlInjectionMatchSet
	setName := fmt.Sprintf("sqlInjectionMatchSet-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalSqlInjectionMatchSetConfig_noTuples(setName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalSqlInjectionMatchSetExists("aws_wafregional_sql_injection_match_set.sql_injection_match_set", &ipset),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "name", setName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_sql_injection_match_set.sql_injection_match_set", "sql_injection_match_tuple.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRegionalSqlInjectionMatchSetDisappears(v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}

			for _, sqlInjectionMatchTuple := range v.SqlInjectionMatchTuples {
				sqlInjectionMatchTupleUpdate := &waf.SqlInjectionMatchSetUpdate{
					Action: aws.String("DELETE"),
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
			return fmt.Errorf("Failed updating SQL Injection Match Set: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}
			return conn.DeleteSqlInjectionMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Failed deleting SQL Injection Match Set: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalSqlInjectionMatchSetExists(n string, v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SqlInjectionMatchSet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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

func testAccCheckAWSWafRegionalSqlInjectionMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_sql_injection_match_set" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSWafRegionalSqlInjectionMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"
  sql_injection_match_tuple {
    text_transformation = "URL_DECODE"
    field_to_match {
      type = "QUERY_STRING"
    }
  }
}`, name)
}

func testAccAWSWafRegionalSqlInjectionMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"
  sql_injection_match_tuple {
    text_transformation = "NONE"
    field_to_match {
      type = "HEADER"
      data = "User-Agent"
    }
  }
}`, name)
}

func testAccAWSWafRegionalSqlInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_sql_injection_match_set" "sql_injection_match_set" {
  name = "%s"
}`, name)
}
