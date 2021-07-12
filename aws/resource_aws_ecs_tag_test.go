package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEcsTag_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEcsTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEcsTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcsTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
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

func TestAccAWSEcsTag_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEcsTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEcsTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcsTagExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcsTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/11951
func TestAccAWSEcsTag_ResourceArn_BatchComputeEnvironment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSBatch(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEcsTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEcsTagConfigResourceArnBatchComputeEnvironment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcsTagExists(resourceName),
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

func TestAccAWSEcsTag_Value(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEcsTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEcsTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcsTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEcsTagConfig(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEcsTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccEcsTagConfig(rName string, key string, value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_ecs_tag" "test" {
  resource_arn = aws_ecs_cluster.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}

func testAccEcsTagConfigResourceArnBatchComputeEnvironment(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "batch.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q
  service_role             = aws_iam_role.test.arn
  type                     = "UNMANAGED"

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_ecs_tag" "test" {
  resource_arn = aws_batch_compute_environment.test.ecs_cluster_arn
  key          = "testkey"
  value        = "testvalue"
}
`, rName)
}
