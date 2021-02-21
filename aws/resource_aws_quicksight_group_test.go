package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSQuickSightGroup_basic(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("group/default/%s", rName1)),
				),
			},
			{
				Config: testAccAWSQuickSightGroupConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName2),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("group/default/%s", rName2)),
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

func TestAccAWSQuickSightGroup_withDescription(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupConfigWithDescription(rName, "Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 1"),
				),
			},
			{
				Config: testAccAWSQuickSightGroupConfigWithDescription(rName, "Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "description", "Description 2"),
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

func TestAccAWSQuickSightGroup_disappears(t *testing.T) {
	var group quicksight.Group
	resourceName := "aws_quicksight_group.default"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupExists(resourceName, &group),
					testAccCheckQuickSightGroupDisappears(&group),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickSightGroupExists(resourceName string, group *quicksight.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, groupName, err := resourceAwsQuickSightGroupParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		input := &quicksight.DescribeGroupInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		}

		output, err := conn.DescribeGroup(input)

		if err != nil {
			return err
		}

		if output == nil || output.Group == nil {
			return fmt.Errorf("QuickSight Group (%s) not found", rs.Primary.ID)
		}

		*group = *output.Group

		return nil
	}
}

func testAccCheckQuickSightGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).quicksightconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_group" {
			continue
		}

		awsAccountID, namespace, groupName, err := resourceAwsQuickSightGroupParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeGroup(&quicksight.DescribeGroupInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			GroupName:    aws.String(groupName),
		})
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("QuickSight Group '%s' was not deleted properly", rs.Primary.ID)
	}

	return nil
}

func testAccCheckQuickSightGroupDisappears(v *quicksight.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		arn, err := arn.Parse(aws.StringValue(v.Arn))
		if err != nil {
			return err
		}

		parts := strings.SplitN(arn.Resource, "/", 3)

		input := &quicksight.DeleteGroupInput{
			AwsAccountId: aws.String(arn.AccountID),
			Namespace:    aws.String(parts[1]),
			GroupName:    v.GroupName,
		}

		if _, err := conn.DeleteGroup(input); err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSQuickSightGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_group" "default" {
  group_name = %[1]q
}
`, rName)
}

func testAccAWSQuickSightGroupConfigWithDescription(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "default" {
  aws_account_id = data.aws_caller_identity.current.account_id
  group_name     = %[1]q
  description    = %[2]q
}
`, rName, description)
}
