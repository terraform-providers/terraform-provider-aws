package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/waf/lister"
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
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).wafconn

	var sweeperErrs *multierror.Error

	input := &waf.ListWebACLsInput{}

	err = lister.ListWebACLsPages(conn, input, func(page *waf.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, webACL := range page.WebACLs {
			if webACL == nil {
				continue
			}

			id := aws.StringValue(webACL.WebACLId)

			r := resourceAwsWafWebAcl()
			d := r.Data(nil)
			d.SetId(id)

			// Need to Read first to fill in rules argument
			err := r.Read(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading WAF Web ACL (%s): %w", id, err)
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
				sweeperErr := fmt.Errorf("error deleting WAF Web ACL (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAF Web ACL sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAF Web ACLs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSWafWebAcl_basic(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
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
		ErrorCheck:        testAccErrorCheck(t, waf.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafWebAcl(), resourceName),
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
		ErrorCheck:   testAccErrorCheck(t, waf.EndpointsID),
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

		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading WAF Web ACL (%s): %w", rs.Primary.ID, err)
		}

		if resp != nil && resp.WebACL != nil {
			return fmt.Errorf("WAF Web ACL (%s) still exists", rs.Primary.ID)
		}
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
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, rName)
}

func testAccAWSWafWebAclConfig_DefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = %q
  }
}
`, rName, defaultAction)
}

func testAccAWSWafWebAclConfig_Rules_Single_Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %[1]q
  name        = %[1]q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

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
`, rName)
}

func testAccAWSWafWebAclConfig_Rules_Single_RuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  metric_name = %[1]q
  name        = %[1]q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

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
`, rName)
}

func testAccAWSWafWebAclConfig_Rules_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %[1]q
  name        = %[1]q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_rule_group" "test" {
  metric_name = %[1]q
  name        = %[1]q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

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
`, rName)
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
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSWafWebAclConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %[1]q
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
