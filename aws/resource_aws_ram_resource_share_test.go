package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
)

func TestAccAwsRamResourceShare_basic(t *testing.T) {
	var resourceShare ram.ResourceShare
	resourceName := "aws_ram_resource_share.example"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigName(shareName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", shareName),
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

func TestAccAwsRamResourceShare_AllowExternalPrincipals(t *testing.T) {
	var resourceShare1, resourceShare2 ram.ResourceShare
	resourceName := "aws_ram_resource_share.example"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigAllowExternalPrincipals(shareName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRamResourceShareConfigAllowExternalPrincipals(shareName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "true"),
				),
			},
		},
	})
}

func TestAccAwsRamResourceShare_Name(t *testing.T) {
	var resourceShare1, resourceShare2 ram.ResourceShare
	resourceName := "aws_ram_resource_share.example"
	shareName1 := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	shareName2 := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigName(shareName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "name", shareName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRamResourceShareConfigName(shareName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "name", shareName2),
				),
			},
		},
	})
}

func TestAccAwsRamResourceShare_Tags(t *testing.T) {
	var resourceShare1, resourceShare2, resourceShare3 ram.ResourceShare
	resourceName := "aws_ram_resource_share.example"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigTags1(shareName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare1),
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
				Config: testAccAwsRamResourceShareConfigTags2(shareName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRamResourceShareConfigTags1(shareName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsRamResourceShareExists(resourceName string, resourceShare *ram.ResourceShare) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShares(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShares) == 0 {
			return fmt.Errorf("No RAM resource share found")
		}

		resourceShare = output.ResourceShares[0]

		if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusActive {
			return fmt.Errorf("RAM resource share (%s) delet(ing|ed)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsRamResourceShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share" {
			continue
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShares(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShares) > 0 {
			resourceShare := output.ResourceShares[0]
			if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusDeleted {
				return fmt.Errorf("RAM resource share (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAwsRamResourceShareConfigAllowExternalPrincipals(shareName string, allowExternalPrincipals bool) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  allow_external_principals = %t
  name                      = %q
}
`, allowExternalPrincipals, shareName)
}

func testAccAwsRamResourceShareConfigName(shareName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  name = %q
}
`, shareName)
}

func testAccAwsRamResourceShareConfigTags1(shareName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  name = %q

  tags = {
    %q = %q
  }
}
`, shareName, tagKey1, tagValue1)
}

func testAccAwsRamResourceShareConfigTags2(shareName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  name = %q

  tags = {
    %q = %q
    %q = %q
  }
}
`, shareName, tagKey1, tagValue1, tagKey2, tagValue2)
}
