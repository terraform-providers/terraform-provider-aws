package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSPlacementGroup_basic(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPlacementGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "strategy", "cluster"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("placement-group/%s", rName)),
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

func TestAccAWSPlacementGroup_tags(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPlacementGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPlacementGroupExists(resourceName, &pg),
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
				Config: testAccAWSPlacementGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSPlacementGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPlacementGroupExists(resourceName, &pg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccAWSPlacementGroup_disappears(t *testing.T) {
	var pg ec2.PlacementGroup
	resourceName := "aws_placement_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPlacementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPlacementGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPlacementGroupExists(resourceName, &pg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsPlacementGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSPlacementGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_placement_group" {
			continue
		}

		_, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
			GroupNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if isAWSErr(err, "InvalidPlacementGroup.Unknown", "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("still exists")
	}
	return nil
}

func testAccCheckAWSPlacementGroupExists(n string, pg *ec2.PlacementGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Placement Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribePlacementGroups(&ec2.DescribePlacementGroupsInput{
			GroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Placement Group error: %v", err)
		}

		*pg = *resp.PlacementGroups[0]

		return nil
	}
}

func testAccAWSPlacementGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %q
  strategy = "cluster"
}
`, rName)
}

func testAccAWSPlacementGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSPlacementGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
