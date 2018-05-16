package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAppsyncDatasource_ddb(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncDatasourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDatasourceConfig_ddb(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncDatasourceExists("aws_appsync_datasource.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_datasource.test", "arn"),
				),
			},
		},
	})
}

func TestAccAwsAppsyncDatasource_es(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncDatasourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDatasourceConfig_es(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncDatasourceExists("aws_appsync_datasource.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_datasource.test", "arn"),
				),
			},
		},
	})
}

func TestAccAwsAppsyncDatasource_lambda(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncDatasourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDatasourceConfig_lambda(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncDatasourceExists("aws_appsync_datasource.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_datasource.test", "arn"),
				),
			},
		},
	})
}

func TestAccAwsAppsyncDatasource_update(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppsyncDatasourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppsyncDatasourceConfig_ddb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncDatasourceExists("aws_appsync_datasource.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_datasource.test", "arn"),
					resource.TestCheckResourceAttr("aws_appsync_datasource.test", "type", "AMAZON_DYNAMODB"),
				),
			},
			{
				Config: testAccAppsyncDatasourceConfig_update_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppsyncDatasourceExists("aws_appsync_datasource.test"),
					resource.TestCheckResourceAttrSet("aws_appsync_datasource.test", "arn"),
					resource.TestCheckResourceAttr("aws_appsync_datasource.test", "type", "AWS_LAMBDA"),
				),
			},
		},
	})
}

func testAccCheckAwsAppsyncDatasourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appsyncconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appsync_datasource" {
			continue
		}

		input := &appsync.GetDataSourceInput{
			ApiId: aws.String(rs.Primary.Attributes["api_id"]),
			Name:  aws.String(rs.Primary.Attributes["name"]),
		}

		_, err := conn.GetDataSource(input)
		if err != nil {
			if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsAppsyncDatasourceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAppsyncDatasourceConfig_ddb(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_dynamodb_table" "test" {
  name = "tf-ddb-%s"
  read_capacity  = 1
  write_capacity = 1
  hash_key = "UserId"
  attribute {
    name = "UserId"
    type = "S"
  }
}

resource "aws_iam_role" "test" {
  name = "tf-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-rolepolicy-%s"
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "dynamodb:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_dynamodb_table.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "tf_appsync_%s"
  type = "AMAZON_DYNAMODB"
  dynamodb_config {
    region = "${data.aws_region.current.name}"
    table_name = "${aws_dynamodb_table.test.name}"
  }
  service_role_arn = "${aws_iam_role.test.arn}"
}
`, rName, rName, rName, rName, rName)
}

func testAccAppsyncDatasourceConfig_es(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-es-%s"
  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role" "test" {
  name = "tf-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-rolepolicy-%s"
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "es:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_elasticsearch_domain.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "tf_appsync_%s"
  type = "AMAZON_ELASTICSEARCH"
  elasticsearch_config {
    region = "${data.aws_region.current.name}"
    endpoint = "https://${aws_elasticsearch_domain.test.endpoint}"
  }
  service_role_arn = "${aws_iam_role.test.arn}"
}
`, rName, rName, rName, rName, rName)
}

func testAccAppsyncDatasourceConfig_lambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_iam_role" "test_lambda" {
  name = "tf-lambdarole-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  function_name = "tf-lambda-%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.test_lambda.arn}"
  handler = "exports.test"
  runtime = "nodejs6.10"
}

resource "aws_iam_role" "test" {
  name = "tf-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-rolepolicy-%s"
  role = "${aws_iam_role.test.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_lambda_function.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "tf_appsync_%s"
  type = "AWS_LAMBDA"
  lambda_config {
    function_arn = "${aws_lambda_function.test.arn}"
  }
  service_role_arn = "${aws_iam_role.test.arn}"
}
`, rName, rName, rName, rName, rName, rName)
}

func testAccAppsyncDatasourceConfig_update_lambda(rName string) string {
	return fmt.Sprintf(`
resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name = "tf_appsync_%s"
}

resource "aws_iam_role" "test_lambda" {
  name = "tf-lambdarole-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  function_name = "tf-lambda-%s"
  filename = "test-fixtures/lambdatest.zip"
  role = "${aws_iam_role.test_lambda.arn}"
  handler = "exports.test"
  runtime = "nodejs6.10"
}

resource "aws_iam_role" "test_applambda" {
  name = "tf-approle-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "appsync.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test_applambda" {
  name = "tf-approlepolicy-%s"
  role = "${aws_iam_role.test_applambda.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "lambda:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_lambda_function.test.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_appsync_datasource" "test" {
  api_id = "${aws_appsync_graphql_api.test.id}"
  name = "tf_appsync_%s"
  type = "AWS_LAMBDA"
  lambda_config {
    function_arn = "${aws_lambda_function.test.arn}"
  }
  service_role_arn = "${aws_iam_role.test_applambda.arn}"
}
`, rName, rName, rName, rName, rName, rName)
}
