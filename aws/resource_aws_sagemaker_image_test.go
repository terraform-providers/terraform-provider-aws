package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_image", &resource.Sweeper{
		Name: "aws_sagemaker_image",
		F:    testSweepSagemakerImages,
	})
}

func testSweepSagemakerImages(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListImagesPages(&sagemaker.ListImagesInput{}, func(page *sagemaker.ListImagesOutput, lastPage bool) bool {
		for _, Image := range page.Images {
			name := aws.StringValue(Image.ImageName)

			input := &sagemaker.DeleteImageInput{
				ImageName: Image.ImageName,
			}

			log.Printf("[INFO] Deleting SageMaker Image: %s", name)
			if _, err := conn.DeleteImage(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Image (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Image sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Images: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerImage_basic(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
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

func TestAccAWSSagemakerImage_description(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccAWSSagemakerImageDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerImage_displayName(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageDisplayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", ""),
				),
			},
			{
				Config: testAccAWSSagemakerImageDisplayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerImage_tags(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerImageConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerImageConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerImage_disappears(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerImageExists(resourceName, &image),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerImageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_image" {
			continue
		}

		Image, err := finder.ImageByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(Image.ImageName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Image %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerImageExists(n string, image *sagemaker.DescribeImageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Image ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.ImageByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccAWSSagemakerImageConfigBase(rName string) string {
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
`, rName)
}

func testAccAWSSagemakerImageBasicConfig(rName string) string {
	return testAccAWSSagemakerImageConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}
`, rName)
}

func testAccAWSSagemakerImageDescription(rName string) string {
	return testAccAWSSagemakerImageConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name  = %[1]q
  role_arn    = aws_iam_role.test.arn
  description = %[1]q
}
`, rName)
}

func testAccAWSSagemakerImageDisplayName(rName string) string {
	return testAccAWSSagemakerImageConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name   = %[1]q
  role_arn     = aws_iam_role.test.arn
  display_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerImageConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerImageConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerImageConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerImageConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
