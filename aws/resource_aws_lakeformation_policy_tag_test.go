package aws

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSLakeFormationPolicyTag_basic(t *testing.T) {
	resourceName := "aws_lakeformation_policy_tag.test"
	rKey := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPolicyTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPolicyTagConfig_basic(rKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPolicyTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rKey),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value"),
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
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

func testAccAWSLakeFormationPolicyTag_disappears(t *testing.T) {
	resourceName := "aws_lakeformation_policy_tag.test"
	rKey := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPolicyTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPolicyTagConfig_basic(rKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPolicyTagExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLakeFormationPolicyTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSLakeFormationPolicyTag_values(t *testing.T) {
	resourceName := "aws_lakeformation_policy_tag.test"
	rKey := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPolicyTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccAWSLakeFormationPolicyTagConfig_values(rKey, []string{"value1", "value2"}),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPolicyTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rKey),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test an update that adds, removes and retains a tag value
				Config: testAccAWSLakeFormationPolicyTagConfig_values(rKey, []string{"value1", "value3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPolicyTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rKey),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					resource.TestCheckResourceAttr(resourceName, "values.1", "value3"),
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
		},
	})
}

func testAccCheckAWSLakeFormationPolicyTagsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_policy_tag" {
			continue
		}

		catalogID, tagKey, err := readPolicyTagID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		if _, err := conn.GetLFTag(input); err != nil {
			if isAWSErr(err, lakeformation.ErrCodeEntityNotFoundException, "") {
				continue
			}
			// If the lake formation admin has been revoked, there will be access denied instead of entity not found
			if isAWSErr(err, lakeformation.ErrCodeAccessDeniedException, "") {
				continue
			}
			return err
		}
		return fmt.Errorf("Lake Formation Policy Tag (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSLakeFormationPolicyTagExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogID, tagKey, err := readPolicyTagID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn
		_, err = conn.GetLFTag(input)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSLakeFormationPolicyTagConfig_basic(rKey string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_policy_tag" "test" {
  key    = %[1]q
  values = ["value"]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rKey)
}

func testAccAWSLakeFormationPolicyTagConfig_values(rKey string, values []string) string {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = strconv.Quote(v)
	}

	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_caller_identity.current.arn]
}

resource "aws_lakeformation_policy_tag" "test" {
  key    = %[1]q
  values = [%s]

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rKey, strings.Join(quotedValues, ","))
}
