package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual Bitbucket OAuth dance for this
// functionality to work. Additionally, the Bitbucket user that the Terraform AWS
// user logs in as must have access to the Bitbucket repository.
func testAccAWSCodeBuildBitbucketSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_BITBUCKET_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://terraform@bitbucket.org/terraform/aws-test.git"
	}
	return sourceLocation
}

// This is used for testing aws_codebuild_webhook as well as aws_codebuild_project.
// The Terraform AWS user must have done the manual GitHub OAuth dance for this
// functionality to work. Additionally, the GitHub user that the Terraform AWS
// user logs in as must have access to the GitHub repository.
func testAccAWSCodeBuildGitHubSourceLocationFromEnv() string {
	sourceLocation := os.Getenv("AWS_CODEBUILD_GITHUB_SOURCE_LOCATION")
	if sourceLocation == "" {
		return "https://github.com/hashibot-test/aws-test.git"
	}
	return sourceLocation
}

func TestAccAWSCodeBuildProject_basic(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codebuild", fmt.Sprintf("project/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "480"),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestMatchResourceAttr(resourceName, "encryption_key", regexp.MustCompile(`^arn:[^:]+:kms:[^:]+:[^:]+:alias/aws/s3$`)),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.compute_type", "BUILD_GENERAL1_SMALL"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.type", "LINUX_CONTAINER"),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.image_pull_credentials_type", "CODEBUILD"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", "DISABLED"),
					resource.TestMatchResourceAttr(resourceName, "service_role", regexp.MustCompile(`^arn:[^:]+:iam::[^:]+:role/tf-acc-test-[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.auth.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.git_clone_depth", "0"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.insecure_ssl", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.location", "https://github.com/hashibot-test/aws-test.git"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.report_build_status", "false"),
					resource.TestCheckResourceAttr(resourceName, "source.1441597390.type", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
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

func TestAccAWSCodeBuildProject_BadgeEnabled(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_BadgeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "badge_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "badge_url", regexp.MustCompile(`\b(https?).*\b`)),
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

func TestAccAWSCodeBuildProject_BuildTimeout(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_BuildTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_BuildTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "build_timeout", "240"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_QueuedTimeout(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_QueuedTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_QueuedTimeout(rName, 240),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "queued_timeout", "240"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Cache(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_Cache(rName, "", "S3"),
				ExpectError: regexp.MustCompile(`cache location is required when cache type is "S3"`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "", "NO_CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "some-bucket", "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", "some-bucket"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Cache(rName, "some-new-bucket", "S3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.location", "some-new-bucket"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "S3"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_LocalCache(rName, "LOCAL_DOCKER_LAYER_CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "cache.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.modes.0", "LOCAL_DOCKER_LAYER_CACHE"),
					resource.TestCheckResourceAttr(resourceName, "cache.0.type", "LOCAL"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Description(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SourceVersion(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_SourceVersion(rName, "master"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source_version", "master"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_EncryptionKey(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_EncryptionKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestMatchResourceAttr(resourceName, "encryption_key", regexp.MustCompile(`.+`)),
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

func TestAccAWSCodeBuildProject_Environment_EnvironmentVariable(t *testing.T) {
	var project1, project2, project3 codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_One(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project1),
					resource.TestCheckResourceAttr(resourceName, "environment.1380979031.environment_variable.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Two(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project2),
					resource.TestCheckResourceAttr(resourceName, "environment.4178155002.environment_variable.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Zero(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project3),
					resource.TestCheckResourceAttr(resourceName, "environment.2300252877.environment_variable.#", "0"),
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

func TestAccAWSCodeBuildProject_Environment_EnvironmentVariable_Type(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, "PLAINTEXT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.4178155002.environment_variable.0.type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "environment.4178155002.environment_variable.1.type", "PLAINTEXT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, "PARAMETER_STORE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.1701234421.environment_variable.0.type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "environment.1701234421.environment_variable.1.type", "PARAMETER_STORE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Environment_Certificate(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	oName := "certificate.pem"
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_Certificate(rName, bName, oName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					testAccCheckAWSCodeBuildProjectCertificate(&project, fmt.Sprintf("%s/%s", bName, oName)),
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

func TestAccAWSCodeBuildProject_LogsConfig_CloudWatchLogs(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_CloudWatchLogs(rName, "ENABLED", "group-name", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", ""),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_CloudWatchLogs(rName, "ENABLED", "group-name", "stream-name"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.group_name", "group-name"),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.stream_name", "stream-name"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_CloudWatchLogs(rName, "DISABLED", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.cloudwatch_logs.0.status", "DISABLED"),
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

func TestAccAWSCodeBuildProject_LogsConfig_S3Logs(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_S3Logs(rName, bName, "ENABLED", bName+"/build-log", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", "ENABLED"),
					resource.TestMatchResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", regexp.MustCompile(`tf-acc-test-bucket-[0-9]+/build-log$`)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", "false"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_S3Logs(rName, bName, "ENABLED", bName+"/build-log", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", "ENABLED"),
					resource.TestMatchResourceAttr(resourceName, "logs_config.0.s3_logs.0.location", regexp.MustCompile(`tf-acc-test-bucket-[0-9]+/build-log$`)),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.encryption_disabled", "true"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_LogsConfig_S3Logs(rName, bName, "DISABLED", "", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "logs_config.0.s3_logs.0.status", "DISABLED"),
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

func TestAccAWSCodeBuildProject_Source_Auth(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_Source_Auth(rName, "FAKERESOURCE1", "INVALID"),
				ExpectError: regexp.MustCompile(`expected source.0.auth.0.type to be one of`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Auth(rName, "FAKERESOURCE1", "OAUTH"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.2706882902.resource", "FAKERESOURCE1"),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.auth.2706882902.type", "OAUTH"),
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

func TestAccAWSCodeBuildProject_Source_GitCloneDepth(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.1181740906.git_clone_depth", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.974047921.git_clone_depth", "2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_GitSubmodulesConfig(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitSubmodulesConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3389748318.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.3389748318.git_submodules_config.0.fetch_submodules", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_GitSubmodulesConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3338377709.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.3338377709.git_submodules_config.0.fetch_submodules", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondarySources_GitSubmodulesConfig(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_SecondarySources_GitSubmodulesConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.2336845252.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.2336845252.git_submodules_config.0.fetch_submodules", "true"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.2080741754.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.2080741754.git_submodules_config.0.fetch_submodules", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_SecondarySources_GitSubmodulesConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.3511868825.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.3511868825.git_submodules_config.0.fetch_submodules", "false"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.1651171204.git_submodules_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.1651171204.git_submodules_config.0.fetch_submodules", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_InsecureSSL(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.1976396802.insecure_ssl", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.insecure_ssl", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_ReportBuildStatus_Bitbucket(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_Bitbucket(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2876219937.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_Bitbucket(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3210444828.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_ReportBuildStatus_GitHub(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHub(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.4215890488.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHub(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3680505372.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_ReportBuildStatus_GitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2964899175.report_build_status", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.553628638.report_build_status", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Source_Type_Bitbucket(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_Bitbucket(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3210444828.type", "BITBUCKET"),
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

func TestAccAWSCodeBuildProject_Source_Type_CodeCommit(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3715340088.type", "CODECOMMIT"),
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

func TestAccAWSCodeBuildProject_Source_Type_CodePipeline(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_CodePipeline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2280907000.type", "CODEPIPELINE"),
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

func TestAccAWSCodeBuildProject_Source_Type_GitHubEnterprise(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_GitHubEnterprise(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.553628638.type", "GITHUB_ENTERPRISE"),
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

func TestAccAWSCodeBuildProject_Source_Type_S3(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_S3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
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

func TestAccAWSCodeBuildProject_Source_Type_NoSource(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Source_Type_NoSource(rName, "", rBuildspec),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.2726343112.type", "NO_SOURCE"),
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

func TestAccAWSCodeBuildProject_Source_Type_NoSourceInvalid(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rBuildspec := `
version: 0.2
phases:
  build:
    commands:
      - rspec hello_world_spec.rb`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_Source_Type_NoSource(rName, "", ""),
				ExpectError: regexp.MustCompile("`build_spec` must be set when source's `type` is `NO_SOURCE`"),
			},
			{
				Config:      testAccAWSCodeBuildProjectConfig_Source_Type_NoSource(rName, "location", rBuildspec),
				ExpectError: regexp.MustCompile("`location` must be empty when source's `type` is `NO_SOURCE`"),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Tags(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Tags(rName, "tag2", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Tags(rName, "tag2", "tag2value-updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value-updated"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_VpcConfig(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_VpcConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_VpcConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnets.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile(`^vpc-`)),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_WindowsContainer(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_WindowsContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.compute_type", "BUILD_GENERAL1_MEDIUM"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.environment_variable.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.image", "2"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.privileged_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.image_pull_credentials_type", "CODEBUILD"),
					resource.TestCheckResourceAttr(resourceName, "environment.2306861956.type", "WINDOWS_CONTAINER"),
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

func TestAccAWSCodeBuildProject_ARMContainer(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_ARMContainer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
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

func TestAccAWSCodeBuildProject_Artifacts_ArtifactIdentifier(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_ArtifactIdentifier(rName, bName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_ArtifactIdentifier(rName, bName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.artifact_identifier", artifactIdentifier2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_EncryptionDisabled(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_EncryptionDisabled(rName, bName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_EncryptionDisabled(rName, bName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.encryption_disabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_Location(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	bName2 := acctest.RandomWithPrefix("tf-acc-test-bucket2")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Location(rName, bName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", bName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Location(rName, bName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.location", bName2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_Name(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Name(rName, bName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Name(rName, bName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.name", name2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_NamespaceType(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_NamespaceType(rName, codebuild.ArtifactNamespaceBuildId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", codebuild.ArtifactNamespaceBuildId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_NamespaceType(rName, codebuild.ArtifactNamespaceNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.namespace_type", codebuild.ArtifactNamespaceNone),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_OverrideArtifactName(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_OverrideArtifactName(rName, bName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_OverrideArtifactName(rName, bName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.override_artifact_name", "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_Packaging(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Packaging(rName, codebuild.ArtifactPackagingZip),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", codebuild.ArtifactPackagingZip),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Packaging(rName, codebuild.ArtifactPackagingNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.packaging", codebuild.ArtifactPackagingNone),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_Path(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Path(rName, "path1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Path(rName, "path2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.path", "path2"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_Artifacts_Type(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	type1 := codebuild.ArtifactsTypeS3
	type2 := codebuild.ArtifactsTypeCodepipeline

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Type(rName, bName, type1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_Artifacts_Type(rName, bName, type2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "artifacts.0.type", type2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts(rName, bName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "2"),
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

func TestAccAWSCodeBuildProject_SecondaryArtifacts_ArtifactIdentifier(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	artifactIdentifier1 := "artifactIdentifier1"
	artifactIdentifier2 := "artifactIdentifier2"

	hash1 := artifactHash(artifactIdentifier1, "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash(artifactIdentifier2, "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName, bName, artifactIdentifier1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.artifact_identifier", hash1), artifactIdentifier1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName, bName, artifactIdentifier2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.artifact_identifier", hash2), artifactIdentifier2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_OverrideArtifactName(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	hash1 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "true", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName, bName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.override_artifact_name", hash1), "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName, bName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.override_artifact_name", hash2), "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_EncryptionDisabled(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	hash1 := artifactHash("secondaryArtifact1", "true", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName, bName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.encryption_disabled", hash1), "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName, bName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.encryption_disabled", hash2), "false"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_Location(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	bName2 := acctest.RandomWithPrefix("tf-acc-test-bucket2")
	resourceName := "aws_codebuild_project.test"

	hash1 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", bName2, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Location(rName, bName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.location", hash1), bName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Location(rName, bName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.location", hash2), bName2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_Name(t *testing.T) {
	t.Skip("Currently no solution to allow updates on name attribute")

	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	name1 := "name1"
	name2 := "name2"

	hash1 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Name(rName, bName, name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.name", hash1), name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Name(rName, bName, name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.name", hash2), name2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_NamespaceType(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	hash1 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceBuildId, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_NamespaceType(rName, codebuild.ArtifactNamespaceBuildId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.namespace_type", hash1), codebuild.ArtifactNamespaceBuildId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_NamespaceType(rName, codebuild.ArtifactNamespaceNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.namespace_type", hash2), codebuild.ArtifactNamespaceNone),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_Packaging(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	hash1 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingZip, "", codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Packaging(rName, codebuild.ArtifactPackagingZip),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.packaging", hash1), codebuild.ArtifactPackagingZip),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Packaging(rName, codebuild.ArtifactPackagingNone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.packaging", hash2), codebuild.ArtifactPackagingNone),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_Path(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	path1 := "path1"
	path2 := "path2"

	hash1 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, path1, codebuild.ArtifactsTypeS3)
	hash2 := artifactHash("secondaryArtifact1", "false", rName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, path2, codebuild.ArtifactsTypeS3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Path(rName, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.path", hash1), path1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Path(rName, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.path", hash2), path2),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_SecondaryArtifacts_Type(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	bName := acctest.RandomWithPrefix("tf-acc-test-bucket")
	resourceName := "aws_codebuild_project.test"

	type1 := codebuild.ArtifactsTypeS3

	hash1 := artifactHash("secondaryArtifact1", "false", bName, codebuild.ArtifactNamespaceNone, "false", codebuild.ArtifactPackagingNone, "", type1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Type(rName, bName, type1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "secondary_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("secondary_artifacts.%d.type", hash1), type1),
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

func TestAccAWSCodeBuildProject_SecondarySources_CodeCommit(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_SecondarySources_CodeCommit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "source.3715340088.type", "CODECOMMIT"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.493771744.source_identifier", "secondarySource1"),
					resource.TestCheckResourceAttr(resourceName, "secondary_sources.1385902896.source_identifier", "secondarySource2"),
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

func TestAWSCodeBuildProject_nameValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{Value: "_test", ErrCount: 1},
		{Value: "test", ErrCount: 0},
		{Value: "1_test", ErrCount: 0},
		{Value: "test**1", ErrCount: 1},
		{Value: acctest.RandString(256), ErrCount: 1},
	}

	for _, tc := range cases {
		_, errors := validateAwsCodeBuildProjectName(tc.Value, "aws_codebuild_project")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS CodeBuild project name to trigger a validation error - %s", errors)
		}
	}
}

func TestAccAWSCodeBuildProject_Environment_RegistryCredential(t *testing.T) {
	var project codebuild.Project
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codebuild_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCodeBuild(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_RegistryCredential1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_Environment_RegistryCredential2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName, &project),
				),
			},
		},
	})
}

func testAccCheckAWSCodeBuildProjectExists(n string, project *codebuild.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodeBuild Project ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codebuildconn

		out, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if len(out.Projects) < 1 {
			return fmt.Errorf("No project found")
		}

		*project = *out.Projects[0]

		return nil
	}
}

func testAccCheckAWSCodeBuildProjectDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_project" {
			continue
		}

		out, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if out != nil && len(out.Projects) > 0 {
			return fmt.Errorf("Expected AWS CodeBuild Project to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccCheckAWSCodeBuildProjectCertificate(project *codebuild.Project, expectedCertificate string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(project.Environment.Certificate) != expectedCertificate {
			return fmt.Errorf("CodeBuild Project certificate (%s) did not match: %s", aws.StringValue(project.Environment.Certificate), expectedCertificate)
		}
		return nil
	}
}

func testAccPreCheckAWSCodeBuild(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	input := &codebuild.BatchGetProjectsInput{
		Names: []*string{aws.String("tf-acc-test-precheck")},
	}

	_, err := conn.BatchGetProjects(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSCodeBuildProjectConfig_Base_Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = "${aws_iam_role.test.name}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeVpcs"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_basic(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name           = "%s"
  service_role   = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "%s"
    type     = "GITHUB"
  }
}
`, rName, testAccAWSCodeBuildGitHubSourceLocationFromEnv())
}

