package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAppFlowFlow_basic(t *testing.T) {
	var flow appflow.DescribeFlowOutput

	flowName := acctest.RandomWithPrefix("tf-acc-test-appflow-flow")
	sourceBucketName := acctest.RandomWithPrefix("tf-acc-test-source-bucket")
	destinationBucketName := acctest.RandomWithPrefix("tf-acc-test-destination-bucket")
	resourceName := "aws_appflow_flow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, appflow.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAccessAnalyzerAnalyzerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAppFlowFlowConfig_fromS3_toS3(sourceBucketName, destinationBucketName, flowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAppFlowFlowExists(resourceName, &flow),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, "description", "Flow from source S3 bucket to destination S3 bucket"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.connector_profile_name", "destination"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.connector_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.bucket_name", destinationBucketName),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.bucket_prefix", "data"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.s3_output_format_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.s3_output_format_config.0.aggregation_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.s3_output_format_config.0.aggregation_config.0.aggregation_type", "None"),
					resource.TestCheckResourceAttr(resourceName, "destination_flow_config_list.0.destination_connector_properties.0.s3.0.s3_output_format_config.0.file_type", "JSON"),
					testAccCheckResourceAttrRegionalARN(resourceName, "flow_arn", "appflow", fmt.Sprintf("flow/%s", flowName)),
					resource.TestCheckResourceAttr(resourceName, "flow_name", flowName),
					resource.TestCheckResourceAttrSet(resourceName, "flow_status"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_arn"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.connector_profile_name", "source"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.connector_type", "S3"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.source_connector_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.source_connector_properties.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.source_connector_properties.0.s3.0.bucket_name", sourceBucketName),
					resource.TestCheckResourceAttr(resourceName, "source_flow_config.0.source_connector_properties.0.s3.0.bucket_prefix", "emails"),
					resource.TestCheckResourceAttr(resourceName, "tasks.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.connector_operator.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.connector_operator.0.s3", "PROJECTION"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.source_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.source_fields.0", "email"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.task_properties.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tasks.0.task_type", "Filter"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.connector_operator.0.s3", "NO_OP"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.destination_field", "email"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.source_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.source_fields.0", "email"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.task_properties.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tasks.1.task_type", "Map"),
					resource.TestCheckResourceAttr(resourceName, "trigger_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trigger_config.0.trigger_type", "OnDemand"),
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

func testAccCheckAWSAppFlowFlowExists(n string, res *appflow.DescribeFlowOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).appflowconn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		req := &appflow.DescribeFlowInput{
			FlowName: aws.String(rs.Primary.Attributes["flow_name"]),
		}
		describe, err := conn.DescribeFlow(req)
		if err != nil {
			return err
		}

		*res = *describe

		return nil
	}
}

func testAccAWSAppFlowFlowConfig_fromS3_toS3(sourceBucketName string, destinationBucketName string, flowName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source_bucket" {
  bucket = %[1]q
  acl    = "private"
}

resource "aws_s3_bucket_policy" "source_bucket" {
  bucket = aws_s3_bucket.source_bucket.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appflow.amazonaws.com"
      },
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Resource": [
        "${aws_s3_bucket.source_bucket.arn}",
        "${aws_s3_bucket.source_bucket.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "destination_bucket" {
  bucket = %[2]q
  acl    = "private"
}

resource "aws_s3_bucket_policy" "destination_bucket" {
  bucket = aws_s3_bucket.destination_bucket.id

  policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "appflow.amazonaws.com"
      },
      "Action": [
        "s3:PutObject",
        "s3:AbortMultipartUpload",
        "s3:ListMultipartUploadParts",
        "s3:ListBucketMultipartUploads",
        "s3:GetBucketAcl",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "${aws_s3_bucket.destination_bucket.arn}",
        "${aws_s3_bucket.destination_bucket.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_appflow_flow" "test" {
  flow_name   = %[3]q
  description = "Flow from source S3 bucket to destination S3 bucket"

  destination_flow_config_list {
    connector_profile_name = "destination"
    connector_type         = "S3"
    destination_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket.destination_bucket.id
        bucket_prefix = "data"
        s3_output_format_config {
          aggregation_config {
            aggregation_type = "None"
          }
          file_type = "JSON"
        }
      }
    }
  }

  source_flow_config {
    connector_profile_name = "source"
    connector_type         = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket.source_bucket.id
        bucket_prefix = "emails"
      }
    }
  }

  tasks {
    connector_operator {
      s3 = "PROJECTION"
    }
    source_fields = [
      "email",
    ]
    task_properties = {}
    task_type       = "Filter"
  }
  tasks {
    connector_operator {
      s3 = "NO_OP"
    }
    destination_field = "email"
    source_fields = [
      "email",
    ]
    task_properties = {}
    task_type       = "Map"
  }

  trigger_config {
    trigger_type = "OnDemand"
  }

  depends_on = [
    aws_s3_bucket_policy.source_bucket,
    aws_s3_bucket_policy.destination_bucket,
  ]
}
`, sourceBucketName, destinationBucketName, flowName)
}
