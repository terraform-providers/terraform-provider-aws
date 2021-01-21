package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWorkspaceBundle_basic(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsWorkspaceBundleConfig("wsb-b0s22j3d7"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-b0s22j3d7"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "PERFORMANCE"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Performance with Windows 7"),
					resource.TestCheckResourceAttr(dataSourceName, "owner", "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", "100"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsWorkspaceBundle_byOwnerName(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsWorkspaceBundleConfig_byOwnerName("AMAZON", "Value with Windows 10 and Office 2016"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "bundle_id", "wsb-df76rqys9"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "compute_type.0.name", "VALUE"),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestCheckResourceAttr(dataSourceName, "name", "Value with Windows 10 and Office 2016"),
					resource.TestCheckResourceAttr(dataSourceName, "owner", "Amazon"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "root_storage.0.capacity", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "user_storage.0.capacity", "10"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsWorkspaceBundle_bundleIDAndNameConflict(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWorkspaceBundleConfig_bundleIDAndOwnerNameConflict("wsb-df76rqys9", "AMAZON", "Value with Windows 10 and Office 2016"),
				ExpectError: regexp.MustCompile("\"bundle_id\": conflicts with owner"),
			},
		},
	})
}

func TestAccDataSourceAwsWorkspaceBundle_privateOwner(t *testing.T) {
	dataSourceName := "data.aws_workspaces_bundle.test"
	bundleName := os.Getenv("AWS_WORKSPACES_BUNDLE_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccWorkspacesBundlePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsWorkspaceBundleConfig_privateOwner(bundleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", bundleName),
				),
			},
		},
	})
}

func testAccWorkspacesBundlePreCheck(t *testing.T) {
	if os.Getenv("AWS_WORKSPACES_BUNDLE_NAME") == "" {
		t.Skip("AWS_WORKSPACES_BUNDLE_NAME env var must be set for AWS WorkSpaces private bundle acceptance test. This is required until AWS provides bundle creation API.")
	}
}

func testAccDataSourceAwsWorkspaceBundleConfig(bundleID string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
}
`, bundleID)
}

func testAccDataSourceAwsWorkspaceBundleConfig_byOwnerName(owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  owner = %q
  name  = %q
}
`, owner, name)
}

func testAccDataSourceAwsWorkspaceBundleConfig_bundleIDAndOwnerNameConflict(bundleID, owner, name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
  owner     = %q
  name      = %q
}
`, bundleID, owner, name)
}

func testAccDataSourceAwsWorkspaceBundleConfig_privateOwner(name string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  name = %q
}
`, name)
}
