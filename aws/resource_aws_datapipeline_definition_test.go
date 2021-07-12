package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/datapipeline"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDataPipelineDefinition_basic(t *testing.T) {
	var conf datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "cron"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(
						"aws_datapipeline_pipeline.default", "id",
						resourceName, "pipeline_id"),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.schedule_type", "cron"),
					resource.TestCheckResourceAttr(resourceName, "default.0.failure_and_rerun_mode", "none"),
					resource.TestCheckResourceAttrPair(
						"aws_iam_role.role", "arn",
						resourceName, "default.0.role"),
					resource.TestCheckResourceAttrPair(
						"aws_iam_instance_profile.resource_role", "arn",
						resourceName, "default.0.resource_role"),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "ondemand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.schedule_type", "ondemand"),
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

func TestAccAWSDataPipelineDefinition_options(t *testing.T) {
	var conf datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	bucketName1 := fmt.Sprintf("tf-test-pipeline-log-bucket-%s", acctest.RandString(5))
	bucketName2 := fmt.Sprintf("tf-test-pipeline-log-bucket-2-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfigDefaultWithOptions(rName, bucketName1, "cascade"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.failure_and_rerun_mode", "cascade"),
					resource.TestCheckResourceAttr(resourceName, "default.0.pipeline_log_uri", fmt.Sprintf("s3://%s/log_prefix", bucketName1)),
				),
			},
			{
				Config: testAccAWSDataPipelineDefinitionConfigDefaultWithOptions(rName, bucketName2, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "default.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default.0.failure_and_rerun_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "default.0.pipeline_log_uri", fmt.Sprintf("s3://%s/log_prefix", bucketName2)),
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

func TestAccAWSDataPipelineDefinition_disappears(t *testing.T) {
	var description datapipeline.PipelineDescription
	var definition datapipeline.GetPipelineDefinitionOutput
	rName := fmt.Sprintf("tf-datapipeline-%s", acctest.RandString(5))
	resourceName := "aws_datapipeline_definition.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataPipeline(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataPipelineDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataPipelineDefinitionConfig(rName, "ondemand"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataPipelinePipelineExists("aws_datapipeline_pipeline.default", &description),
					testAccCheckAWSDataPipelineDefinitionExists(resourceName, &definition),
					testAccCheckAWSDataPipelinePipelineDisappears(&description),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDataPipelineDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_definition" {
			continue
		}
		// Try to find the Pipeline
		pipelineDefinition, err := resourceAwsDataPipelineDefinitionRetrieve(rs.Primary.ID, conn)
		if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
			continue
		}

		if err != nil {
			return err
		}
		if pipelineDefinition != nil {
			return fmt.Errorf("Pipeline Definition still exists")
		}
	}

	return nil
}

func testAccCheckAWSDataPipelineDefinitionExists(n string, v *datapipeline.GetPipelineDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).datapipelineconn

		pipelineDefinition, err := resourceAwsDataPipelineDefinitionRetrieve(rs.Primary.ID, conn)

		if err != nil {
			return err
		}
		if pipelineDefinition == nil {
			return fmt.Errorf("DataPipeline Definition not found")
		}

		*v = *pipelineDefinition
		return nil
	}
}

func testAccAWSDataPipelineDefinitionConfigCommon(rName string) string {
	return fmt.Sprintf(testAccAWSDataPipelinePipelineConfig(rName)+`
resource "aws_iam_role" "role" {
  name = "tf-test-datapipeline-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "elasticmapreduce.amazonaws.com",
          "datapipeline.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
data "aws_iam_policy" "role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}
resource "aws_iam_role_policy_attachment" "role" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${data.aws_iam_policy.role.arn}"
}
resource "aws_iam_role" "resource_role" {
  name = "tf-test-datapipeline-resource-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
data "aws_iam_policy" "resource_role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}
resource "aws_iam_role_policy_attachment" "resource_role" {
  role       = "${aws_iam_role.resource_role.name}"
  policy_arn = "${data.aws_iam_policy.resource_role.arn}"
}
resource "aws_iam_instance_profile" "resource_role" {
  name = "tf-test-datapipeline-resource-role-profile-%[1]s"
  role = "${aws_iam_role.resource_role.name}"
}
`, rName)
}

func testAccAWSDataPipelineDefinitionConfig(rName, scheduleType string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`
resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"
  default {
	schedule_type = %[2]q
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }
}`, rName, scheduleType)

}

func testAccAWSDataPipelineDefinitionConfigDefaultWithOptions(rName, bucketName, failureAndRerunMode string) string {
	return fmt.Sprintf(testAccAWSDataPipelineDefinitionConfigCommon(rName)+`
resource "aws_s3_bucket" "log_bucket" {
  bucket = %[1]q
  acl    = "log-delivery-write"
		
  force_destroy = true
}
resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"
  default {
	schedule_type          = "ondemand"
	role                   = "${aws_iam_role.role.arn}"
	resource_role          = "${aws_iam_instance_profile.resource_role.arn}"
	pipeline_log_uri       = "s3://${aws_s3_bucket.log_bucket.id}/log_prefix"
	failure_and_rerun_mode = %[2]q
  }
}`, bucketName, failureAndRerunMode)

}