func testAccAWSCodebuildProjectConfig_BadgeEnabled(rName string, badgeEnabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  badge_enabled = %t
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, badgeEnabled, rName)
}

func testAccAWSCodeBuildProjectConfig_BuildTimeout(rName string, buildTimeout int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  build_timeout = %d
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, buildTimeout, rName)
}

func testAccAWSCodeBuildProjectConfig_QueuedTimeout(rName string, queuedTimeout int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  queued_timeout = %d
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, queuedTimeout, rName)
}

func testAccAWSCodeBuildProjectConfig_Cache(rName, cacheLocation, cacheType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    location = "%s"
    type     = "%s"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, cacheLocation, cacheType)
}

func testAccAWSCodeBuildProjectConfig_LocalCache(rName, modeType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
		type     = "LOCAL"
		modes    = ["%s"]
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, modeType)
}

func testAccAWSCodeBuildProjectConfig_Description(rName, description string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  description  = "%s"
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, description, rName)
}

func testAccAWSCodeBuildProjectConfig_SourceVersion(rName, sourceVersion string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name           = "%s"
  service_role   = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source_version = "%s"
  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, sourceVersion)
}

func testAccAWSCodeBuildProjectConfig_EncryptionKey(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_codebuild_project" "test" {
  encryption_key = "${aws_kms_key.test.arn}"
  name           = "%s"
  service_role   = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_One(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = "SOME_KEY"
      value = "SOME_VALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Two(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = "SOME_KEY"
      value = "SOME_VALUE"
    }

    environment_variable {
      name  = "SOME_KEY2"
      value = "SOME_VALUE2"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Zero(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_EnvironmentVariable_Type(rName, environmentVariableType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable {
      name  = "SOME_KEY"
      value = "SOME_VALUE"
    }

    environment_variable {
      name  = "SOME_KEY2"
      value = "SOME_VALUE2"
      type  = "%s"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, environmentVariableType)
}

func testAccAWSCodeBuildProjectConfig_Environment_Certificate(rName string, bName string, oName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_s3_bucket_object" "test" {
  bucket  = "${aws_s3_bucket.test.bucket}"
  key     = "%s"
  content = "foo"
}

resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
    certificate  = "${aws_s3_bucket.test.bucket}/${aws_s3_bucket_object.test.key}"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, oName, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_RegistryCredential1(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = "${aws_secretsmanager_secret_version.test.arn}"
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-1"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "${jsonencode(map("username", "user", "password", "pass"))}"
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Environment_RegistryCredential2(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "2"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "SERVICE_ROLE"

    registry_credential {
      credential          = "${aws_secretsmanager_secret_version.test.arn}"
      credential_provider = "SECRETS_MANAGER"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s-2"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = "${aws_secretsmanager_secret.test.id}"
  secret_string = "${jsonencode(map("username", "user", "password", "pass"))}"
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_LogsConfig_CloudWatchLogs(rName, status, gName, sName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
	type     = "GITHUB"
  }

  logs_config {
    cloudwatch_logs {
	  status = %q
	  group_name  = %q
	  stream_name = %q
    }
  }
}
`, rName, status, gName, sName)
}

func testAccAWSCodeBuildProjectConfig_LogsConfig_S3Logs(rName, bName, status, location string, encryptionDisabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
	type     = "GITHUB"
  }

  logs_config {
    s3_logs {
	  status   = %q
	  location = %q
	  encryption_disabled = %t
    }
  }
}
`, rName, status, location, encryptionDisabled)
}

func testAccAWSCodeBuildProjectConfig_Source_Auth(rName, authResource, authType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/hashicorp/packer.git"

    auth {
      resource = "%s"
      type     = "%s"
    }
  }
}
`, rName, authResource, authType)
}

func testAccAWSCodeBuildProjectConfig_Source_GitCloneDepth(rName string, gitCloneDepth int) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    git_clone_depth = %d
    location        = "https://github.com/hashicorp/packer.git"
    type            = "GITHUB"
  }
}
`, rName, gitCloneDepth)
}

