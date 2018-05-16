package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeBuildProject_basic(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(name, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "5"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_vpc(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(name,
					testAccAWSCodeBuildProjectConfig_vpcConfig("\"${aws_subnet.codebuild_subnet.id}\",\"${aws_subnet.codebuild_subnet_2.id}\""), testAccAWSCodeBuildProjectConfig_vpcResources()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "5"),
					resource.TestCheckResourceAttrSet("aws_codebuild_project.foo", "vpc_config.0.vpc_id"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "vpc_config.0.subnets.#", "2"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basic(name, testAccAWSCodeBuildProjectConfig_vpcConfig("\"${aws_subnet.codebuild_subnet.id}\""), testAccAWSCodeBuildProjectConfig_vpcResources()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "5"),
					resource.TestCheckResourceAttrSet("aws_codebuild_project.foo", "vpc_config.0.vpc_id"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "vpc_config.0.subnets.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basicUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "50"),
					resource.TestCheckNoResourceAttr("aws_codebuild_project.foo", "vpc_config"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_cache(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_cache(name, testAccAWSCodeBuildProjectConfig_cacheConfig("S3", "")),
				ExpectError: regexp.MustCompile(`cache location is required when cache type is "S3"`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_cache(name, testAccAWSCodeBuildProjectConfig_cacheConfig("NO_CACHE", "")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_cache(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.type", "NO_CACHE"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_cache(name, testAccAWSCodeBuildProjectConfig_cacheConfig("S3", "some-bucket")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.type", "S3"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.location", "some-bucket"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_cache(name, testAccAWSCodeBuildProjectConfig_cacheConfig("S3", "some-new-bucket")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.type", "S3"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.location", "some-new-bucket"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_cache(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.#", "1"),
					resource.TestCheckResourceAttr("aws_codebuild_project.foo", "cache.0.type", "NO_CACHE"),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_sourceAuth(t *testing.T) {
	authResource := "FAKERESOURCE1"
	authType := "OAUTH"
	name := acctest.RandString(10)
	resourceName := "aws_codebuild_project.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCodeBuildProjectConfig_sourceAuth(name, authResource, "INVALID"),
				ExpectError: regexp.MustCompile(`expected source.0.auth.0.type to be one of`),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_sourceAuth(name, authResource, authType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "source.1060593600.auth.2706882902.resource", authResource),
					resource.TestCheckResourceAttr(resourceName, "source.1060593600.auth.2706882902.type", authType),
				),
			},
		},
	})
}

func TestAccAWSCodeBuildProject_default_build_timeout(t *testing.T) {
	name := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeBuildProjectConfig_default_timeout(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "60"),
				),
			},
			{
				Config: testAccAWSCodeBuildProjectConfig_basicUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildProjectExists("aws_codebuild_project.foo"),
					resource.TestCheckResourceAttr(
						"aws_codebuild_project.foo", "build_timeout", "50"),
				),
			},
		},
	})
}

func longTestData() string {
	data := `
	test-test-test-test-test-test-test-test-test-test-
	test-test-test-test-test-test-test-test-test-test-
	test-test-test-test-test-test-test-test-test-test-
	test-test-test-test-test-test-test-test-test-test-
	test-test-test-test-test-test-test-test-test-test-
	test-test-test-test-test-test-test-test-test-test-
	`

	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, data)
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
		{Value: longTestData(), ErrCount: 1},
	}

	for _, tc := range cases {
		_, errors := validateAwsCodeBuildProjectName(tc.Value, "aws_codebuild_project")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS CodeBuild project name to trigger a validation error - %s", errors)
		}
	}
}

func testAccCheckAWSCodeBuildProjectExists(n string) resource.TestCheckFunc {
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

	return fmt.Errorf("Default error in CodeBuild Test")
}

