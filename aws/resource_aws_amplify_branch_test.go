package aws

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfamplify "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func testAccAWSAmplifyBranch_basic(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/branches/.+`)),
					resource.TestCheckResourceAttr(resourceName, "associated_resources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "backend_environment_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
					resource.TestCheckResourceAttr(resourceName, "custom_domains.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "framework", ""),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", ""),
					resource.TestCheckResourceAttr(resourceName, "source_branch", ""),
					resource.TestCheckResourceAttr(resourceName, "stage", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccAWSAmplifyBranch_disappears(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAmplifyBranch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAmplifyBranch_Tags(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSAmplifyBranch_BasicAuthCredentials(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	credentials1 := base64.StdEncoding.EncodeToString([]byte("username1:password1"))
	credentials2 := base64.StdEncoding.EncodeToString([]byte("username2:password2"))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, credentials1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials1),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, credentials2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", credentials2),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "true"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					// Clearing basic_auth_credentials not reflected in API.
					// resource.TestCheckResourceAttr(resourceName, "basic_auth_credentials", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_basic_auth", "false"),
				),
			},
		},
	})
}

func testAccAWSAmplifyBranch_EnvironmentVariables(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_branch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigEnvironmentVariablesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR1", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.ENVVAR2", "2"),
				),
			},
			{
				Config: testAccAWSAmplifyBranchConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttr(resourceName, "environment_variables.%", "0"),
				),
			},
		},
	})
}

func testAccAWSAmplifyBranch_OptionalArguments(t *testing.T) {
	var branch amplify.Branch
	rName := acctest.RandomWithPrefix("tf-acc-test")
	environmentName := acctest.RandStringFromCharSet(9, acctest.CharSetAlpha)
	resourceName := "aws_amplify_branch.test"
	backendEnvironment1ResourceName := "aws_amplify_backend_environment.test1"
	backendEnvironment2ResourceName := "aws_amplify_backend_environment.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBranchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBranchConfigOptionalArguments(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "testdisplayname1"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "false"),
					resource.TestCheckResourceAttr(resourceName, "framework", "React"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr1"),
					resource.TestCheckResourceAttr(resourceName, "stage", "DEVELOPMENT"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyBranchConfigOptionalArgumentsUpdated(rName, environmentName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyBranchExists(resourceName, &branch),
					resource.TestCheckResourceAttrPair(resourceName, "backend_environment_arn", backendEnvironment2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription2"),
					resource.TestCheckResourceAttr(resourceName, "display_name", "testdisplayname2"),
					resource.TestCheckResourceAttr(resourceName, "enable_auto_build", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_notification", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_performance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_pull_request_preview", "true"),
					resource.TestCheckResourceAttr(resourceName, "framework", "Angular"),
					resource.TestCheckResourceAttr(resourceName, "pull_request_environment_name", "testpr2"),
					resource.TestCheckResourceAttr(resourceName, "stage", "EXPERIMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "ttl", "15"),
				),
			},
		},
	})
}

func testAccCheckAWSAmplifyBranchExists(resourceName string, v *amplify.Branch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Branch ID is set")
		}

		appID, branchName, err := tfamplify.BranchParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		branch, err := finder.BranchByAppIDAndBranchName(conn, appID, branchName)

		if err != nil {
			return err
		}

		*v = *branch

		return nil
	}
}

func testAccCheckAWSAmplifyBranchDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_branch" {
			continue
		}

		appID, branchName, err := tfamplify.BranchParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.BranchByAppIDAndBranchName(conn, appID, branchName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Branch %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyBranchConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}
`, rName)
}

func testAccAWSAmplifyBranchConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSAmplifyBranchConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSAmplifyBranchConfigBasicAuthCredentials(rName, basicAuthCredentials string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  basic_auth_credentials = %[2]q
  enable_basic_auth      = true
}
`, rName, basicAuthCredentials)
}

func testAccAWSAmplifyBranchConfigEnvironmentVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "1"
  }
}
`, rName)
}

func testAccAWSAmplifyBranchConfigEnvironmentVariablesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  environment_variables = {
    ENVVAR1 = "2",
    ENVVAR2 = "2"
  }
}
`, rName)
}

func testAccAWSAmplifyBranchConfigOptionalArguments(rName, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test1" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sa"
}

resource "aws_amplify_backend_environment" "test2" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sb"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  backend_environment_arn       = aws_amplify_backend_environment.test1.arn
  description                   = "testdescription1"
  display_name                  = "testdisplayname1"
  enable_auto_build             = false
  enable_notification           = true
  enable_performance_mode       = true
  enable_pull_request_preview   = false
  framework                     = "React"
  pull_request_environment_name = "testpr1"
  stage                         = "DEVELOPMENT"
  ttl                           = "10"
}
`, rName, environmentName)
}

func testAccAWSAmplifyBranchConfigOptionalArgumentsUpdated(rName, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test1" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sa"
}

resource "aws_amplify_backend_environment" "test2" {
  app_id           = aws_amplify_app.test.id
  environment_name = "%[2]sb"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q

  backend_environment_arn       = aws_amplify_backend_environment.test2.arn
  description                   = "testdescription2"
  display_name                  = "testdisplayname2"
  enable_auto_build             = true
  enable_notification           = false
  enable_performance_mode       = true
  enable_pull_request_preview   = true
  framework                     = "Angular"
  pull_request_environment_name = "testpr2"
  stage                         = "EXPERIMENTAL"
  ttl                           = "15"
}
`, rName, environmentName)
}
