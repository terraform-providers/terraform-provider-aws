package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMParameter_basic(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("parameter/%s", name)),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "tier", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttr(resourceName, "data_type", "text"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccAWSSSMParameter_Tier(t *testing.T) {
	var parameter1, parameter2, parameter3 ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterConfigTier(rName, "Advanced"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &parameter1),
					resource.TestCheckResourceAttr(resourceName, "tier", "Advanced"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterConfigTier(rName, "Standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &parameter2),
					resource.TestCheckResourceAttr(resourceName, "tier", "Standard"),
				),
			},
			{
				Config: testAccAWSSSMParameterConfigTier(rName, "Advanced"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &parameter3),
					resource.TestCheckResourceAttr(resourceName, "tier", "Advanced"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_disappears(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					testAccCheckAWSSSMParameterDisappears(&param),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMParameter_overwrite(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterBasicConfigOverwrite(name, "String", "test3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "test3"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_tags(t *testing.T) {
	var param ssm.Parameter
	rName := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterBasicConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSSMParameterBasicConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_updateType(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "SecureString", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_updateDescription(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfigOverwrite(name, "String", "test2"),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterBasicConfigOverwriteWithoutDescription(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_changeNameForcesNew(t *testing.T) {
	var beforeParam, afterParam ssm.Parameter
	before := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	after := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(before, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &beforeParam),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterBasicConfig(after, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &afterParam),
					testAccCheckAWSSSMParameterRecreated(t, &beforeParam, &afterParam),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_fullPath(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("/path/%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ssm", fmt.Sprintf("parameter%s", name)),
					resource.TestCheckResourceAttr(resourceName, "value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccAWSSSMParameter_secure(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "SecureString", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccAWSSSMParameter_DataType_AwsEc2Image(t *testing.T) {
	var param ssm.Parameter
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_parameter.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterConfigDataTypeAwsEc2Image(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "data_type", "aws:ec2:image"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccAWSSSMParameter_secure_with_key(t *testing.T) {
	var param ssm.Parameter
	randString := acctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.secret_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterSecureConfigWithKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/"+randString),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
		},
	})
}

func TestAccAWSSSMParameter_secure_keyUpdate(t *testing.T) {
	var param ssm.Parameter
	randString := acctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)
	resourceName := "aws_ssm_parameter.secret_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterSecureConfig(name, "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"overwrite"},
			},
			{
				Config: testAccAWSSSMParameterSecureConfigWithKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists(resourceName, &param),
					resource.TestCheckResourceAttr(resourceName, "value", "secret"),
					resource.TestCheckResourceAttr(resourceName, "type", "SecureString"),
					resource.TestCheckResourceAttr(resourceName, "key_id", "alias/"+randString),
				),
			},
		},
	})
}

func testAccCheckAWSSSMParameterRecreated(t *testing.T,
	before, after *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.Name == *after.Name {
			t.Fatalf("Expected change of SSM Param Names, but both were %v", *before.Name)
		}
		return nil
	}
}

func testAccCheckAWSSSMParameterExists(n string, param *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Parameter ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
			WithDecryption: aws.Bool(true),
		}

		resp, err := conn.GetParameters(paramInput)
		if err != nil {
			return err
		}

		if len(resp.Parameters) == 0 {
			return fmt.Errorf("Expected AWS SSM Parameter to be created, but wasn't found")
		}

		*param = *resp.Parameters[0]

		return nil
	}
}

func testAccCheckAWSSSMParameterDisappears(param *ssm.Parameter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		paramInput := &ssm.DeleteParameterInput{
			Name: param.Name,
		}

		_, err := conn.DeleteParameter(paramInput)

		return err
	}
}

func testAccCheckAWSSSMParameterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_parameter" {
			continue
		}

		paramInput := &ssm.GetParametersInput{
			Names: []*string{
				aws.String(rs.Primary.Attributes["name"]),
			},
		}

		resp, _ := conn.GetParameters(paramInput)

		if len(resp.Parameters) > 0 {
			return fmt.Errorf("Expected AWS SSM Parameter to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMParameterBasicConfig(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = %[2]q
  value = %[3]q
}
`, rName, pType, value)
}

func testAccAWSSSMParameterConfigTier(rName, tier string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  tier  = %[2]q
  type  = "String"
  value = "test2"
}
`, rName, tier)
}

func testAccAWSSSMParameterConfigDataTypeAwsEc2Image(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = %[1]q
  data_type = "aws:ec2:image"
  type      = "String"
  value     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
}
`, rName))
}

func testAccAWSSSMParameterBasicConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSSMParameterBasicConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSSMParameterBasicConfigOverwrite(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name        = "test_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "%[2]s"
  value       = "%[3]s"
  overwrite   = true
}
`, rName, pType, value)
}

func testAccAWSSSMParameterBasicConfigOverwriteWithoutDescription(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name      = "test_parameter-%[1]s"
  type      = "%[2]s"
  value     = "%[3]s"
  overwrite = true
}
`, rName, pType, value)
}

func testAccAWSSSMParameterSecureConfig(rName string, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_test" {
  name        = "test_secure_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "SecureString"
  value       = "%[2]s"
}
`, rName, value)
}

func testAccAWSSSMParameterSecureConfigWithKey(rName string, value string, keyAlias string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_test" {
  name        = "test_secure_parameter-%[1]s"
  description = "description for parameter %[1]s"
  type        = "SecureString"
  value       = "%[2]s"
  key_id      = "alias/%[3]s"
  depends_on  = [aws_kms_alias.test_alias]
}

resource "aws_kms_key" "test_key" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test_alias" {
  name          = "alias/%[3]s"
  target_key_id = aws_kms_key.test_key.id
}
`, rName, value, keyAlias)
}

func TestAWSSSMParameterShouldUpdate(t *testing.T) {
	data := resourceAwsSsmParameter().TestResourceData()
	failure := false

	if !shouldUpdateSsmParameter(data) {
		t.Logf("Existing resources should be overwritten if the values don't match!")
		failure = true
	}

	data.MarkNewResource()
	if shouldUpdateSsmParameter(data) {
		t.Logf("New resources must never be overwritten, this will overwrite parameters created outside of the system")
		failure = true
	}

	data = resourceAwsSsmParameter().TestResourceData()
	data.Set("overwrite", true)
	if !shouldUpdateSsmParameter(data) {
		t.Logf("Resources should always be overwritten if the user requests it")
		failure = true
	}

	data.Set("overwrite", false)
	if shouldUpdateSsmParameter(data) {
		t.Logf("Resources should never be overwritten if the user requests it")
		failure = true
	}
	if failure {
		t.Fail()
	}
}
