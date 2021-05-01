package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_TagFilter(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigTagFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_IncludeComplianceDetails(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigIncludeComplianceDetails(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "resource_tag_mapping_list.0.compliance_details.0.compliance_status", "true"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_ResourceTypeFilters(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigResourceTypeFilters(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsResourceGroupsTaggingAPIResources_ResourceArnList(t *testing.T) {
	dataSourceName := "data.aws_resourcegroupstaggingapi_resources.test"
	resourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, resourcegroupstaggingapi.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigResourceARNList(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "resource_tag_mapping_list.*", map[string]string{
						"tags.Key": rName,
					}),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "resource_tag_mapping_list.*.resource_arn", resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigTagFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  tag_filter {
    key    = "Key"
    values = [aws_vpc.test.tags["Key"]]
  }
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigResourceTypeFilters(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_type_filters = ["ec2:vpc"]

  depends_on = [aws_vpc.test]
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigResourceARNList(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  resource_arn_list = [aws_vpc.test.arn]
}
`, rName)
}

func testAccDataSourceAwsResourceGroupsTaggingAPIResourcesConfigIncludeComplianceDetails(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Key = %[1]q
  }
}

data "aws_resourcegroupstaggingapi_resources" "test" {
  include_compliance_details  = true
  exclude_compliant_resources = false
  resource_arn_list           = [aws_vpc.test.arn]
}
`, rName)
}
