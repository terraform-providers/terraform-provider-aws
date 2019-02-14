package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
)

func TestAccAWSWafRegionalWebAcl_basic(t *testing.T) {
	var v waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					testAccMatchResourceAttrRegionalARN("aws_wafregional_web_acl.waf_acl", "arn", "waf-regional", regexp.MustCompile(`webacl/.+`)),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "logging_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_createRateBased(t *testing.T) {
	var v waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfigRateBased(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_createGroup(t *testing.T) {
	var v waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfigGroup(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_changeNameForceNew(t *testing.T) {
	var before, after waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	wafAclNewName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclName),
				),
			},
			{
				Config: testAccAWSWafRegionalWebAclConfig_changeName(wafAclNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_changeDefaultAction(t *testing.T) {
	var before, after waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	wafAclNewName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclName),
				),
			},
			{
				Config: testAccAWSWafRegionalWebAclConfig_changeDefaultAction(wafAclNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "BLOCK"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "metric_name", wafAclNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_disappears(t *testing.T) {
	var v waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					testAccCheckAWSWafRegionalWebAclDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_noRules(t *testing.T) {
	var v waf.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig_noRules(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_changeRules(t *testing.T) {
	var v waf.WebACL
	var r waf.Rule
	var idx int
	wafAclName := fmt.Sprintf("wafacl%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfig(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &r),
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "1"),
					computeWafRegionalWebAclRuleIndex(&r.RuleId, 1, "REGULAR", "BLOCK", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_web_acl.waf_acl", "rule.%d.priority", &idx, "1"),
				),
			},
			{
				Config: testAccAWSWafRegionalWebAclConfig_changeRules(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists("aws_wafregional_web_acl.waf_acl", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "name", wafAclName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_web_acl.waf_acl", "rule.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalWebAcl_LoggingConfiguration(t *testing.T) {
	var webACL1, webACL2, webACL3 waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.waf_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalWebAclConfigLoggingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists(resourceName, &webACL1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.0.field_to_match.#", "2"),
				),
			},
			// Test logging configuration update
			{
				Config: testAccAWSWafRegionalWebAclConfigLoggingConfigurationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists(resourceName, &webACL2),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "0"),
				),
			},
			// Test logging configuration removal
			{
				Config: testAccAWSWafRegionalWebAclConfig_noRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalWebAclExists(resourceName, &webACL3),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
				),
			},
		},
	})
}

// Calculates the index which isn't static because ruleId is generated as part of the test
func computeWafRegionalWebAclRuleIndex(ruleId **string, priority int, ruleType string, actionType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := resourceAwsWafRegionalWebAcl().Schema["rule"].Elem.(*schema.Resource)
		actionMap := map[string]interface{}{
			"type": actionType,
		}
		m := map[string]interface{}{
			"rule_id":         **ruleId,
			"type":            ruleType,
			"priority":        priority,
			"action":          []interface{}{actionMap},
			"override_action": []interface{}{},
		}

		f := schema.HashResource(ruleResource)
		*idx = f(m)

		return nil
	}
}

func testAccCheckAWSWafRegionalWebAclDisappears(v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateWebACLInput{
				ChangeToken: token,
				WebACLId:    v.WebACLId,
			}

			for _, activatedRule := range v.Rules {
				webACLUpdate := &waf.WebACLUpdate{
					Action: aws.String("DELETE"),
					ActivatedRule: &waf.ActivatedRule{
						Priority: activatedRule.Priority,
						RuleId:   activatedRule.RuleId,
						Action:   activatedRule.Action,
					},
				}
				req.Updates = append(req.Updates, webACLUpdate)
			}

			return conn.UpdateWebACL(req)
		})
		if err != nil {
			return fmt.Errorf("Error getting change token for waf ACL: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteWebACLInput{
				ChangeToken: token,
				WebACLId:    v.WebACLId,
			}
			return conn.DeleteWebACL(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF ACL: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalWebAclDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_web_acl" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetWebACL(
			&waf.GetWebACLInput{
				WebACLId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.WebACL.WebACLId == rs.Primary.ID {
				return fmt.Errorf("WebACL %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WebACL is already destroyed
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRegionalWebAclExists(n string, v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WebACL ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetWebACL(&waf.GetWebACLInput{
			WebACLId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.WebACL.WebACLId == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafRegionalWebAclConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
       type = "BLOCK"
    }
    priority = 1 
    rule_id = "${aws_wafregional_rule.wafrule.id}"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfigRateBased(name string) string {
	return fmt.Sprintf(`

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"

  rate_key   = "IP"
  rate_limit = 2000
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
       type = "BLOCK"
    }
    priority = 1
    type = "RATE_BASED"
    rule_id = "${aws_wafregional_rate_based_rule.wafrule.id}"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfigGroup(name string) string {
	return fmt.Sprintf(`

resource "aws_wafregional_rule_group" "wafrulegroup" {
  name = "%s"
  metric_name = "%s"
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
  rule {
    override_action {
       type = "NONE"
    }
    priority = 1
    type = "GROUP"
    rule_id = "${aws_wafregional_rule_group.wafrulegroup.id}" # todo
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
       type = "BLOCK"
    }
    priority = 1 
    rule_id = "${aws_wafregional_rule.wafrule.id}"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfig_changeDefaultAction(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "BLOCK"
  }
  rule {
    action {
       type = "BLOCK"
    }
    priority = 1 
    rule_id = "${aws_wafregional_rule.wafrule.id}"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfig_noRules(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
}`, name, name)
}

func testAccAWSWafRegionalWebAclConfig_changeRules(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
}

resource "aws_wafregional_web_acl" "waf_acl" {
  name = "%s"
  metric_name = "%s"
  default_action {
    type = "ALLOW"
  }
  rule {
    action {
       type = "ALLOW"
    }
    priority = 3
    rule_id = "${aws_wafregional_rule.wafrule.id}"
  }

  rule {
    action {
       type = "BLOCK"
    }
    priority = 99
    rule_id = "${aws_wafregional_rule.wafrule.id}"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalWebAclConfigLoggingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "waf_acl" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = "${aws_kinesis_firehose_delivery_stream.test.arn}"

    redacted_fields {
      field_to_match {
        type = "URI"
      }

      field_to_match {
        data = "referer"
        type = "HEADER"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = "${aws_iam_role.test.arn}"
    bucket_arn = "${aws_s3_bucket.test.arn}"
  }
}
`, rName)
}

func testAccAWSWafRegionalWebAclConfigLoggingConfigurationUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "waf_acl" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = "${aws_kinesis_firehose_delivery_stream.test.arn}"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "s3"

  s3_configuration {
    role_arn   = "${aws_iam_role.test.arn}"
    bucket_arn = "${aws_s3_bucket.test.arn}"
  }
}
`, rName)
}
