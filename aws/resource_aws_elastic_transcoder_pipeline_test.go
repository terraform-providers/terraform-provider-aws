package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSElasticTranscoderPipeline_basic(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPipelineConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elastictranscoder", regexp.MustCompile(`pipeline/.+`)),
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

func TestAccAWSElasticTranscoderPipeline_kmsKey(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPipelineConfigKmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
					resource.TestCheckResourceAttrPair(resourceName, "aws_kms_key_arn", keyResourceName, "arn"),
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

func TestAccAWSElasticTranscoderPipeline_notifications(t *testing.T) {
	pipeline := elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderNotifications(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, &pipeline),
					testAccCheckAWSElasticTranscoderPipeline_notifications(&pipeline, []string{"warning", "completed"}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// update and check that we have 1 less notification
			{
				Config: awsElasticTranscoderNotifications_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, &pipeline),
					testAccCheckAWSElasticTranscoderPipeline_notifications(&pipeline, []string{"completed"}),
				),
			},
		},
	})
}

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckAWSElasticTranscoderPipeline_notifications(
	p *elastictranscoder.Pipeline, notifications []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		var notes []string
		if aws.StringValue(p.Notifications.Completed) != "" {
			notes = append(notes, "completed")
		}
		if aws.StringValue(p.Notifications.Error) != "" {
			notes = append(notes, "error")
		}
		if aws.StringValue(p.Notifications.Progressing) != "" {
			notes = append(notes, "progressing")
		}
		if aws.StringValue(p.Notifications.Warning) != "" {
			notes = append(notes, "warning")
		}

		if len(notes) != len(notifications) {
			return fmt.Errorf("ETC notifications didn't match:\n\texpected: %#v\n\tgot: %#v\n\n", notifications, notes)
		}

		sort.Strings(notes)
		sort.Strings(notifications)

		if !reflect.DeepEqual(notes, notifications) {
			return fmt.Errorf("ETC notifications were not equal:\n\texpected: %#v\n\tgot: %#v\n\n", notifications, notes)
		}

		return nil
	}
}

func TestAccAWSElasticTranscoderPipeline_withContentConfig(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPipelineWithContentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: awsElasticTranscoderPipelineWithContentConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
				),
			},
		},
	})
}

func TestAccAWSElasticTranscoderPipeline_withPermissions(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPipelineWithPerms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
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

func TestAccAWSElasticTranscoderPipeline_disappears(t *testing.T) {
	pipeline := &elastictranscoder.Pipeline{}
	resourceName := "aws_elastictranscoder_pipeline.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticTranscoder(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckElasticTranscoderPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: awsElasticTranscoderPipelineConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticTranscoderPipelineExists(resourceName, pipeline),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsElasticTranscoderPipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSElasticTranscoderPipelineExists(n string, res *elastictranscoder.Pipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *out.Pipeline

		return nil
	}
}

func testAccCheckElasticTranscoderPipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastictranscoder_pipline" {
			continue
		}

		out, err := conn.ReadPipeline(&elastictranscoder.ReadPipelineInput{
			Id: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return fmt.Errorf("unexpected error: %w", err)
		}

		if out.Pipeline != nil && aws.StringValue(out.Pipeline.Id) == rs.Primary.ID {
			return fmt.Errorf("Elastic Transcoder Pipeline still exists")
		}
	}
	return nil
}

func testAccPreCheckAWSElasticTranscoder(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).elastictranscoderconn

	input := &elastictranscoder.ListPipelinesInput{}

	_, err := conn.ListPipelines(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func awsElasticTranscoderPipelineConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}
`, rName)
}

func awsElasticTranscoderPipelineConfigKmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket    = aws_s3_bucket.test.bucket
  output_bucket   = aws_s3_bucket.test.bucket
  name            = %[1]q
  role            = aws_iam_role.test.arn
  aws_kms_key_arn = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}
`, rName)
}

func awsElasticTranscoderPipelineWithContentConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.content_bucket.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "%[1]s-content"
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
  acl    = "private"
}
`, rName)
}

func awsElasticTranscoderPipelineWithContentConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.input_bucket.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.content_bucket.bucket
    storage_class = "Standard"
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.thumb_bucket.bucket
    storage_class = "Standard"
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "content_bucket" {
  bucket = "%[1]s-content"
  acl    = "private"
}

resource "aws_s3_bucket" "input_bucket" {
  bucket = "%[1]s-input"
  acl    = "private"
}

resource "aws_s3_bucket" "thumb_bucket" {
  bucket = "%[1]s-thumb"
  acl    = "private"
}
`, rName)
}

func awsElasticTranscoderPipelineWithPerms(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket = aws_s3_bucket.test.bucket
  name         = %[1]q
  role         = aws_iam_role.test.arn

  content_config {
    bucket        = aws_s3_bucket.test.bucket
    storage_class = "Standard"
  }

  content_config_permissions {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }

  thumbnail_config {
    bucket        = aws_s3_bucket.test.bucket
    storage_class = "Standard"
  }

  thumbnail_config_permissions {
    grantee_type = "Group"
    grantee      = "AuthenticatedUsers"
    access       = ["FullControl"]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}
`, rName)
}

func awsElasticTranscoderNotifications(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn

  notifications {
    completed = aws_sns_topic.test.arn
    warning   = aws_sns_topic.test.arn
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "AWSAccountTopicAccess",
  "Statement": [
    {
      "Sid": "*",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sns:Publish",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func awsElasticTranscoderNotifications_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_pipeline" "test" {
  input_bucket  = aws_s3_bucket.test.bucket
  output_bucket = aws_s3_bucket.test.bucket
  name          = %[1]q
  role          = aws_iam_role.test.arn

  notifications {
    completed = aws_sns_topic.test.arn
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_sns_topic" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "AWSAccountTopicAccess",
  "Statement": [
    {
      "Sid": "*",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sns:Publish",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}
