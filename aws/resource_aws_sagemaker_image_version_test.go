package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func TestAccAWSSagemakerImageVersion_basic(t *testing.T) {

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageVersionBasicConfig(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageVersionExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					testAccCheckResourceAttrRegionalARN(resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
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

func TestAccAWSSagemakerImageVersion_disappears(t *testing.T) {

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageVersionBasicConfig(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageVersionExists(resourceName, &image),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerImageVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSagemakerImageVersion_disappears_image(t *testing.T) {

	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageVersionBasicConfig(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageVersionExists(resourceName, &image),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerImage(), "aws_sagemaker_image.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerImageVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_image_version" {
			continue
		}

		imageVersion, err := finder.ImageVersionByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(imageVersion.ImageVersionArn) == rs.Primary.Attributes["arn"] {
			return fmt.Errorf("SageMaker Image Version %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerImageVersionExists(n string, image *sagemaker.DescribeImageVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Image ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.ImageVersionByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccAWSSagemakerImageVersionBasicConfig(rName, baseImage string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}
`, rName, baseImage)
}
