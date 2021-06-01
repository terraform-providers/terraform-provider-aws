package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSIdentityStoreGroupDataSource_DisplayName(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreGroupDataSourceConfigDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_id"),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", name),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreGroupDataSource_GroupID(t *testing.T) {
	dataSourceName := "data.aws_identitystore_group.test"
	name := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")
	groupID := os.Getenv("AWS_IDENTITY_STORE_GROUP_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
			testAccPreCheckAWSIdentityStoreGroupID(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreGroupDataSourceConfigGroupID(name, groupID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_id", groupID),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreGroupDataSource_NonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIdentityStoreGroupDataSourceConfigNonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store Group found matching criteria`),
			},
		},
	})
}

func testAccPreCheckAWSIdentityStoreGroupName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_NAME env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccPreCheckAWSIdentityStoreGroupID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_GROUP_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_GROUP_ID env var must be set for AWS Identity Store Group acceptance test. " +
			"This is required until ListGroups API returns results without filtering by name.")
	}
}

func testAccAWSIdentityStoreGroupDataSourceConfigDisplayName(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

func testAccAWSIdentityStoreGroupDataSourceConfigGroupID(name, id string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }

  group_id = %q

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name, id)
}

const testAccAWSIdentityStoreGroupDataSourceConfigNonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_group" "test" {
  filter {
    attribute_path  = "DisplayName"
    attribute_value = "does-not-exist"
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`