func testAccAWSCodeBuildProjectConfig_Source_GitSubmodulesConfig(rName string, fetchSubmodules bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"

    git_submodules_config {
      fetch_submodules = %t
    }
  }
}
`, rName, fetchSubmodules)
}

func testAccAWSCodeBuildProjectConfig_SecondarySources_GitSubmodulesConfig(rName string, fetchSubmodules bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%[1]s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type     = "CODECOMMIT"
    source_identifier = "secondarySource1"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }

  secondary_sources {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type     = "CODECOMMIT"
    source_identifier = "secondarySource2"

    git_submodules_config {
      fetch_submodules = %[2]t
    }
  }
}
`, rName, fetchSubmodules)
}

func testAccAWSCodeBuildProjectConfig_Source_InsecureSSL(rName string, insecureSSL bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    insecure_ssl = %t
    location     = "https://github.com/hashicorp/packer.git"
    type         = "GITHUB"
  }
}
`, rName, insecureSSL)
}

func testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_Bitbucket(rName string, reportBuildStatus bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = "https://terraform@bitbucket.org/terraform/aws-test.git"
    report_build_status = %t
    type                = "BITBUCKET"
  }
}
`, rName, reportBuildStatus)
}

func testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHub(rName string, reportBuildStatus bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = "https://github.com/hashicorp/packer.git"
    report_build_status = %t
    type                = "GITHUB"
  }
}
`, rName, reportBuildStatus)
}

func testAccAWSCodeBuildProjectConfig_Source_ReportBuildStatus_GitHubEnterprise(rName string, reportBuildStatus bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location            = "https://example.com/organization/repository.git"
    report_build_status = %t
    type                = "GITHUB_ENTERPRISE"
  }
}
`, rName, reportBuildStatus)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_Bitbucket(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = %q
    type     = "BITBUCKET"
  }
}
`, rName, testAccAWSCodeBuildBitbucketSourceLocationFromEnv())
}

func testAccAWSCodeBuildProjectConfig_Source_Type_CodeCommit(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_CodePipeline(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "CODEPIPELINE"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type = "CODEPIPELINE"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_GitHubEnterprise(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://example.com/organization/repository.git"
    type     = "GITHUB_ENTERPRISE"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_S3(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket  = "${aws_s3_bucket.test.bucket}"
  content = "test"
  key     = "test.txt"
}

resource "aws_codebuild_project" "test" {
  name         = %[1]q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "${aws_s3_bucket.test.bucket}/${aws_s3_bucket_object.test.key}"
    type     = "S3"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_Source_Type_NoSource(rName string, rLocation string, rBuildspec string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type      = "NO_SOURCE"
    location  = "%s"
    buildspec = %q
  }
}
`, rName, rLocation, rBuildspec)
}

