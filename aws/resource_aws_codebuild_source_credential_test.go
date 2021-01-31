package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeBuildSourceCredential_basic(t *testing.T) {
	var sourceCredentialsInfo codebuild.SourceCredentialsInfo
	token := acctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildSourceCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildSourceCredential_Basic("PERSONAL_ACCESS_TOKEN", "GITHUB", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/github`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				Config: testAccAWSCodeBuildSourceCredential_Basic("PERSONAL_ACCESS_TOKEN", "GITHUB_ENTERPRISE", token),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/github_enterprise`)),
					resource.TestCheckResourceAttr(resourceName, "server_type", "GITHUB_ENTERPRISE"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "PERSONAL_ACCESS_TOKEN"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "user_name"},
			},
		},
	})
}

func TestAccAWSCodeBuildSourceCredential_BasicAuth(t *testing.T) {
	var sourceCredentialsInfo codebuild.SourceCredentialsInfo
	token := acctest.RandomWithPrefix("token")
	resourceName := "aws_codebuild_source_credential.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildSourceCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildSourceCredential_BasicAuth(token, "user1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codebuild", regexp.MustCompile(`token/bitbucket`)),
					resource.TestCheckResourceAttr(resourceName, "user_name", "user1"),
					resource.TestCheckResourceAttr(resourceName, "server_type", "BITBUCKET"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "BASIC_AUTH"),
				),
			},
			{
				Config: testAccAWSCodeBuildSourceCredential_BasicAuth(token, "user2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildSourceCredentialExists(resourceName, &sourceCredentialsInfo),
					resource.TestCheckResourceAttr(resourceName, "user_name", "user2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "user_name"},
			},
		},
	})
}

func testAccCheckAWSCodeBuildSourceCredentialDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_source_credential" {
			continue
		}

		resp, err := conn.ListSourceCredentials(&codebuild.ListSourceCredentialsInput{})
		if err != nil {
			return err
		}

		if len(resp.SourceCredentialsInfos) == 0 {
			return nil
		}

		for _, sourceCredentialsInfo := range resp.SourceCredentialsInfos {
			if rs.Primary.ID == aws.StringValue(sourceCredentialsInfo.Arn) {
				return fmt.Errorf("Found Source Credential %s", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAWSCodeBuildSourceCredentialExists(name string, sourceCredential *codebuild.SourceCredentialsInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		resp, err := conn.ListSourceCredentials(&codebuild.ListSourceCredentialsInput{})
		if err != nil {
			return err
		}

		if len(resp.SourceCredentialsInfos) == 0 {
			return fmt.Errorf("Source Credential %s not found", rs.Primary.ID)
		}

		for _, sourceCredentialsInfo := range resp.SourceCredentialsInfos {
			if rs.Primary.ID == aws.StringValue(sourceCredentialsInfo.Arn) {
				*sourceCredential = *sourceCredentialsInfo
				return nil
			}
		}

		return fmt.Errorf("Source Credential %s not found", rs.Primary.ID)
	}
}

func testAccAWSCodeBuildSourceCredential_Basic(authType, serverType, token string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "%s"
  server_type = "%s"
  token       = "%s"
}
`, authType, serverType, token)
}

func testAccAWSCodeBuildSourceCredential_BasicAuth(token, userName string) string {
	return fmt.Sprintf(`
resource "aws_codebuild_source_credential" "test" {
  auth_type   = "BASIC_AUTH"
  server_type = "BITBUCKET"
  token       = "%s"
  user_name   = "%s"
}
`, token, userName)
}
