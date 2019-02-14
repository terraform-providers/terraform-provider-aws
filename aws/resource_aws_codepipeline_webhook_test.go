package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodePipelineWebhook_basic(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	resourceName := "aws_codepipeline_webhook.bar"
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists("aws_codepipeline.bar"),
					testAccCheckAWSCodePipelineWebhookExists(resourceName),
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
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	resourceName := "aws_codepipeline_webhook.bar"
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_ipAuth(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists("aws_codepipeline.bar"),
					testAccCheckAWSCodePipelineWebhookExists(resourceName),
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
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("Environment variable GITHUB_TOKEN is not set")
	}

	resourceName := "aws_codepipeline_webhook.bar"
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineWebhookConfig_unauthenticated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists("aws_codepipeline.bar"),
					testAccCheckAWSCodePipelineWebhookExists(resourceName),
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

func testAccCheckAWSCodePipelineWebhookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No webhook ARN is set as ID")
		}

		conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

		_, err := getCodePipelineWebhook(conn, rs.Primary.ID)
		return err
	}
}

func testAccAWSCodePipelineWebhookConfig_basic(rName string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, fmt.Sprintf(`
resource "aws_codepipeline_webhook" "bar" {
    name            = "test-webhook-%s" 
    authentication  = "GITHUB_HMAC" 
    target_action   = "Source"
    target_pipeline = "${aws_codepipeline.bar.name}"

    authentication_configuration {
      secret_token = "super-secret"
    }

    filter {
      json_path    = "$.ref"
      match_equals = "refs/head/{Branch}"
    }
}
`, rName))
}

func testAccAWSCodePipelineWebhookConfig_ipAuth(rName string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, fmt.Sprintf(`
resource "aws_codepipeline_webhook" "bar" {
    name            = "test-webhook-%s" 
    authentication  = "IP" 
    target_action   = "Source"
    target_pipeline = "${aws_codepipeline.bar.name}"

    authentication_configuration {
      allowed_ip_range = "0.0.0.0/0"
    }

    filter {
      json_path    = "$.ref"
      match_equals = "refs/head/{Branch}"
    }
}
`, rName))
}

func testAccAWSCodePipelineWebhookConfig_unauthenticated(rName string) string {
	return testAccAWSCodePipelineWebhookConfig_codePipeline(rName, fmt.Sprintf(`
resource "aws_codepipeline_webhook" "bar" {
    name            = "test-webhook-%s" 
    authentication  = "UNAUTHENTICATED" 
    target_action   = "Source"
    target_pipeline = "${aws_codepipeline.bar.name}"

    filter {
      json_path    = "$.ref"
      match_equals = "refs/head/{Branch}"
    }
}
`, rName))
}

func testAccAWSCodePipelineWebhookConfig_codePipeline(rName string, webhook string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "foo" {
  bucket = "tf-test-pipeline-%s"
  acl    = "private"
}

resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%s"

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

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = "${aws_iam_role.codepipeline_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.foo.arn}",
        "${aws_s3_bucket.foo.arn}/*"
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

resource "aws_codepipeline" "bar" {
  name     = "test-pipeline-%s"
  role_arn = "${aws_iam_role.codepipeline_role.arn}"

  artifact_store {
    location = "${aws_s3_bucket.foo.bucket}"
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
        Owner  = "lifesum-terraform"
        Repo   = "test"
        Branch = "master"
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

%s
`, rName, rName, rName, webhook)
}
