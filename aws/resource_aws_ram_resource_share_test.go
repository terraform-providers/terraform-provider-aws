package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsRamResourceShare_basic(t *testing.T) {
	var resourceShare ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ram", regexp.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	resourceName := "aws_ram_resource_share.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigAllowExternalPrincipals(rName, false),
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
				Config: testAccAwsRamResourceShareConfigAllowExternalPrincipals(rName, true),
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
	resourceName := "aws_ram_resource_share.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRamResourceShareConfigName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAwsRamResourceShare_Tags(t *testing.T) {
	var resourceShare1, resourceShare2, resourceShare3 ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ram.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsRamResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfigTags1(rName, "key1", "value1"),
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
				Config: testAccAwsRamResourceShareConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsRamResourceShareConfigTags1(rName, "key2", "value2"),
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

func testAccAwsRamResourceShareConfigAllowExternalPrincipals(rName string, allowExternalPrincipals bool) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = %[1]t
  name                      = %[2]q
}
`, allowExternalPrincipals, rName)
}

func testAccAwsRamResourceShareConfigName(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAwsRamResourceShareConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsRamResourceShareConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