func testAccAWSCodeBuildProjectConfig_Tags(rName, tagKey, tagValue string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  tags = {
    tag1 = "tag1value"
    %s = "%s"
  }
}
`, rName, tagKey, tagValue)
}

func testAccAWSCodeBuildProjectConfig_VpcConfig1(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidInputException: CodeBuild currently doesn't support VPC in us-west-2d, please select subnets in other availability zones.
  blacklisted_zone_ids = ["usw2-az4"]
  state                = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 1

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.0.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-codebuild-project"
  }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  vpc_config {
    security_group_ids = ["${aws_security_group.test.id}"]
    subnets            = ["${aws_subnet.test.*.id[0]}"]
    vpc_id             = "${aws_vpc.test.id}"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_VpcConfig2(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidInputException: CodeBuild currently doesn't support VPC in us-west-2d, please select subnets in other availability zones.
  blacklisted_zone_ids = ["usw2-az4"]
  state                = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-codebuild-project"
  }
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://github.com/hashicorp/packer.git"
    type     = "GITHUB"
  }

  vpc_config {
    security_group_ids = ["${aws_security_group.test.id}"]
    subnets            = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
    vpc_id             = "${aws_vpc.test.id}"
  }
}
`, rName)
}

func testAccAWSCodeBuildProjectConfig_WindowsContainer(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_MEDIUM"
    image        = "2"
    type         = "WINDOWS_CONTAINER"
  }

  source {
    location = "%s"
    type     = "GITHUB"
  }
}
`, rName, testAccAWSCodeBuildGitHubSourceLocationFromEnv())
}

func testAccAWSCodeBuildProjectConfig_ARMContainer(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = "%s"
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_LARGE"
    image        = "2"
    type         = "ARM_CONTAINER"
  }

  source {
    location = "%s"
    type     = "GITHUB"
  }
}
`, rName, testAccAWSCodeBuildGitHubSourceLocationFromEnv())
}

