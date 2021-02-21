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
	resource.AddTestSweepers("aws_sagemaker_model_package_group", &resource.Sweeper{
		Name: "aws_sagemaker_model_package_group",
		F:    testSweepSagemakerModelPackageGroups,
	})
}

func testSweepSagemakerModelPackageGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListModelPackageGroupsPages(&sagemaker.ListModelPackageGroupsInput{}, func(page *sagemaker.ListModelPackageGroupsOutput, lastPage bool) bool {
		for _, ModelPackageGroup := range page.ModelPackageGroupSummaryList {
			name := aws.StringValue(ModelPackageGroup.ModelPackageGroupName)

			input := &sagemaker.DeleteModelPackageGroupInput{
				ModelPackageGroupName: ModelPackageGroup.ModelPackageGroupName,
			}

			log.Printf("[INFO] Deleting SageMaker Model Package Group: %s", name)
			if _, err := conn.DeleteModelPackageGroup(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Model Package Group (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Model Package Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Model Package Groups: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerModelPackageGroup_basic(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("model-package-group/%s", rName)),
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

func TestAccAWSSagemakerModelPackageGroup_description(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_description", rName),
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

func TestAccAWSSagemakerModelPackageGroup_tags(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
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
				Config: testAccAWSSagemakerModelPackageGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerModelPackageGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerModelPackageGroup_disappears(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupExists(resourceName, &mpg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerModelPackageGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerModelPackageGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model_package_group" {
			continue
		}

		ModelPackageGroup, err := finder.ModelPackageGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if aws.StringValue(ModelPackageGroup.ModelPackageGroupName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Model Package Group %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerModelPackageGroupExists(n string, mpg *sagemaker.DescribeModelPackageGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Model Package Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.ModelPackageGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccAWSSagemakerModelPackageGroupBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q
}
`, rName)
}

func testAccAWSSagemakerModelPackageGroupDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name        = %[1]q
  model_package_group_description = %[1]q
}
`, rName)
}

func testAccAWSSagemakerModelPackageGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerModelPackageGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
