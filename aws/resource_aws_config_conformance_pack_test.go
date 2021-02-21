package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccConfigConformancePack_basic(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConfigConformancePack_forceNew(t *testing.T) {
	var before, after configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-update")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &before),
				),
			},
			{
				Config: testAccConfigConformancePackBasicConfig(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &after),
					testAccCheckConfigConformancePackRecreated(&before, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rNameUpdated))),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConfigConformancePack_disappears(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConfigConformancePack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigConformancePack_inputParameters(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackInputParameterConfig(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey",
						"parameter_value": "TestValue",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func testAccConfigConformancePack_S3Delivery(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackS3DeliveryConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", rName),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func testAccConfigConformancePack_S3Template(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackS3TemplateConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_s3_uri"},
			},
		},
	})
}

func testAccConfigConformancePack_updateInputParameters(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackInputParameterConfig(rName, "TestKey", "TestValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigConformancePackUpdateInputParameterConfig(rName, "TestKey1", "TestKey2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey1",
						"parameter_value": "TestValue1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "input_parameter.*", map[string]string{
						"parameter_name":  "TestKey2",
						"parameter_value": "TestValue2",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
			{
				Config: testAccConfigConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
		},
	})
}

func testAccConfigConformancePack_updateS3Delivery(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackS3DeliveryConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigConformancePackS3DeliveryConfig(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", bucketName),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_body"},
			},
		},
	})
}

func testAccConfigConformancePack_updateS3Template(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackS3TemplateConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigConformancePackS3TemplateConfig(rName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"template_s3_uri"},
			},
		},
	})
}

func testAccConfigConformancePack_updateTemplateBody(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
				),
			},
			{
				Config: testAccConfigConformancePackUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
				},
			},
		},
	})
}

func testAccConfigConformancePack_S3TemplateAndTemplateBody(t *testing.T) {
	var pack configservice.ConformancePackDetail
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_conformance_pack.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConfigConformancePackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConformancePackS3TemplateAndTemplateBodyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigConformancePackExists(resourceName, &pack),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(fmt.Sprintf("conformance-pack/%s/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_bucket", ""),
					resource.TestCheckResourceAttr(resourceName, "delivery_s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "input_parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_body",
					"template_s3_uri",
				},
			},
		},
	})
}

func testAccCheckConfigConformancePackDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_conformance_pack" {
			continue
		}

		pack, err := configDescribeConformancePack(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing Config Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack != nil {
			return fmt.Errorf("Config Conformance Pack (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckConfigConformancePackExists(resourceName string, detail *configservice.ConformancePackDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not Found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).configconn

		pack, err := configDescribeConformancePack(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error describing Config Conformance Pack (%s): %w", rs.Primary.ID, err)
		}

		if pack == nil {
			return fmt.Errorf("Config Conformance Pack (%s) not found", rs.Primary.ID)
		}

		*detail = *pack

		return nil
	}
}

func testAccCheckConfigConformancePackRecreated(before, after *configservice.ConformancePackDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.ConformancePackArn) == aws.StringValue(after.ConformancePackArn) {
			return fmt.Errorf("AWS Config Conformance Pack was not recreated")
		}
		return nil
	}
}

func testAccConfigConformancePackConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_config_configuration_recorder" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]
  name       = %[1]q
  role_arn   = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRole"
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccConfigConformancePackBasicConfig(rName string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test]
  name          = %q
  template_body = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}

func testAccConfigConformancePackUpdateConfig(rName string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on    = [aws_config_configuration_recorder.test]
  name          = %q
  template_body = <<EOT
Resources:
  IAMGroupHasUsersCheck:
    Properties:
      ConfigRuleName: IAMGroupHasUsersCheck
      Source:
        Owner: AWS
        SourceIdentifier: IAM_GROUP_HAS_USERS_CHECK
    Type: AWS::Config::ConfigRule
EOT
}
`, rName))
}

func testAccConfigConformancePackInputParameterConfig(rName, pName, pValue string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on = [aws_config_configuration_recorder.test]
  name       = %q

  input_parameter {
    parameter_name  = %[2]q
    parameter_value = %q
  }

  template_body = <<EOT
Parameters:
  %[2]s:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName, pName, pValue))
}

func testAccConfigConformancePackUpdateInputParameterConfig(rName, pName1, pName2 string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_config_conformance_pack" "test" {
  depends_on = [
  aws_config_configuration_recorder.test]
  name = %[1]q

  input_parameter {
    parameter_name  = %[2]q
    parameter_value = "TestValue1"
  }

  input_parameter {
    parameter_name  = %[3]q
    parameter_value = "TestValue2"
  }

  template_body = <<EOT
Parameters:
  %[2]s:
    Type: String
  %[3]s:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, rName, pName1, pName2))
}

func testAccConfigConformancePackS3DeliveryConfig(rName, bucketName string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_config_conformance_pack" "test" {
  depends_on             = [aws_config_configuration_recorder.test]
  name                   = %[2]q
  delivery_s3_bucket     = aws_s3_bucket.test.bucket
  delivery_s3_key_prefix = %[1]q
  template_body          = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
`, bucketName, rName))
}

func testAccConfigConformancePackS3TemplateConfig(rName, bucketName string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}

resource "aws_config_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test]
  name            = %q
  template_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_bucket_object.test.id}"
}
`, bucketName, rName))
}

func testAccConfigConformancePackS3TemplateAndTemplateBodyConfig(rName string) string {
	return composeConfig(testAccConfigConformancePackConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = %[1]q
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}

resource "aws_config_conformance_pack" "test" {
  depends_on      = [aws_config_configuration_recorder.test]
  name            = %[1]q
  template_body   = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
  template_s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_bucket_object.test.id}"
}
`, rName))
}
