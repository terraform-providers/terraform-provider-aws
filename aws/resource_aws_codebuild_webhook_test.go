package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeBuildWebhook_Bitbucket(t *testing.T) {
	var webhook codebuild.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildWebhookConfig_Bitbucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "secret", ""),
					resource.TestMatchResourceAttr(resourceName, "url", regexp.MustCompile(`^https://`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccAWSCodeBuildWebhook_GitHub(t *testing.T) {
	var webhook codebuild.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildWebhookConfig_GitHub(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", ""),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexp.MustCompile(`^https://`)),
					resource.TestCheckResourceAttr(resourceName, "secret", ""),
					resource.TestMatchResourceAttr(resourceName, "url", regexp.MustCompile(`^https://`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccAWSCodeBuildWebhook_GitHubEnterprise(t *testing.T) {
	var webhook codebuild.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildWebhookConfig_GitHubEnterprise(rName, "dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "dev"),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexp.MustCompile(`^https://`)),
					resource.TestMatchResourceAttr(resourceName, "secret", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
			{
				Config: testAccAWSCodeBuildWebhookConfig_GitHubEnterprise(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
					resource.TestCheckResourceAttr(resourceName, "project_name", rName),
					resource.TestMatchResourceAttr(resourceName, "payload_url", regexp.MustCompile(`^https://`)),
					resource.TestMatchResourceAttr(resourceName, "secret", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttr(resourceName, "url", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccAWSCodeBuildWebhook_BranchFilter(t *testing.T) {
	var webhook codebuild.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildWebhookConfig_BranchFilter(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "master"),
				),
			},
			{
				Config: testAccAWSCodeBuildWebhookConfig_BranchFilter(rName, "dev"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_filter", "dev"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func TestAccAWSCodeBuildWebhook_FilterGroup(t *testing.T) {
	var webhook codebuild.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildWebhookConfig_FilterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists(resourceName, &webhook),
					testAccCheckAWSCodeBuildWebhookFilter(&webhook, [][]*codebuild.WebhookFilter{
						{
							{
								Type:                  aws.String("EVENT"),
								Pattern:               aws.String("PUSH"),
								ExcludeMatchedPattern: aws.Bool(false),
							},
							{
								Type:                  aws.String("HEAD_REF"),
								Pattern:               aws.String("refs/heads/master"),
								ExcludeMatchedPattern: aws.Bool(true),
							},
						},
						{
							{
								Type:                  aws.String("EVENT"),
								Pattern:               aws.String("PULL_REQUEST_UPDATED"),
								ExcludeMatchedPattern: aws.Bool(false),
							},
						},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func testAccCheckAWSCodeBuildWebhookFilter(webhook *codebuild.Webhook, expectedFilters [][]*codebuild.WebhookFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if webhook == nil {
			return fmt.Errorf("webhook missing")
		}

		if !reflect.DeepEqual(webhook.FilterGroups, expectedFilters) {
			return fmt.Errorf("expected webhook filter configuration (%v), got: %v", expectedFilters, webhook.FilterGroups)
		}

		return nil
	}
}

func testAccCheckAWSCodeBuildWebhookDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_webhook" {
			continue
		}

		resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if len(resp.Projects) == 0 {
			return nil
		}

		project := resp.Projects[0]
		if project.Webhook != nil && project.Webhook.Url != nil {
			return fmt.Errorf("Found CodeBuild Project %q Webhook: %s", rs.Primary.ID, project.Webhook)
		}
	}
	return nil
}

func testAccCheckAWSCodeBuildWebhookExists(name string, webhook *codebuild.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if len(resp.Projects) == 0 {
			return fmt.Errorf("CodeBuild Project %q not found", rs.Primary.ID)
		}

		project := resp.Projects[0]
		if project.Webhook == nil || aws.StringValue(project.Webhook.PayloadUrl) == "" {
			return fmt.Errorf("CodeBuild Project %q Webhook not found", rs.Primary.ID)
		}

		*webhook = *project.Webhook

		return nil
	}
}

func testAccAWSCodeBuildWebhookConfig_Bitbucket(rName string) string {
	return fmt.Sprintf(testAccAWSCodeBuildProjectConfig_Source_Type_Bitbucket(rName) + `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
}
`)
}

func testAccAWSCodeBuildWebhookConfig_GitHub(rName string) string {
	return fmt.Sprintf(testAccAWSCodeBuildProjectConfig_basic(rName) + `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name
}
`)
}

func testAccAWSCodeBuildWebhookConfig_GitHubEnterprise(rName string, branchFilter string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = aws_iam_role.test.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://example.com/terraform/acceptance-testing.git"
    type     = "GITHUB_ENTERPRISE"
  }
}

resource "aws_codebuild_webhook" "test" {
  project_name  = aws_codebuild_project.test.name
  branch_filter = "%s"
}
`, rName, branchFilter)
}

func testAccAWSCodeBuildWebhookConfig_BranchFilter(rName, branchFilter string) string {
	return fmt.Sprintf(testAccAWSCodeBuildProjectConfig_basic(rName)+`
resource "aws_codebuild_webhook" "test" {
  branch_filter = "%s"
  project_name  = aws_codebuild_project.test.name
}
`, branchFilter)
}

func testAccAWSCodeBuildWebhookConfig_FilterGroup(rName string) string {
	return fmt.Sprintf(testAccAWSCodeBuildProjectConfig_basic(rName) + `
resource "aws_codebuild_webhook" "test" {
  project_name = aws_codebuild_project.test.name

  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PUSH"
    }

    filter {
      type                    = "HEAD_REF"
      pattern                 = "refs/heads/master"
      exclude_matched_pattern = true
    }
  }

  filter_group {
    filter {
      type    = "EVENT"
      pattern = "PULL_REQUEST_UPDATED"
    }
  }
}
`)
}