func testAccAWSCodebuildProjectConfig_Artifacts_ArtifactIdentifier(rName string, bName string, artifactIdentifier string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    artifact_identifier        = %[2]q
    location                   = "${aws_s3_bucket.test.bucket}"
    type                       = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, artifactIdentifier)
}

func testAccAWSCodebuildProjectConfig_Artifacts_EncryptionDisabled(rName string, bName string, encryptionDisabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    encryption_disabled = %t
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, encryptionDisabled)
}

func testAccAWSCodebuildProjectConfig_Artifacts_Location(rName, bName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location  = "${aws_s3_bucket.test.bucket}"
    type      = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodebuildProjectConfig_Artifacts_Name(rName string, bName string, name string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    name        = %[2]q
    location    = "${aws_s3_bucket.test.bucket}"
    type        = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, name)
}

func testAccAWSCodebuildProjectConfig_Artifacts_NamespaceType(rName, namespaceType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location       = "${aws_s3_bucket.test.bucket}"
    namespace_type = %[2]q
    type           = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, namespaceType)
}

func testAccAWSCodebuildProjectConfig_Artifacts_OverrideArtifactName(rName string, bName string, overrideArtifactName bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    override_artifact_name = %t
    location               = "${aws_s3_bucket.test.bucket}"
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, overrideArtifactName)
}

