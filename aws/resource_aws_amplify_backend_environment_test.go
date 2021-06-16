package aws

import (
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

func testAccAWSAmplifyBackendEnvironment_basic(t *testing.T) {
	var env amplify.BackendEnvironment
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfigBasic(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_artifacts"),
					resource.TestCheckResourceAttr(resourceName, "environment_name", environmentName),
					resource.TestCheckResourceAttrSet(resourceName, "stack_name"),
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

func testAccAWSAmplifyBackendEnvironment_disappears(t *testing.T) {
	var env amplify.BackendEnvironment
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfigBasic(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAmplifyBackendEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAmplifyBackendEnvironment_DeploymentArtifacts_StackName(t *testing.T) {
	var env amplify.BackendEnvironment
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_backend_environment.test"

	environmentName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   testAccErrorCheck(t, amplify.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyBackendEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyBackendEnvironmentConfigDeploymentArtifactsAndStackName(rName, environmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName, &env),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/[^/]+/backendenvironments/.+`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_artifacts", rName),
					resource.TestCheckResourceAttr(resourceName, "environment_name", environmentName),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
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

func testAccCheckAWSAmplifyBackendEnvironmentExists(resourceName string, v *amplify.BackendEnvironment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Backend Environment ID is set")
		}

		appID, environmentName, err := tfamplify.BackendEnvironmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		backendEnvironment, err := finder.BackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

		if err != nil {
			return err
		}

		*v = *backendEnvironment

		return nil
	}
}

func testAccCheckAWSAmplifyBackendEnvironmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).amplifyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_backend_environment" {
			continue
		}

		appID, environmentName, err := tfamplify.BackendEnvironmentParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.BackendEnvironmentByAppIDAndEnvironmentName(conn, appID, environmentName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Backend Environment %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyBackendEnvironmentConfigBasic(rName string, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = %[2]q
}
`, rName, environmentName)
}

func testAccAWSAmplifyBackendEnvironmentConfigDeploymentArtifactsAndStackName(rName string, environmentName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_backend_environment" "test" {
  app_id           = aws_amplify_app.test.id
  environment_name = %[2]q

  deployment_artifacts = %[1]q
  stack_name           = %[1]q
}
`, rName, environmentName)
}
