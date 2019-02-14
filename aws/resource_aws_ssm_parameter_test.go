package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMParameter_importBasic(t *testing.T) {
	resourceName := "aws_ssm_parameter.foo"
	randName := acctest.RandString(5)
	randValue := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(randName, "String", randValue),
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

func TestAccAWSSSMParameter_basic(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					resource.TestMatchResourceAttr("aws_ssm_parameter.foo", "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:parameter/%s$", name))),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "value", "bar"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "type", "String"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_disappears(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					testAccCheckAWSSSMParameterDisappears(&param),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMParameter_update(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "bar"),
			},
			{
				Config: testAccAWSSSMParameterBasicConfigOverwrite(name, "String", "baz1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "value", "baz1"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "type", "String"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_updateDescription(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfigOverwrite(name, "String", "bar"),
			},
			{
				Config: testAccAWSSSMParameterBasicConfigOverwriteWithoutDescription(name, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "description", ""),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_changeNameForcesNew(t *testing.T) {
	var beforeParam, afterParam ssm.Parameter
	before := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))
	after := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(before, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &beforeParam),
				),
			},
			{
				Config: testAccAWSSSMParameterBasicConfig(after, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &afterParam),
					testAccCheckAWSSSMParameterRecreated(t, &beforeParam, &afterParam),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_fullPath(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("/path/%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "String", "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					resource.TestMatchResourceAttr("aws_ssm_parameter.foo", "arn",
						regexp.MustCompile(fmt.Sprintf("^arn:aws:ssm:[a-z0-9-]+:[0-9]{12}:parameter%s$", name))),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "value", "bar"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "type", "String"),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_secure(t *testing.T) {
	var param ssm.Parameter
	name := fmt.Sprintf("%s_%s", t.Name(), acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterBasicConfig(name, "SecureString", "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "value", "secret"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "type", "SecureString"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.foo", "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_secure_with_key(t *testing.T) {
	var param ssm.Parameter
	randString := acctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterSecureConfigWithKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.secret_foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "value", "secret"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "type", "SecureString"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "key_id", "alias/"+randString),
				),
			},
		},
	})
}

func TestAccAWSSSMParameter_secure_keyUpdate(t *testing.T) {
	var param ssm.Parameter
	randString := acctest.RandString(10)
	name := fmt.Sprintf("%s_%s", t.Name(), randString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMParameterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMParameterSecureConfig(name, "secret"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.secret_foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "value", "secret"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "type", "SecureString"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "key_id", "alias/aws/ssm"), // Default SSM key id
				),
			},
			{
				Config: testAccAWSSSMParameterSecureConfigWithKey(name, "secret", randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMParameterExists("aws_ssm_parameter.secret_foo", &param),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "value", "secret"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "type", "SecureString"),
					resource.TestCheckResourceAttr("aws_ssm_parameter.secret_foo", "key_id", "alias/"+randString),
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
resource "aws_ssm_parameter" "foo" {
  name  = "%s"
  type  = "%s"
  value = "%s"
}
`, rName, pType, value)
}

func testAccAWSSSMParameterBasicConfigOverwrite(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "foo" {
  name  = "test_parameter-%[1]s"
  description  = "description for parameter %[1]s"
  type  = "%[2]s"
  value = "%[3]s"
  overwrite = true
}
`, rName, pType, value)
}

func testAccAWSSSMParameterBasicConfigOverwriteWithoutDescription(rName, pType, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "foo" {
  name  = "test_parameter-%[1]s"
  type  = "%[2]s"
  value = "%[3]s"
  overwrite = true
}
`, rName, pType, value)
}

func testAccAWSSSMParameterSecureConfig(rName string, value string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_foo" {
  name  = "test_secure_parameter-%[1]s"
  description  = "description for parameter %[1]s"
  type  = "SecureString"
  value = "%[2]s"
}
`, rName, value)
}

func testAccAWSSSMParameterSecureConfigWithKey(rName string, value string, keyAlias string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "secret_foo" {
  name  = "test_secure_parameter-%[1]s"
  description  = "description for parameter %[1]s"
  type  = "SecureString"
  value = "%[2]s"
  key_id = "alias/%[3]s"
  depends_on = ["aws_kms_alias.test_alias"]
}

resource "aws_kms_key" "test_key" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test_alias" {
  name          = "alias/%[3]s"
  target_key_id = "${aws_kms_key.test_key.id}"
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