func testAccAWSCodebuildProjectConfig_Artifacts_Packaging(rName, packaging string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location  = "${aws_s3_bucket.test.bucket}"
    packaging = %[2]q
    type      = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, packaging)
}

func testAccAWSCodebuildProjectConfig_Artifacts_Path(rName, path string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location = "${aws_s3_bucket.test.bucket}"
    path     = %[2]q
    type     = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, path)
}

func testAccAWSCodebuildProjectConfig_Artifacts_Type(rName string, bName string, artifactType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    type        = %[2]q
    location    = "${aws_s3_bucket.test.bucket}"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = %[2]q
    location = "${aws_s3_bucket.test.bucket}/"
  }
}
`, rName, artifactType)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts(rName string, bName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact2"
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_ArtifactIdentifier(rName string, bName string, artifactIdentifier string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = %[2]q
    location               = "${aws_s3_bucket.test.bucket}"
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, artifactIdentifier)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_EncryptionDisabled(rName string, bName string, encryptionDisabled bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    encryption_disabled = %t
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, encryptionDisabled)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Location(rName, bName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location  = "${aws_s3_bucket.test.bucket}"
    type      = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Name(rName string, bName string, name string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    name                = %[2]q
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, name)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_NamespaceType(rName, namespaceType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location       = "${aws_s3_bucket.test.bucket}"
    type           = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    namespace_type      = %[2]q
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, namespaceType)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_OverrideArtifactName(rName string, bName string, overrideArtifactName bool) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = "%s"
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = "secondaryArtifact1"
    override_artifact_name = %t
    location               = "${aws_s3_bucket.test.bucket}"
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, overrideArtifactName)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Packaging(rName, packaging string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location  = "${aws_s3_bucket.test.bucket}"
    type      = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = "secondaryArtifact1"
    packaging           = %[2]q
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, packaging)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Path(rName, path string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          = %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location = "${aws_s3_bucket.test.bucket}"
    type     = "S3"
  }

  secondary_artifacts {
    artifact_identifier    = "secondaryArtifact1"
    path                   = %[2]q
    location               = "${aws_s3_bucket.test.bucket}"
    type                   = "S3"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }
}
`, rName, path)
}