func testAccAWSCodeBuildProjectConfig_basic(rName, vpcConfig, vpcResources string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-%s"
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

resource "aws_iam_policy" "codebuild_policy" {
  name        = "codebuild-policy-%s"
  path        = "/service-role/"
  description = "Policy used in trust relationship with CodeBuild"
  policy      = <<POLICY
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
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment-%s"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name          = "test-project-%s"
  description   = "test_codebuild_project"
  build_timeout = "5"
  service_role  = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable = {
      "name"  = "SOME_KEY"
      "value" = "SOME_VALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  tags {
    "Environment" = "Test"
  }
	%s
}
%s
`, rName, rName, rName, rName, vpcConfig, vpcResources)
}

func testAccAWSCodeBuildProjectConfig_basicUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-%s"
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

resource "aws_iam_policy" "codebuild_policy" {
    name        = "codebuild-policy-%s"
    path        = "/service-role/"
    description = "Policy used in trust relationship with CodeBuild"
    policy      = <<POLICY
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
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment-%s"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name         = "test-project-%s"
  description  = "test_codebuild_project"
  build_timeout      = "50"
  service_role = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable = {
      "name"  = "SOME_OTHERKEY"
      "value" = "SOME_OTHERVALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  tags {
    "Environment" = "Test"
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSCodeBuildProjectConfig_cache(rName, cacheConfig string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-%s"
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

resource "aws_iam_policy" "codebuild_policy" {
  name        = "codebuild-policy-%s"
  path        = "/service-role/"
  description = "Policy used in trust relationship with CodeBuild"
  policy      = <<POLICY
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
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment-%s"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name          = "test-project-%s"
  description   = "test_codebuild_project"
  build_timeout = "5"
  service_role  = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  %s

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable = {
      "name"  = "SOME_KEY"
      "value" = "SOME_VALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  tags {
    "Environment" = "Test"
  }
}
`, rName, rName, rName, rName, cacheConfig)
}

func testAccAWSCodeBuildProjectConfig_default_timeout(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-%s"
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

resource "aws_iam_policy" "codebuild_policy" {
    name        = "codebuild-policy-%s"
    path        = "/service-role/"
    description = "Policy used in trust relationship with CodeBuild"
    policy      = <<POLICY
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
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment-%s"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name         = "test-project-%s"
  description  = "test_codebuild_project"

  service_role = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable = {
      "name"  = "SOME_OTHERKEY"
      "value" = "SOME_OTHERVALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"
  }

  tags {
    "Environment" = "Test"
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSCodeBuildProjectConfig_sourceAuth(rName, authResource, authType string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codebuild_role" {
  name = "codebuild-role-%[1]s"
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

resource "aws_iam_policy" "codebuild_policy" {
    name        = "codebuild-policy-%[1]s"
    path        = "/service-role/"
    description = "Policy used in trust relationship with CodeBuild"
    policy      = <<POLICY
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
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "codebuild_policy_attachment" {
  name       = "codebuild-policy-attachment-%[1]s"
  policy_arn = "${aws_iam_policy.codebuild_policy.arn}"
  roles      = ["${aws_iam_role.codebuild_role.id}"]
}

resource "aws_codebuild_project" "foo" {
  name         = "test-project-%[1]s"
  description  = "test_codebuild_project"
  build_timeout      = "5"
  service_role = "${aws_iam_role.codebuild_role.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "2"
    type         = "LINUX_CONTAINER"

    environment_variable = {
      "name"  = "SOME_KEY"
      "value" = "SOME_VALUE"
    }
  }

  source {
    type     = "GITHUB"
    location = "https://github.com/hashicorp/packer.git"

    auth {
      resource = "%[2]s"
      type     = "%[3]s"
    }
  }

  tags {
    "Environment" = "Test"
  }
}
`, rName, authResource, authType)
}

func testAccAWSCodeBuildProjectConfig_vpcResources() string {
	return fmt.Sprintf(`
	resource "aws_vpc" "codebuild_vpc" {
		cidr_block = "10.0.0.0/16"
	}

	resource "aws_subnet" "codebuild_subnet" {
		vpc_id     = "${aws_vpc.codebuild_vpc.id}"
		cidr_block = "10.0.0.0/24"
		tags {
			Name = "tf-acc-codebuild-project-1"
		}
	}

	resource "aws_subnet" "codebuild_subnet_2" {
		vpc_id     = "${aws_vpc.codebuild_vpc.id}"
		cidr_block = "10.0.1.0/24"
		tags {
			Name = "tf-acc-codebuild-project-2"
		}
	}


	resource "aws_security_group" "codebuild_security_group" {
		vpc_id = "${aws_vpc.codebuild_vpc.id}"
	}
`)
}

func testAccAWSCodeBuildProjectConfig_vpcConfig(subnets string) string {
	return fmt.Sprintf(`
  vpc_config {
    vpc_id = "${aws_vpc.codebuild_vpc.id}"

    subnets = [ %s ]

    security_group_ids = [
      "${aws_security_group.codebuild_security_group.id}"
    ]
  }
`, subnets)
}

func testAccAWSCodeBuildProjectConfig_cacheConfig(cacheType, cacheLocation string) string {
	return fmt.Sprintf(`
  cache {
    type     = "%s"
    location = "%s"
  }
`, cacheType, cacheLocation)
}
