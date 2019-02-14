package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
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
					resource.TestCheckResourceAttr(dataSourceName, "description", "Windows 7 Experience provided by Windows Server 2008 R2, 2 vCPU, 7.5GiB Memory, 100GB Storage"),
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

func testAccDataSourceAwsWorkspaceBundleConfig(bundleID string) string {
	return fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = %q
}
`, bundleID)
}