func testAccAWSCodebuildProjectConfig_SecondaryArtifacts_Type(rName string, bName string, artifactType string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + testAccAWSCodeBuildProjectConfig_Base_Bucket(bName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name          =  %[1]q
  service_role  = "${aws_iam_role.test.arn}"

  artifacts {
    location            = "${aws_s3_bucket.test.bucket}"
    type                = "S3"
  }

  secondary_artifacts {
    artifact_identifier = "secondaryArtifact1"
    type                = %[2]q
    location            = "${aws_s3_bucket.test.bucket}"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }
}
`, rName, artifactType)
}

func testAccAWSCodeBuildProjectConfig_SecondarySources_CodeCommit(rName string) string {
	return testAccAWSCodeBuildProjectConfig_Base_ServiceRole(rName) + fmt.Sprintf(`
resource "aws_codebuild_project" "test" {
  name         = %q
  service_role = "${aws_iam_role.test.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"
  }

  source {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/repo-name"
    type     = "CODECOMMIT"
  }

  secondary_sources {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/second-repo-name"
    type     = "CODECOMMIT"
    source_identifier = "secondarySource1"
  }

  secondary_sources {
    location = "https://git-codecommit.region-id.amazonaws.com/v1/repos/third-repo-name"
    type     = "CODECOMMIT"
    source_identifier = "secondarySource2"
  }
}
`, rName)
}

func artifactHash(artifactIdentifier, encryptionDisabled, location, nameSpaceType, overrideArtifactName,
	packaging, path, artifactType string) int {
	buf := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s-", artifactIdentifier,
		encryptionDisabled, location, nameSpaceType, overrideArtifactName, packaging, path, artifactType)
	hash := hashcode.String(buf)
	return hash
}
