package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsWafv2WebACLLoggingConfiguration_basic(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACLLoggingConfiguration_update(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_updateTwoRedactedFields(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "referer",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "user-agent",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_updateOneRedactedField(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "redacted_fields.*", map[string]string{
						"single_header.0.name": "user-agent",
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

func TestAccAwsWafv2WebACLLoggingConfiguration_changeResourceARNForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameNew := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttr(webACLResourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttr(webACLResourceName, "name", rNameNew),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACLLoggingConfiguration_changeLogDestinationConfigsForceNew(t *testing.T) {
	var before, after wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameNew := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	kinesisResourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &before),
					resource.TestCheckResourceAttr(kinesisResourceName, "name", fmt.Sprintf("aws-waf-logs-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_updateLogDestination(rName, rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &after),
					resource.TestCheckResourceAttr(kinesisResourceName, "name", fmt.Sprintf("aws-waf-logs-%s", rNameNew)),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", webACLResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "redacted_fields.#", "0"),
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

func TestAccAwsWafv2WebACLLoggingConfiguration_disappears(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2WebACLLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsWafv2WebACLLoggingConfiguration_disappears_WebAcl(t *testing.T) {
	var v wafv2.LoggingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl_logging_configuration.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLLoggingConfiguration_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLLoggingConfigurationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2WebACL(), webACLResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsWafv2WebACLLoggingConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl_logging_configuration" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetLoggingConfiguration(
			&wafv2.GetLoggingConfigurationInput{
				ResourceArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Continue checking resources in state if a WebACL Logging Configuration is already destroyed
			if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
				continue
			}
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}
		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			return fmt.Errorf("WAFv2 WebACL Logging Configuration for WebACL ARN %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWafv2WebACLLoggingConfigurationExists(n string, v *wafv2.LoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL Logging Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetLoggingConfiguration(&wafv2.GetLoggingConfigurationInput{
			ResourceArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.LoggingConfiguration == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL Logging Configuration")
		}

		if aws.StringValue(resp.LoggingConfiguration.ResourceArn) == rs.Primary.ID {
			*v = *resp.LoggingConfiguration
			return nil
		}

		return fmt.Errorf("WAFv2 WebACL Logging Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccWebACLLoggingConfigurationDependenciesConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "firehose" {
  name = "%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF

}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
  acl    = "private"
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s"
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": "iam:CreateServiceLinkedRole",
      "Resource": "arn:${data.aws_partition.current.partition}:iam::*:role/aws-service-role/wafv2.${data.aws_partition.current.dns_suffix}/AWSServiceRoleForWAFV2Logging",
      "Condition": {
        "StringLike": {
          "iam:AWSServiceName": "wafv2.${data.aws_partition.current.dns_suffix}"
        }
      }
    }
  ]
}
EOF

}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = "aws-waf-logs-%s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

const testAccWebACLLoggingConfigurationResourceConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]
}
`

const testAccWebACLLoggingConfigurationResourceUpdateTwoRedactedFieldsConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    single_header {
      name = "referer"
    }
  }

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

const testAccWebACLLoggingConfigurationResourceUpdateOneRedactedFieldConfig = `
resource "aws_wafv2_web_acl_logging_configuration" "test" {
  resource_arn            = aws_wafv2_web_acl.test.arn
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.test.arn]

  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
`

func testAccAwsWafv2WebACLLoggingConfiguration_basic(rName string) string {
	return composeConfig(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_updateLogDestination(rName, rNameNew string) string {
	return composeConfig(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rNameNew),
		testAccWebACLLoggingConfigurationResourceConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_updateTwoRedactedFields(rName string) string {
	return composeConfig(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResourceUpdateTwoRedactedFieldsConfig)
}

func testAccAwsWafv2WebACLLoggingConfiguration_updateOneRedactedField(rName string) string {
	return composeConfig(
		testAccWebACLLoggingConfigurationDependenciesConfig(rName),
		testAccWebACLLoggingConfigurationKinesisDependencyConfig(rName),
		testAccWebACLLoggingConfigurationResourceUpdateOneRedactedFieldConfig)
}
