package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSIdentityStoreUserDataSource_UserName(t *testing.T) {
	dataSourceName := "data.aws_identitystore_user.test"
	name := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreUserName(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreUserDataSourceConfigDisplayName(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, "user_name", name),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreUserDataSource_UserID(t *testing.T) {
	dataSourceName := "data.aws_identitystore_user.test"
	name := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")
	userID := os.Getenv("AWS_IDENTITY_STORE_USER_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreUserName(t)
			testAccPreCheckAWSIdentityStoreUserID(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIdentityStoreUserDataSourceConfigUserID(name, userID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user_id", userID),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
				),
			},
		},
	})
}

func TestAccAWSIdentityStoreUserDataSource_NonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIdentityStoreUserDataSourceConfigNonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store User found matching criteria`),
			},
		},
	})
}

func testAccPreCheckAWSIdentityStoreUserName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_NAME env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccPreCheckAWSIdentityStoreUserID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_ID env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccAWSIdentityStoreUserDataSourceConfigDisplayName(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

func testAccAWSIdentityStoreUserDataSourceConfigUserID(name, id string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }

  user_id = %q

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name, id)
}

const testAccAWSIdentityStoreUserDataSourceConfigNonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = "does-not-exist"
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`
