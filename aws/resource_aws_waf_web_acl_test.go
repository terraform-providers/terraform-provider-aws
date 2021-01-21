package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_waf_web_acl", &resource.Sweeper{
		Name: "aws_waf_web_acl",
		F:    testSweepWafWebAcls,
	})
}

func testSweepWafWebAcls(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	input := &waf.ListWebACLsInput{}

	for {
		output, err := conn.ListWebACLs(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Web ACL sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Web ACLs: %s", err)
		}

		for _, webACL := range output.WebACLs {
			deleteInput := &waf.DeleteWebACLInput{
				WebACLId: webACL.WebACLId,
			}
			id := aws.StringValue(webACL.WebACLId)
			wr := newWafRetryer(conn)

			_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Web ACL: %s", id)
				return conn.DeleteWebACL(deleteInput)
			})

			if isAWSErr(err, waf.ErrCodeNonEmptyEntityException, "") {
				getWebACLInput := &waf.GetWebACLInput{
					WebACLId: webACL.WebACLId,
				}

				getWebACLOutput, getWebACLErr := conn.GetWebACL(getWebACLInput)

				if getWebACLErr != nil {
					return fmt.Errorf("error getting WAF Regional Web ACL (%s): %s", id, getWebACLErr)
				}

				var updates []*waf.WebACLUpdate
				updateWebACLInput := &waf.UpdateWebACLInput{
					DefaultAction: getWebACLOutput.WebACL.DefaultAction,
					Updates:       updates,
					WebACLId:      webACL.WebACLId,
				}

				for _, rule := range getWebACLOutput.WebACL.Rules {
					update := &waf.WebACLUpdate{
						Action:        aws.String(waf.ChangeActionDelete),
						ActivatedRule: rule,
					}

					updateWebACLInput.Updates = append(updateWebACLInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(func(token *string) (interface{}, error) {
					updateWebACLInput.ChangeToken = token
					log.Printf("[INFO] Removing Rules from WAF Regional Web ACL: %s", id)
					return conn.UpdateWebACL(updateWebACLInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing rules from WAF Regional Web ACL (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Web ACL: %s", id)
					return conn.DeleteWebACL(deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Web ACL (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func TestAccAWSWafWebAcl_basic(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`webacl/.+`)),
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

func TestAccAWSWafWebAcl_changeNameForceNew(t *testing.T) {
	var webACL waf.WebACL
	rName1 := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	rName2 := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfig_Required(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func TestAccAWSWafWebAcl_DefaultAction(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfig_DefaultAction(rName, "BLOCK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "BLOCK"),
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

func TestAccAWSWafWebAcl_Rules(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			// Test creating with rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Single_Rule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test adding rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
			// Test removing rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Single_RuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWafWebAcl_LoggingConfiguration(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSWaf(t)
			testAccPreCheckWafLoggingConfiguration(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Logging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.0.field_to_match.#", "2"),
				),
			},
			// Test resource import
			{
				Config:            testAccAWSWafWebAclConfig_Logging(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test logging configuration update
			{
				Config: testAccAWSWafWebAclConfig_LoggingUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", "0"),
				),
			},
			// Test logging configuration removal
			{
				Config: testAccAWSWafWebAclConfig_LoggingRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafWebAcl_disappears(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					testAccCheckAWSWafWebAclDisappears(&webACL),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafWebAcl_Tags(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func testAccCheckAWSWafWebAclDisappears(v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)

		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
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

func testAccCheckAWSWafWebAclDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_web_acl" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
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
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafWebAclExists(n string, v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WebACL ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
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

func testAccAWSWafWebAclConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }
}
`, rName, rName)
}

func testAccAWSWafWebAclConfig_DefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = %q
  }
}
`, rName, rName, defaultAction)
}

func testAccAWSWafWebAclConfig_Rules_Single_Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %q
  name        = %q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule.test.id

    action {
      type = "BLOCK"
    }
  }
}
`, rName, rName, rName, rName, rName)
}

func testAccAWSWafWebAclConfig_Rules_Single_RuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  metric_name = %q
  name        = %q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule_group.test.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSWafWebAclConfig_Rules_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %q
  name        = %q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_rule_group" "test" {
  metric_name = %q
  name        = %q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule.test.id

    action {
      type = "BLOCK"
    }
  }

  rules {
    priority = 2
    rule_id  = aws_waf_rule_group.test.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
`, rName, rName, rName, rName, rName, rName, rName)
}

func testAccAWSWafWebAclConfig_Logging(rName string) string {
	return composeConfig(
		testAccWafLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn

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
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}

func testAccAWSWafWebAclConfig_LoggingRemoved(rName string) string {
	return composeConfig(
		testAccWafLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, rName))
}

func testAccAWSWafWebAclConfig_LoggingUpdate(rName string) string {
	return composeConfig(
		testAccWafLoggingConfigurationRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn
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
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName))
}

func testAccAWSWafWebAclConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
  }
}
`, rName, rName, tag1Key, tag1Value)
}

func testAccAWSWafWebAclConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
