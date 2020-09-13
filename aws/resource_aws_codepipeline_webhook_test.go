package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodePipelineWebhook_basic(t *testing.T) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var v codepipeline.ListWebhookItem
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
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

func TestAccAWSCodePipelineWebhook_ipAuth(t *testing.T) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var v codepipeline.ListWebhookItem
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_ipAuth(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.allowed_ip_range", "0.0.0.0/0"),
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

func TestAccAWSCodePipelineWebhook_unauthenticated(t *testing.T) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var v codepipeline.ListWebhookItem
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_unauthenticated(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
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

func TestAccAWSCodePipelineWebhook_tags(t *testing.T) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var v1, v2, v3 codepipeline.ListWebhookItem
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfigWithTags(rName, "tag1value", "tag2value", githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccAWSCodePipelineWebhookConfigWithTags(rName, "tag1valueUpdate", "tag2valueUpdate", githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1valueUpdate"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2valueUpdate"),
					func(s *terraform.State) error {
						if aws.StringValue(v2.Url) != aws.StringValue(v1.Url) {
							return fmt.Errorf("Codepipeline webhook recreated when changing tags")
						}
						return nil
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					func(s *terraform.State) error {
						if aws.StringValue(v3.Url) != aws.StringValue(v2.Url) {
							return fmt.Errorf("Codepipeline webhook recreated when deleting tags")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccAWSCodePipelineWebhook_UpdateAuthenticationConfiguration_SecretToken(t *testing.T) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var v1, v2 codepipeline.ListWebhookItem
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codepipeline_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodePipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_basic(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "super-secret"),
				),
			},
			{
				Config: testAccAWSCodePipelineWebhookConfig_secretTokenUpdated(rName, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineWebhookExists(resourceName, &v2),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_configuration.0.secret_token", "even-more-secret"),
					func(s *terraform.State) error {
						if aws.StringValue(v2.Url) == aws.StringValue(v1.Url) {
							return fmt.Errorf("Codepipeline webhook not recreated when updating authentication_configuration.secret_token")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckAWSCodePipelineWebhookExists(n string, webhook *codepipeline.ListWebhookItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No webhook ARN is set as ID")
		}

		conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

		resp, err := getCodePipelineWebhook(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*webhook = *resp

		return nil
	}
}

func testAccAWSCodePipelineWebhookConfig_basic(rName, githubToken string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "super-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccAWSCodePipelineWebhookConfig_ipAuth(rName, githubToken string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "IP"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    allowed_ip_range = "0.0.0.0/0"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccAWSCodePipelineWebhookConfig_unauthenticated(rName, githubToken string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "UNAUTHENTICATED"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccAWSCodePipelineWebhookConfigWithTags(rName, tag1, tag2, githubToken string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "super-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }

  tags = {
    Name = %[1]q
    tag1 = %[2]q
    tag2 = %[3]q
  }
}
`, rName, tag1, tag2)
}

func testAccAWSCodePipelineWebhookConfig_secretTokenUpdated(rName, githubToken string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken) + fmt.Sprintf(`
resource "aws_codepipeline_webhook" "test" {
  name            = %[1]q
  authentication  = "GITHUB_HMAC"
  target_action   = "Source"
  target_pipeline = aws_codepipeline.test.name

  authentication_configuration {
    secret_token = "even-more-secret"
  }

  filter {
    json_path    = "$.ref"
    match_equals = "refs/head/{Branch}"
  }
}
`, rName)
}

func testAccAWSCodePipelineWebhookConfig_codePipeline(rName, githubToken string) string {
	return fmt.Sprintf(`
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
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_codepipeline" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner      = "lifesum-terraform"
        Repo       = "test"
        Branch     = "master"
        OAuthToken = %[2]q
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName, githubToken)
}
