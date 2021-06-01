package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func TestAccAwsEc2ManagedPrefixList_basic(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func TestAccAwsEc2ManagedPrefixList_disappears(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ManagedPrefixList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_AddressFamily_IPv6(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_AddressFamily(rName, "IPv6"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv6"),
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

func TestAccAwsEc2ManagedPrefixList_Entry(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry1(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "Test2",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry2(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "3.0.0.0/8",
						"description": "Test3",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAwsEc2ManagedPrefixList_Name(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_Tags(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ManagedPrefixListDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list" {
			continue
		}

		pl, err := finder.ManagedPrefixListByID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidPrefixListIDNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading EC2 Managed Prefix List (%s): %w", rs.Primary.ID, err)
		}

		if pl != nil {
			return fmt.Errorf("EC2 Managed Prefix List (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsEc2ManagedPrefixListExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		pl, err := finder.ManagedPrefixListByID(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading EC2 Managed Prefix List (%s): %w", rs.Primary.ID, err)
		}

		if pl == nil {
			return fmt.Errorf("EC2 Managed Prefix List (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsEc2ManagedPrefixListConfig_AddressFamily(rName string, addressFamily string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = %[2]q
  max_entries    = 1
  name           = %[1]q
}
`, rName, addressFamily)
}

func testAccAwsEc2ManagedPrefixListConfig_Entry1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "2.0.0.0/8"
    description = "Test2"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Entry2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "3.0.0.0/8"
    description = "Test3"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsEc2ManagedPrefixListConfig_Tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
